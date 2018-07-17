package core

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"regexp"
	"time"

	"github.com/hunterlong/statup/types"
	"github.com/hunterlong/statup/utils"
)

type FailureData types.FailureData

func CheckServices() {
	CoreApp.Services, _ = SelectAllServices()
	utils.Log(1, fmt.Sprintf("Starting monitoring process for %v Services", len(CoreApp.Services)))
	for _, v := range CoreApp.Services {
		obj := v
		//go obj.StartCheckins()
		obj.stopRoutine = make(chan struct{})
		go obj.CheckQueue()
	}
}

func (s *Service) CheckQueue() {
	for {
		select {
		case <-s.stopRoutine:
			return
		default:
			s.Check()
			if s.Interval < 1 {
				s.Interval = 1
			}
			msg := fmt.Sprintf("Service: %v | Online: %v | Latency: %0.0fms", s.Name, s.Online, (s.Latency * 1000))
			utils.Log(1, msg)
			time.Sleep(time.Duration(s.Interval) * time.Second)
		}
	}
}

func (s *Service) DNSCheck() (float64, error) {
	t1 := time.Now()
	url, err := url.Parse(s.Domain)
	if err != nil {
		return 0, err
	}
	_, err = net.LookupIP(url.Host)
	if err != nil {
		return 0, err
	}
	t2 := time.Now()
	subTime := t2.Sub(t1).Seconds()
	return subTime, err
}

func (s *Service) Check() *Service {
	switch s.Type {
	case "http":
		return s.CheckHTTP()
	case "tcp":
		return s.CheckTCP()
	}

	s.Failure(fmt.Sprintf("Unknown service type %s", s.Type))
	return s
}

func (s *Service) CheckHTTP() *Service {
	dnsLookup, err := s.DNSCheck()
	if err != nil {
		s.Failure(fmt.Sprintf("Could not get IP address for domain %v, %v", s.Domain, err))
		return s
	}
	s.dnsLookup = dnsLookup
	t1 := time.Now()
	client := http.Client{
		Timeout: 30 * time.Second,
	}

	domain := s.Domain
	if s.Port != 0 {
		domain += fmt.Sprintf(":%d", s.Port)
	}

	var response *http.Response
	if s.Method == "POST" {
		response, err = client.Post(domain, "application/json", bytes.NewBuffer([]byte(s.PostData)))
	} else {
		response, err = client.Get(domain)
	}
	if err != nil {
		s.Failure(fmt.Sprintf("HTTP Error %v", err))
		return s
	}
	response.Header.Set("User-Agent", "StatupMonitor")
	t2 := time.Now()
	s.Latency = t2.Sub(t1).Seconds()

	defer response.Body.Close()
	contents, err := ioutil.ReadAll(response.Body)
	if err != nil {
		utils.Log(2, err)
	}
	if s.Expected != "" {
		match, err := regexp.MatchString(s.Expected, string(contents))
		if err != nil {
			utils.Log(2, err)
		}
		if !match {
			s.LastResponse = string(contents)
			s.LastStatusCode = response.StatusCode
			s.Failure(fmt.Sprintf("HTTP Response Body did not match '%v'", s.Expected))
			return s
		}
	}
	if s.ExpectedStatus != response.StatusCode {
		s.LastResponse = string(contents)
		s.LastStatusCode = response.StatusCode
		s.Failure(fmt.Sprintf("HTTP Status Code %v did not match %v", response.StatusCode, s.ExpectedStatus))
		return s
	}
	s.LastResponse = string(contents)
	s.LastStatusCode = response.StatusCode
	s.Online = true
	s.Record()
	return s
}

func (s *Service) CheckTCP() *Service {
	t1 := time.Now()

	conn, err := net.Dial("tcp", s.Domain)
	if err != nil {
		s.Failure(fmt.Sprintf("TCP Dial Error %v", err))
		return s
	}
	if err := conn.Close(); err != nil {
		s.Failure(fmt.Sprintf("TCP Socket Close Error %v", err))
		return s
	}
	t2 := time.Now()
	s.Latency = t2.Sub(t1).Seconds()
	s.LastResponse = ""
	s.LastStatusCode = 200
	s.Online = true
	s.Record()
	return s
}

type HitData struct {
	Latency float64
}

func (s *Service) Record() {
	s.Online = true
	s.LastOnline = time.Now()
	data := HitData{
		Latency: s.Latency,
	}
	s.CreateHit(data)
	OnSuccess(s)
}

func (s *Service) Failure(issue string) {
	s.Online = false
	data := FailureData{
		Issue: issue,
	}
	utils.Log(2, fmt.Sprintf("Service %v Failing: %v", s.Name, issue))
	s.CreateFailure(data)
	//SendFailureEmail(s)
	OnFailure(s, data)
}

// Statup
// Copyright (C) 2018.  Hunter Long and the project contributors
// Written by Hunter Long <info@socialeck.com> and the project contributors
//
// https://github.com/hunterlong/statup
//
// The licenses for most software and other practical works are designed
// to take away your freedom to share and change the works.  By contrast,
// the GNU General Public License is intended to guarantee your freedom to
// share and change all versions of a program--to make sure it remains free
// software for all its users.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

package core

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"github.com/hunterlong/statup/core/notifier"
	"github.com/hunterlong/statup/types"
	"github.com/hunterlong/statup/utils"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"
)

// checkServices will start the checking go routine for each service
func checkServices() {
	utils.Log(1, fmt.Sprintf("Starting monitoring process for %v Services", len(CoreApp.Services)))
	for _, ser := range CoreApp.Services {
		//go obj.StartCheckins()
		go ser.CheckQueue(true)
	}
}

// CheckQueue is the main go routine for checking a service
func (s *Service) CheckQueue(record bool) {
	s.Checkpoint = time.Now()
	s.SleepDuration = time.Duration((time.Duration(s.Id) * 100) * time.Millisecond)
CheckLoop:
	for {
		select {
		case <-s.Running:
			utils.Log(1, fmt.Sprintf("Stopping service: %v", s.Name))
			break CheckLoop
		case <-time.After(s.SleepDuration):
			s.Check(record)
			s.Checkpoint = s.Checkpoint.Add(s.duration())
			sleep := s.Checkpoint.Sub(time.Now())
			if !s.Online {
				s.SleepDuration = s.duration()
			} else {
				s.SleepDuration = sleep
			}
		}
		continue
	}
}

// duration returns the amount of duration for a service to check its status
func (s *Service) duration() time.Duration {
	var amount time.Duration
	if s.Interval >= 10000 {
		amount = time.Duration(s.Interval) * time.Microsecond
	} else {
		amount = time.Duration(s.Interval) * time.Second
	}
	return amount
}

func (s *Service) parseHost() string {
	if s.Type == "tcp" || s.Type == "udp" {
		return s.Domain
	} else {
		domain := s.Domain
		hasPort, _ := regexp.MatchString(`\:([0-9]+)`, domain)
		if hasPort {
			splitDomain := strings.Split(s.Domain, ":")
			domain = splitDomain[len(splitDomain)-2]
		}
		host, err := url.Parse(domain)
		if err != nil {
			return s.Domain
		}
		return host.Host
	}
}

// dnsCheck will check the domain name and return a float64 for the amount of time the DNS check took
func (s *Service) dnsCheck() (float64, error) {
	var err error
	t1 := time.Now()
	host := s.parseHost()
	if s.Type == "tcp" {
		_, err = net.LookupHost(host)
	} else {
		_, err = net.LookupIP(host)
	}
	if err != nil {
		return 0, err
	}
	t2 := time.Now()
	subTime := t2.Sub(t1).Seconds()
	return subTime, err
}

// checkTcp will check a TCP service
func (s *Service) checkTcp(record bool) *Service {
	dnsLookup, err := s.dnsCheck()
	if err != nil {
		if record {
			recordFailure(s, fmt.Sprintf("Could not get IP address for TCP service %v, %v", s.Domain, err))
		}
		return s
	}
	s.PingTime = dnsLookup
	t1 := time.Now()
	domain := fmt.Sprintf("%v", s.Domain)
	if s.Port != 0 {
		domain = fmt.Sprintf("%v:%v", s.Domain, s.Port)
	}
	conn, err := net.DialTimeout(s.Type, domain, time.Duration(s.Timeout)*time.Second)
	if err != nil {
		if record {
			recordFailure(s, fmt.Sprintf("%v Dial Error %v", s.Type, err))
		}
		return s
	}
	if err := conn.Close(); err != nil {
		if record {
			recordFailure(s, fmt.Sprintf("TCP Socket Close Error %v", err))
		}
		return s
	}
	t2 := time.Now()
	s.Latency = t2.Sub(t1).Seconds()
	s.LastResponse = ""
	if record {
		recordSuccess(s)
	}
	return s
}

// checkHttp will check a HTTP service
func (s *Service) checkHttp(record bool) *Service {
	dnsLookup, err := s.dnsCheck()
	if err != nil {
		if record {
			recordFailure(s, fmt.Sprintf("Could not get IP address for domain %v, %v", s.Domain, err))
		}
		return s
	}
	s.PingTime = dnsLookup
	t1 := time.Now()
	timeout := time.Duration(time.Duration(s.Timeout) * time.Second)
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
		TLSHandshakeTimeout: timeout,
	}
	client := &http.Client{
		Transport: transport,
		Timeout:   timeout,
	}
	var response *http.Response
	if s.Method == "POST" {
		response, err = client.Post(s.Domain, "application/json", bytes.NewBuffer([]byte(s.PostData.String)))
	} else {
		response, err = client.Get(s.Domain)
	}
	if err != nil {
		if record {
			recordFailure(s, fmt.Sprintf("HTTP Error %v", err))
		}
		return s
	}
	response.Header.Set("Connection", "close")
	response.Header.Set("User-Agent", "StatupMonitor")
	t2 := time.Now()
	s.Latency = t2.Sub(t1).Seconds()
	if err != nil {
		if record {
			recordFailure(s, fmt.Sprintf("HTTP Error %v", err))
		}
		return s
	}
	defer response.Body.Close()
	contents, err := ioutil.ReadAll(response.Body)
	s.LastResponse = string(contents)
	s.LastStatusCode = response.StatusCode

	if s.Expected.String != "" {
		if err != nil {
			utils.Log(2, err)
		}
		match, err := regexp.MatchString(s.Expected.String, string(contents))
		if err != nil {
			utils.Log(2, err)
		}
		if !match {
			if record {
				recordFailure(s, fmt.Sprintf("HTTP Response Body did not match '%v'", s.Expected))
			}
			return s
		}
	}
	if s.ExpectedStatus != response.StatusCode {
		if record {
			recordFailure(s, fmt.Sprintf("HTTP Status Code %v did not match %v", response.StatusCode, s.ExpectedStatus))
		}
		return s
	}
	s.Online = true
	if record {
		recordSuccess(s)
	}
	return s
}

// Check will run checkHttp for HTTP services and checkTcp for TCP services
func (s *Service) Check(record bool) {
	switch s.Type {
	case "http":
		s.checkHttp(record)
	case "tcp", "udp":
		s.checkTcp(record)
	}
}

// recordSuccess will create a new 'hit' record in the database for a successful/online service
func recordSuccess(s *Service) {
	s.Online = true
	s.LastOnline = time.Now()
	hit := &types.Hit{
		Service:   s.Id,
		Latency:   s.Latency,
		PingTime:  s.PingTime,
		CreatedAt: time.Now(),
	}
	utils.Log(1, fmt.Sprintf("Service %v Successful Response: %0.2f ms | Lookup in: %0.2f ms", s.Name, hit.Latency*1000, hit.PingTime*1000))
	s.CreateHit(hit)
	notifier.OnSuccess(s.Service)
}

// recordFailure will create a new 'failure' record in the database for a offline service
func recordFailure(s *Service, issue string) {
	s.Online = false
	fail := &failure{&types.Failure{
		Service:   s.Id,
		Issue:     issue,
		PingTime:  s.PingTime,
		CreatedAt: time.Now(),
	}}
	utils.Log(2, fmt.Sprintf("Service %v Failing: %v | Lookup in: %0.2f ms", s.Name, issue, fail.PingTime*1000))
	s.CreateFailure(fail)
	notifier.OnFailure(s.Service, fail.Failure)
}

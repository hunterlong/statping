package services

import (
	"bytes"
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/smtp"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/statping/statping/types/metrics"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	"github.com/emersion/go-imap/client"
	"github.com/statping/statping/types/failures"
	"github.com/statping/statping/types/hits"
	"github.com/statping/statping/utils"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
)

// checkServices will start the checking go routine for each service
func CheckServices() {
	log.Infoln(fmt.Sprintf("Starting monitoring process for %v Services", len(allServices)))
	for _, s := range allServices {
		time.Sleep(50 * time.Millisecond)
		go ServiceCheckQueue(s, true)
	}
}

// CheckQueue is the main go routine for checking a service
func ServiceCheckQueue(s *Service, record bool) {
	s.Start()
	s.Checkpoint = utils.Now()
	s.SleepDuration = (time.Duration(s.Id) * 100) * time.Millisecond

CheckLoop:
	for {
		select {
		case <-s.Running:
			log.Infoln(fmt.Sprintf("Stopping service: %v", s.Name))
			break CheckLoop
		case <-time.After(s.SleepDuration):
			s.CheckService(record)
			s.UpdateStats()
			s.Checkpoint = s.Checkpoint.Add(s.Duration())
			if !s.Online {
				s.SleepDuration = s.Duration()
			} else {
				s.SleepDuration = s.Checkpoint.Sub(time.Now())
			}
		}
	}
}

func parseHost(s *Service) string {
	if s.Type == "tcp" || s.Type == "udp" || s.Type == "grpc" || s.Type == "smtp" || s.Type == "imap" {
		return s.Domain
	} else {
		u, err := url.Parse(s.Domain)
		if err != nil {
			return s.Domain
		}
		return strings.Split(u.Host, ":")[0]
	}
}

// dnsCheck will check the domain name and return a float64 for the amount of time the DNS check took
func dnsCheck(s *Service) (int64, error) {
	var err error
	t1 := utils.Now()
	host := parseHost(s)
	if s.Type == "tcp" || s.Type == "udp" || s.Type == "grpc" || s.Type == "smtp" {
		_, err = net.LookupHost(host)
	} else {
		_, err = net.LookupIP(host)
	}
	if err != nil {
		return 0, err
	}
	return utils.Now().Sub(t1).Microseconds(), err
}

func isIPv6(address string) bool {
	return strings.Count(address, ":") >= 2
}

// checkIcmp will send a ICMP ping packet to the service
func CheckIcmp(s *Service, record bool) (*Service, error) {
	defer s.updateLastCheck()
	timer := prometheus.NewTimer(metrics.ServiceTimer(s.Name))
	defer timer.ObserveDuration()

	dur, err := utils.Ping(s.Domain, s.Timeout)
	if err != nil {
		if record {
			RecordFailure(s, fmt.Sprintf("Could not send ICMP to service %v, %v", s.Domain, err), "lookup")
		}
		return s, err
	}

	s.PingTime = dur
	s.Latency = dur
	s.LastResponse = ""
	s.Online = true
	if record {
		RecordSuccess(s)
	}
	return s, nil
}

// CheckGrpc will check a gRPC service
func CheckGrpc(s *Service, record bool) (*Service, error) {
	defer s.updateLastCheck()
	timer := prometheus.NewTimer(metrics.ServiceTimer(s.Name))
	defer timer.ObserveDuration()

	// Strip URL scheme if present. Eg: https:// , http://
	if strings.Contains(s.Domain, "://") {
		u, err := url.Parse(s.Domain)
		if err != nil {
			// Unable to parse.
			log.Warnln(fmt.Sprintf("GRPC Service: '%s', Unable to parse URL: '%v'", s.Name, s.Domain))
			if record {
				RecordFailure(s, fmt.Sprintf("Unable to parse GRPC domain %v, %v", s.Domain, err), "parse_domain")
			}
		}

		// Set domain as hostname without port number.
		s.Domain = u.Hostname()
	}

	// Calculate DNS check time
	dnsLookup, err := dnsCheck(s)
	if err != nil {
		if record {
			RecordFailure(s, fmt.Sprintf("Could not get IP address for GRPC service %v, %v", s.Domain, err), "lookup")
		}
		return s, err
	}

	// Connect to grpc service without TLS certs.
	grpcOption := grpc.WithInsecure()

	// Check if TLS is enabled
	// Upgrade GRPC connection if using TLS
	// Force to connect on HTTP2 with TLS. Needed when using a reverse proxy such as nginx.
	if s.VerifySSL.Bool {
		h2creds := credentials.NewTLS(&tls.Config{NextProtos: []string{"h2"}})
		grpcOption = grpc.WithTransportCredentials(h2creds)
	}

	s.PingTime = dnsLookup
	t1 := utils.Now()
	domain := fmt.Sprintf("%v", s.Domain)
	if s.Port != 0 {
		domain = fmt.Sprintf("%v:%v", s.Domain, s.Port)
		if isIPv6(s.Domain) {
			domain = fmt.Sprintf("[%v]:%v", s.Domain, s.Port)
		}
	}

	// Context will cancel the request when timeout is exceeded.
	// Cancel the context when request is served within the timeout limit.
	timeout := time.Duration(s.Timeout) * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	conn, err := grpc.DialContext(ctx, domain, grpcOption, grpc.WithBlock())
	if err != nil {
		if record {
			RecordFailure(s, fmt.Sprintf("Dial Error %v", err), "connection")
		}
		return s, err
	}

	if s.GrpcHealthCheck.Bool {
		// Create a new health check client
		c := healthpb.NewHealthClient(conn)
		in := &healthpb.HealthCheckRequest{}
		res, err := c.Check(ctx, in)
		if err != nil {
			if record {
				RecordFailure(s, fmt.Sprintf("GRPC Error %v", err), "healthcheck")
			}
			return s, nil
		}

		// Record responses
		s.LastResponse = strings.TrimSpace(res.String())
		s.LastStatusCode = int(res.GetStatus())
	}

	if err := conn.Close(); err != nil {
		if record {
			RecordFailure(s, fmt.Sprintf("%v Socket Close Error %v", strings.ToUpper(s.Type), err), "close")
		}
		return s, err
	}

	// Record latency
	s.Latency = utils.Now().Sub(t1).Microseconds()
	s.Online = true

	if s.GrpcHealthCheck.Bool {
		if s.ExpectedStatus != s.LastStatusCode {
			if record {
				RecordFailure(s, fmt.Sprintf("GRPC Service: '%s', Status Code: expected '%v', got '%v'", s.Name, s.ExpectedStatus, s.LastStatusCode), "response_code")
			}
			return s, nil
		}

		if s.Expected.String != s.LastResponse {
			log.Warnln(fmt.Sprintf("GRPC Service: '%s', Response: expected '%v', got '%v'", s.Name, s.Expected.String, s.LastResponse))
			if record {
				RecordFailure(s, fmt.Sprintf("GRPC Response Body '%v' did not match '%v'", s.LastResponse, s.Expected.String), "response_body")
			}
			return s, nil
		}
	}

	if record {
		RecordSuccess(s)
	}

	return s, nil
}

// checkTcp will check a TCP service
func CheckTcp(s *Service, record bool) (*Service, error) {
	defer s.updateLastCheck()
	timer := prometheus.NewTimer(metrics.ServiceTimer(s.Name))
	defer timer.ObserveDuration()

	dnsLookup, err := dnsCheck(s)
	if err != nil {
		if record {
			RecordFailure(s, fmt.Sprintf("Could not get IP address for TCP service %v, %v", s.Domain, err), "lookup")
		}
		return s, err
	}
	s.PingTime = dnsLookup
	t1 := utils.Now()
	domain := fmt.Sprintf("%v", s.Domain)
	if s.Port != 0 {
		domain = fmt.Sprintf("%v:%v", s.Domain, s.Port)
		if isIPv6(s.Domain) {
			domain = fmt.Sprintf("[%v]:%v", s.Domain, s.Port)
		}
	}

	tlsConfig, err := s.LoadTLSCert()
	if err != nil {
		log.Errorln(err)
	}

	// test TCP connection if there is no TLS Certificate set
	if s.TLSCert.String == "" {
		conn, err := net.DialTimeout(s.Type, domain, time.Duration(s.Timeout)*time.Second)
		if err != nil {
			if record {
				RecordFailure(s, fmt.Sprintf("Dial Error: %v", err), "tls")
			}
			return s, err
		}
		defer conn.Close()
	} else {
		// test TCP connection if TLS Certificate was set
		dialer := &net.Dialer{
			KeepAlive: time.Duration(s.Timeout) * time.Second,
			Timeout:   time.Duration(s.Timeout) * time.Second,
		}
		conn, err := tls.DialWithDialer(dialer, s.Type, domain, tlsConfig)
		if err != nil {
			if record {
				RecordFailure(s, fmt.Sprintf("Dial Error: %v", err), "tls")
			}
			return s, err
		}
		defer conn.Close()
	}

	s.Latency = utils.Now().Sub(t1).Microseconds()
	s.LastResponse = ""
	s.Online = true
	if record {
		RecordSuccess(s)
	}
	return s, nil
}

// checkSmtp will check an SMTP service
func CheckSmtp(s *Service, record bool) (*Service, error) {
	defer s.updateLastCheck()
	timer := prometheus.NewTimer(metrics.ServiceTimer(s.Name))
	defer timer.ObserveDuration()

	dnsLookup, err := dnsCheck(s)
	if err != nil {
		if record {
			RecordFailure(s, fmt.Sprintf("Could not get IP address for %s service %v, %v", strings.ToUpper(s.Type), s.Domain, err), "lookup")
		}
		return s, err
	}
	s.PingTime = dnsLookup
	t1 := utils.Now()
	domain := fmt.Sprintf("%v", s.Domain)
	if s.Port != 0 {
		domain = fmt.Sprintf("%v:%v", s.Domain, s.Port)
		if isIPv6(s.Domain) {
			domain = fmt.Sprintf("[%v]:%v", s.Domain, s.Port)
		}
	}

	tlsConfig, err := s.LoadTLSCert()
	if err != nil {
		log.Errorln(err)
	}

	var c *smtp.Client
	var headers []string
	var username, password string
	if s.Headers.Valid {
		headers = strings.Split(s.Headers.String, ",")
	} else {
		headers = nil
	}

	// check if 'Content-Type' header was defined
	for _, header := range headers {
		switch strings.ToLower(strings.Split(header, "=")[0]) {
		case "username":
			username = strings.Split(header, "=")[1]
		case "password":
			password = strings.Split(header, "=")[1]
		}
	}

	if s.requiresTLS() || s.TLSCert.String != "" {
		// test TCP connection if TLS Certificate was set
		dialer := &net.Dialer{
			KeepAlive: time.Duration(s.Timeout) * time.Second,
			Timeout:   time.Duration(s.Timeout) * time.Second,
		}
		conn, err := tls.DialWithDialer(dialer, "tcp", domain, tlsConfig)
		if err != nil {
			if record {
				RecordFailure(s, fmt.Sprintf("Dial Error: %v", err), "tls")
			}
			return s, err
		}
		defer conn.Close()
		c, err = smtp.NewClient(conn, s.Domain)
		if err != nil {
			if record {
				RecordFailure(s, fmt.Sprintf("%s Connection Error: %v", strings.ToUpper(s.Type), err), s.Type)
			}
			return s, err
		}
	} else {
		// test TCP connection if there is no TLS Certificate set
		conn, err := net.DialTimeout("tcp", domain, time.Duration(s.Timeout)*time.Second)
		if err != nil {
			if record {
				RecordFailure(s, fmt.Sprintf("Dial Error: %v", err), "tls")
			}
			return s, err
		}
		defer conn.Close()
		c, err = smtp.NewClient(conn, s.Domain)
		if err != nil {
			if record {
				RecordFailure(s, fmt.Sprintf("%s Connection Error: %v", strings.ToUpper(s.Type), err), s.Type)
			}
			return s, err
		}
	}

	// Auth
	if s.Port != 25 {
		if username == "" || password == "" {
			err = errors.New("no credentials configured")
			if record {
				RecordFailure(s, fmt.Sprintf("%s Authentication Error: %v", strings.ToUpper(s.Type), err), s.Type)
			}
			return s, err
		}

		if err = c.Auth(smtp.PlainAuth("", username, password, s.Domain)); err != nil {
			if record {
				RecordFailure(s, fmt.Sprintf("%s Authentication Error: %v", strings.ToUpper(s.Type), err), s.Type)
			}
			return s, err
		}
	}

	s.Latency = utils.Now().Sub(t1).Microseconds()
	s.LastResponse = ""
	s.Online = true
	if record {
		RecordSuccess(s)
	}
	return s, nil
}

func CheckImap(s *Service, record bool) (*Service, error) {
	defer s.updateLastCheck()
	timer := prometheus.NewTimer(metrics.ServiceTimer(s.Name))
	defer timer.ObserveDuration()

	dnsLookup, err := dnsCheck(s)
	if err != nil {
		if record {
			RecordFailure(s, fmt.Sprintf("Could not get IP address for %s service %v, %v", strings.ToUpper(s.Type), s.Domain, err), "lookup")
		}
		return s, err
	}
	s.PingTime = dnsLookup
	t1 := utils.Now()
	domain := fmt.Sprintf("%v", s.Domain)
	if s.Port != 0 {
		domain = fmt.Sprintf("%v:%v", s.Domain, s.Port)
		if isIPv6(s.Domain) {
			domain = fmt.Sprintf("[%v]:%v", s.Domain, s.Port)
		}
	}

	tlsConfig, err := s.LoadTLSCert()
	if err != nil {
		log.Errorln(err)
	}

	var headers []string
	var username, password string
	if s.Headers.Valid {
		headers = strings.Split(s.Headers.String, ",")
	} else {
		headers = nil
	}

	// check if 'Content-Type' header was defined
	for _, header := range headers {
		switch strings.ToLower(strings.Split(header, "=")[0]) {
		case "username":
			username = strings.Split(header, "=")[1]
		case "password":
			password = strings.Split(header, "=")[1]
		}
	}

	var conn *client.Client
	if s.requiresTLS() || s.TLSCert.String != "" {
		// test TCP connection if TLS Certificate was set
		dialer := &net.Dialer{
			KeepAlive: time.Duration(s.Timeout) * time.Second,
			Timeout:   time.Duration(s.Timeout) * time.Second,
		}
		conn, err = client.DialWithDialerTLS(dialer, domain, tlsConfig)
		if err != nil {
			if record {
				RecordFailure(s, fmt.Sprintf("Dial Error: %v", err), "tls")
			}
			return s, err
		}
	} else {
		// test TCP connection if there is no TLS Certificate set
		dialer := &net.Dialer{
			KeepAlive: time.Duration(s.Timeout) * time.Second,
			Timeout:   time.Duration(s.Timeout) * time.Second,
		}
		conn, err = client.DialWithDialer(dialer, domain)
		if err != nil {
			if record {
				RecordFailure(s, fmt.Sprintf("Dial Error: %v", err), "tls")
			}
			return s, err
		}
	}
	defer conn.Logout()

	// Auth
	if s.Port != 143 {
		if username == "" || password == "" {
			err = errors.New("no credentials configured")
			if record {
				RecordFailure(s, fmt.Sprintf("%s Authentication Error: %v", strings.ToUpper(s.Type), err), s.Type)
			}
			return s, err
		}

		if err = conn.Login(username, password); err != nil {
			if record {
				RecordFailure(s, fmt.Sprintf("%s Authentication Error: %v", strings.ToUpper(s.Type), err), s.Type)
			}
			return s, err
		}
	}

	s.Latency = utils.Now().Sub(t1).Microseconds()
	s.LastResponse = ""
	s.Online = true
	if record {
		RecordSuccess(s)
	}
	return s, nil
}

func (s *Service) updateLastCheck() {
	s.LastCheck = time.Now()
}

// checkHttp will check a HTTP service
func CheckHttp(s *Service, record bool) (*Service, error) {
	defer s.updateLastCheck()
	timer := prometheus.NewTimer(metrics.ServiceTimer(s.Name))
	defer timer.ObserveDuration()

	dnsLookup, err := dnsCheck(s)
	if err != nil {
		if record {
			RecordFailure(s, fmt.Sprintf("Could not get IP address for domain %v, %v", s.Domain, err), "lookup")
		}
		return s, err
	}
	s.PingTime = dnsLookup
	t1 := utils.Now()

	timeout := time.Duration(s.Timeout) * time.Second
	var content []byte
	var res *http.Response
	var data *bytes.Buffer
	var headers []string
	contentType := "application/json" // default Content-Type

	if s.Headers.Valid {
		headers = strings.Split(s.Headers.String, ",")
	} else {
		headers = nil
	}

	// check if 'Content-Type' header was defined
	for _, header := range headers {
		if strings.Split(header, "=")[0] == "Content-Type" {
			contentType = strings.Split(header, "=")[1]
			break
		}
	}

	if s.Redirect.Bool {
		headers = append(headers, "Redirect=true")
	}

	if s.PostData.String != "" {
		data = bytes.NewBuffer([]byte(s.PostData.String))
	} else {
		data = bytes.NewBuffer(nil)
	}

	// force set Content-Type to 'application/json' if requests are made
	// with POST method
	if s.Method == "POST" && contentType != "application/json" {
		contentType = "application/json"
	}

	customTLS, err := s.LoadTLSCert()
	if err != nil {
		log.Errorln(err)
	}

	content, res, err = utils.HttpRequest(s.Domain, s.Method, contentType, headers, data, timeout, s.VerifySSL.Bool, customTLS)
	if err != nil {
		if record {
			RecordFailure(s, fmt.Sprintf("HTTP Error %v", err), "request")
		}
		return s, err
	}
	s.Latency = utils.Now().Sub(t1).Microseconds()
	s.LastResponse = string(content)
	s.LastStatusCode = res.StatusCode

	metrics.Gauge("status_code", float64(res.StatusCode), s.Name)

	if s.Expected.String != "" {
		match, err := regexp.MatchString(s.Expected.String, string(content))
		if err != nil {
			log.Warnln(fmt.Sprintf("Service %v expected: %v to match %v", s.Name, string(content), s.Expected.String))
		}
		if !match {
			if record {
				RecordFailure(s, fmt.Sprintf("HTTP Response Body did not match '%v'", s.Expected), "regex")
			}
			return s, err
		}
	}
	if s.ExpectedStatus != res.StatusCode {
		if record {
			RecordFailure(s, fmt.Sprintf("HTTP Status Code %v did not match %v", res.StatusCode, s.ExpectedStatus), "status_code")
		}
		return s, err
	}
	if record {
		RecordSuccess(s)
	}
	s.Online = true
	return s, err
}

// RecordSuccess will create a new 'hit' record in the database for a successful/online service
func RecordSuccess(s *Service) {
	s.LastOnline = utils.Now()
	s.Online = true
	hit := &hits.Hit{
		Service:   s.Id,
		Latency:   s.Latency,
		PingTime:  s.PingTime,
		CreatedAt: utils.Now(),
	}
	if err := hit.Create(); err != nil {
		log.Error(err)
	}
	log.WithFields(utils.ToFields(hit, s)).Infoln(
		fmt.Sprintf("Service #%d '%v' Successful Response: %s | Lookup in: %s | Online: %v | Interval: %d seconds", s.Id, s.Name, humanMicro(hit.Latency), humanMicro(hit.PingTime), s.Online, s.Interval))
	s.LastLookupTime = hit.PingTime
	s.LastLatency = hit.Latency
	metrics.Gauge("online", 1., s.Name, s.Type)
	metrics.Inc("success", s.Name)
	sendSuccess(s)
}

// RecordFailure will create a new 'Failure' record in the database for a offline service
func RecordFailure(s *Service, issue, reason string) {
	s.LastOffline = utils.Now()

	fail := &failures.Failure{
		Service:   s.Id,
		Issue:     issue,
		PingTime:  s.PingTime,
		CreatedAt: utils.Now(),
		ErrorCode: s.LastStatusCode,
		Reason:    reason,
	}
	log.WithFields(utils.ToFields(fail, s)).
		Warnln(fmt.Sprintf("Service %v Failing: %v | Lookup in: %v", s.Name, issue, humanMicro(fail.PingTime)))

	if err := fail.Create(); err != nil {
		log.Error(err)
	}
	s.Online = false
	s.DownText = s.DowntimeText()

	limitOffset := len(s.Failures)
	if len(s.Failures) >= limitFailures {
		limitOffset = limitFailures - 1
	}

	s.Failures = append([]*failures.Failure{fail}, s.Failures[:limitOffset]...)

	metrics.Gauge("online", 0., s.Name, s.Type)
	metrics.Inc("failure", s.Name)
	sendFailure(s, fail)
}

// Check will run checkHttp for HTTP services and checkTcp for TCP services
// if record param is set to true, it will add a record into the database.
func (s *Service) CheckService(record bool) {
	switch s.Type {
	case "http":
		CheckHttp(s, record)
	case "tcp", "udp":
		CheckTcp(s, record)
	case "grpc":
		CheckGrpc(s, record)
	case "icmp":
		CheckIcmp(s, record)
	case "smtp":
		CheckSmtp(s, record)
	case "imap":
		CheckImap(s, record)
	}
}

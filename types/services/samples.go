package services

import (
	"github.com/statping/statping/types/null"
	"github.com/statping/statping/utils"
	"time"
)

func Example(online bool) Service {
	return Service{
		Id:                  6283,
		Name:                "Statping Example",
		Domain:              "https://statping.com",
		Expected:            null.NewNullString(""),
		ExpectedStatus:      &[]int{200}[0],
		Interval:            int(time.Duration(15 * time.Second).Seconds()),
		Type:                "http",
		Method:              "get",
		PostData:            null.NullString{},
		Port:                443,
		Timeout:             int(time.Duration(2 * time.Second).Seconds()),
		Order:               0,
		VerifySSL:           null.NewNullBool(true),
		Public:              null.NewNullBool(true),
		GroupId:             0,
		TLSCert:             null.NullString{},
		TLSCertKey:          null.NullString{},
		TLSCertRoot:         null.NullString{},
		Headers:             null.NullString{},
		Permalink:           null.NewNullString("example-service"),
		Redirect:            null.NewNullBool(true),
		CreatedAt:           utils.Now().Add(-23 * time.Hour),
		UpdatedAt:           utils.Now().Add(-23 * time.Hour),
		Online:              online,
		Latency:             393443,
		PingTime:            83526,
		Online24Hours:       0.98,
		Online7Days:         0.99,
		AvgResponse:         303443,
		FailuresLast24Hours: 2,
		Checkpoint:          time.Time{},
		SleepDuration:       5 * time.Second,
		LastResponse:        "The example service is hitting this page",
		NotifyAfter:         0,
		notifyAfterCount:    0,
		AllowNotifications:  null.NewNullBool(true),
		UpdateNotify:        null.NewNullBool(true),
		DownText:            "The service was responding with 500 status code",
		LastStatusCode:      200,
		Failures:            nil,
		LastLookupTime:      4600,
		LastLatency:         124399,
		LastCheck:           utils.Now().Add(-37 * time.Second),
		LastOnline:          utils.Now().Add(-37 * time.Second),
		LastOffline:         utils.Now().Add(-75 * time.Second),
		prevOnline:          false,
	}
}

func Samples() error {
	log.Infoln("Inserting Sample Services...")
	createdOn := utils.Now().Add(((-24 * 30) * 3) * time.Hour)
	s1 := &Service{
		Name:           "Google",
		Domain:         "https://google.com",
		ExpectedStatus: &[]int{200}[0],
		Interval:       10,
		Type:           "http",
		Method:         "GET",
		Timeout:        10,
		Order:          1,
		GroupId:        1,
		Public:         null.NewNullBool(true),
		Permalink:      null.NewNullString("google"),
		VerifySSL:      null.NewNullBool(true),
		Redirect:       null.NewNullBool(true),
		NotifyAfter:    3,
		CreatedAt:      createdOn,
	}
	if err := s1.Create(); err != nil {
		return err
	}

	s2 := &Service{
		Name:           "Statping Github",
		Domain:         "https://github.com/statping/statping",
		ExpectedStatus: &[]int{200}[0],
		Interval:       30,
		Type:           "http",
		Method:         "GET",
		Timeout:        20,
		Order:          2,
		Public:         null.NewNullBool(true),
		Permalink:      null.NewNullString("statping_github"),
		VerifySSL:      null.NewNullBool(true),
		NotifyAfter:    1,
		CreatedAt:      createdOn,
	}
	if err := s2.Create(); err != nil {
		return err
	}

	s3 := &Service{
		Name:           "JSON Users Test",
		Domain:         "https://jsonplaceholder.typicode.com/users",
		ExpectedStatus: &[]int{200}[0],
		Interval:       60,
		Type:           "http",
		Method:         "GET",
		Timeout:        30,
		Order:          3,
		Public:         null.NewNullBool(true),
		VerifySSL:      null.NewNullBool(true),
		GroupId:        2,
		NotifyAfter:    2,
		CreatedAt:      createdOn,
	}
	if err := s3.Create(); err != nil {
		return err
	}

	s4 := &Service{
		Name:           "JSON API Tester",
		Domain:         "https://jsonplaceholder.typicode.com/posts",
		ExpectedStatus: &[]int{201}[0],
		Expected:       null.NewNullString(`(title)": "((\\"|[statping])*)"`),
		Interval:       30,
		Type:           "http",
		Method:         "POST",
		PostData:       null.NewNullString(`{ "title": "statping", "body": "bar", "userId": 19999 }`),
		Timeout:        30,
		Order:          4,
		Public:         null.NewNullBool(true),
		VerifySSL:      null.NewNullBool(true),
		Redirect:       null.NewNullBool(true),
		GroupId:        2,
		NotifyAfter:    3,
		CreatedAt:      createdOn,
	}
	if err := s4.Create(); err != nil {
		return err
	}

	s5 := &Service{
		Name:      "Google DNS",
		Domain:    "8.8.8.8",
		Interval:  20,
		Type:      "tcp",
		Port:      53,
		Timeout:   120,
		Order:     5,
		Public:    null.NewNullBool(true),
		GroupId:   1,
		CreatedAt: createdOn,
	}
	if err := s5.Create(); err != nil {
		return err
	}

	s6 := &Service{
		Name:      "Private Service",
		Domain:    "https://push.statping.com",
		Method:    "GET",
		Interval:  30,
		Type:      "http",
		Timeout:   120,
		Order:     6,
		Public:    null.NewNullBool(false),
		Redirect:  null.NewNullBool(true),
		GroupId:   3,
		CreatedAt: createdOn,
	}
	if err := s6.Create(); err != nil {
		return err
	}

	s7 := &Service{
		Name:      "Static Service",
		Domain:    "none",
		Type:      "static",
		Order:     7,
		Public:    null.NewNullBool(true),
		CreatedAt: createdOn,
	}
	if err := s7.Create(); err != nil {
		return err
	}

	return nil
}

package services

import (
	"context"
	"crypto/tls"
	"github.com/gorilla/mux"
	"github.com/statping/statping/database"
	"github.com/statping/statping/types/checkins"
	"github.com/statping/statping/types/failures"
	"github.com/statping/statping/types/hits"
	"github.com/statping/statping/types/incidents"
	"github.com/statping/statping/types/messages"
	"github.com/statping/statping/types/notifications"
	"github.com/statping/statping/types/null"
	"github.com/statping/statping/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	pb "google.golang.org/grpc/examples/route_guide/routeguide"
	"net"
	"net/http"
	"testing"
	"time"
)

var example = &Service{
	Name:           "Example Service",
	Domain:         "https://statping.com",
	ExpectedStatus: &[]int{200}[0],
	Interval:       30,
	Type:           "http",
	Method:         "GET",
	Timeout:        5,
	Order:          1,
	VerifySSL:      null.NewNullBool(true),
	Public:         null.NewNullBool(true),
	GroupId:        1,
	Permalink:      null.NewNullString("statping"),
	LastCheck:      utils.Now().Add(-5 * time.Second),
	LastOffline:    utils.Now().Add(-5 * time.Second),
	LastOnline:     utils.Now().Add(-60 * time.Second),
}

var hit1 = &hits.Hit{
	Service:   1,
	Latency:   123456,
	PingTime:  123456,
	CreatedAt: utils.Now().Add(-120 * time.Second),
}

var hit2 = &hits.Hit{
	Service:   1,
	Latency:   123456,
	PingTime:  123456,
	CreatedAt: utils.Now().Add(-60 * time.Second),
}

var hit3 = &hits.Hit{
	Service:   1,
	Latency:   123456,
	PingTime:  123456,
	CreatedAt: utils.Now().Add(-30 * time.Second),
}

var exmapleCheckin = &checkins.Checkin{
	ServiceId: 1,
	Name:      "Example Checkin",
	Interval:  3,
	ApiKey:    "wdededede",
}

var fail1 = &failures.Failure{
	Issue:     "example not found",
	ErrorCode: 404,
	Service:   1,
	PingTime:  123456,
	CreatedAt: utils.Now().Add(-160 * time.Second),
}

var fail2 = &failures.Failure{
	Issue:     "example 2 not found",
	ErrorCode: 500,
	Service:   1,
	PingTime:  123456,
	CreatedAt: utils.Now().Add(-5 * time.Second),
}

var incident1 = &incidents.Incident{
	Title:       "Theres something going on",
	Description: "this is an example",
	ServiceId:   1,
	CreatedAt:   utils.Now().Add(-30 * time.Second),
}

var incidentUpdate1 = &incidents.IncidentUpdate{
	IncidentId: 1,
	Message:    "This is an update",
	Type:       "pending",
	CreatedAt:  utils.Now().Add(-5 * time.Second),
}

var message1 = &messages.Message{
	Title:             "Example Message",
	Description:       "Used for testing",
	StartOn:           utils.Now().Add(15 * time.Minute),
	EndOn:             utils.Now().Add(30 * time.Minute),
	ServiceId:         1,
	NotifyUsers:       null.NewNullBool(false),
	NotifyMethod:      "",
	NotifyBefore:      null.NewNullInt64(0),
	NotifyBeforeScale: "",
	CreatedAt:         utils.Now(),
	UpdatedAt:         utils.Now(),
}

type exampleGRPC struct {
	pb.UnimplementedRouteGuideServer
}

func (s *exampleGRPC) GetFeature(ctx context.Context, point *pb.Point) (*pb.Feature, error) {
	return &pb.Feature{Location: point}, nil
}

func TestStartExampleEndpoints(t *testing.T) {
	startupDb(t)

	// root CA for Linux:  /etc/ssl/certs/ca-certificates.crt
	// root CA for MacOSX: /opt/local/share/curl/curl-ca-bundle.crt

	tlsCert := utils.Params.GetString("STATPING_DIR") + "/cert.pem"
	tlsCertKey := utils.Params.GetString("STATPING_DIR") + "/key.pem"

	require.FileExists(t, tlsCert)
	require.FileExists(t, tlsCertKey)

	h := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	}

	r := mux.NewRouter()
	r.HandleFunc("/", h)

	// start example HTTP server
	go func(t *testing.T) {
		require.Nil(t, http.ListenAndServe(":15000", r))
	}(t)

	// start example TLS HTTP server
	go func(t *testing.T) {
		require.Nil(t, http.ListenAndServeTLS(":15001", tlsCert, tlsCertKey, r))
	}(t)

	tcpHandle := func(conn net.Conn) {
		defer conn.Close()
		conn.Write([]byte("ok"))
	}

	// start TCP server
	go func(t *testing.T, hdl func(conn net.Conn)) {
		ln, err := net.Listen("tcp", ":15002")
		require.Nil(t, err)
		for {
			conn, err := ln.Accept()
			require.Nil(t, err)
			go hdl(conn)
		}
	}(t, tcpHandle)

	// start TLS TCP server
	go func(t *testing.T, hdl func(conn net.Conn)) {
		cer, err := tls.LoadX509KeyPair(tlsCert, tlsCertKey)
		require.Nil(t, err)
		ln, err := tls.Listen("tcp", ":15003", &tls.Config{Certificates: []tls.Certificate{cer}})
		require.Nil(t, err)
		for {
			conn, err := ln.Accept()
			require.Nil(t, err)
			go hdl(conn)
		}
	}(t, tcpHandle)

	// start GRPC server
	go func(t *testing.T) {
		list, err := net.Listen("tcp", ":15004")
		require.Nil(t, err)
		grpcServer := grpc.NewServer()
		pb.RegisterRouteGuideServer(grpcServer, &exampleGRPC{})
		require.Nil(t, grpcServer.Serve(list))
	}(t)

	time.Sleep(15 * time.Second)
}

func startupDb(t *testing.T) {
	err := utils.InitLogs()
	require.Nil(t, err)
	db, err := database.OpenTester()
	require.Nil(t, err)
	db.AutoMigrate(&Service{}, &notifications.Notification{}, &messages.Message{}, &hits.Hit{}, &checkins.Checkin{}, &checkins.CheckinHit{}, &failures.Failure{}, &incidents.Incident{}, &incidents.IncidentUpdate{})
	checkins.SetDB(db)
	failures.SetDB(db)
	incidents.SetDB(db)
	notifications.SetDB(db)
	messages.SetDB(db)
	hits.SetDB(db)
	SetDB(db)

	db.Create(&example)
	db.Create(&hit1)
	db.Create(&hit2)
	db.Create(&hit3)
	db.Create(&exmapleCheckin)
	db.Create(&fail1)
	db.Create(&fail2)
	db.Create(&incident1)
	db.Create(&incidentUpdate1)
	db.Create(&notification.Notification)
	db.Create(&message1)
}

func TestServices(t *testing.T) {

	tlsCert := utils.Params.GetString("STATPING_DIR") + "/cert.pem"
	tlsCertKey := utils.Params.GetString("STATPING_DIR") + "/key.pem"

	t.Run("Test Find service", func(t *testing.T) {
		item, err := Find(1)
		require.Nil(t, err)
		assert.Equal(t, "Example Service", item.Name)
		assert.NotZero(t, item.LastOnline)
		assert.NotZero(t, item.LastOffline)
		assert.NotZero(t, item.LastCheck)
	})

	t.Run("Test HTTP Check", func(t *testing.T) {
		e := &Service{
			Name:           "Example HTTP",
			Domain:         "http://localhost:15000",
			ExpectedStatus: &[]int{200}[0],
			Type:           "http",
			Method:         "GET",
			Timeout:        5,
			VerifySSL:      null.NewNullBool(false),
		}
		e, err := CheckHttp(e, false)
		require.Nil(t, err)
		assert.True(t, e.Online)
		assert.False(t, e.LastCheck.IsZero())
		assert.NotEqual(t, 0, e.PingTime)
		assert.NotEqual(t, 0, e.Latency)
	})

	t.Run("Test Load TLS Certificates", func(t *testing.T) {
		e := &Service{
			Name:           "Example TLS",
			Domain:         "http://localhost:15001",
			ExpectedStatus: &[]int{200}[0],
			Type:           "http",
			Method:         "GET",
			Timeout:        5,
			VerifySSL:      null.NewNullBool(false),
			TLSCert:        null.NewNullString(tlsCert),
			TLSCertKey:     null.NewNullString(tlsCertKey),
		}
		customTLS, err := e.LoadTLSCert()
		require.Nil(t, err)

		require.NotNil(t, customTLS)
		assert.Nil(t, customTLS.RootCAs)
		assert.NotNil(t, customTLS.Certificates)
		assert.Equal(t, 1, len(customTLS.Certificates))
	})

	t.Run("Test TLS HTTP Check", func(t *testing.T) {
		e := &Service{
			Name:           "Example TLS HTTP",
			Domain:         "https://localhost:15001",
			ExpectedStatus: &[]int{200}[0],
			Type:           "http",
			Method:         "GET",
			Timeout:        15,
			VerifySSL:      null.NewNullBool(false),
			TLSCert:        null.NewNullString(tlsCert),
			TLSCertKey:     null.NewNullString(tlsCertKey),
		}
		e, err := CheckHttp(e, false)
		require.Nil(t, err)
		assert.True(t, e.Online)
		assert.False(t, e.LastCheck.IsZero())
		assert.NotEqual(t, 0, e.PingTime)
		assert.NotEqual(t, 0, e.Latency)
	})

	t.Run("Test TCP Check", func(t *testing.T) {
		e := &Service{
			Name:    "Example TCP",
			Domain:  "localhost",
			Port:    15002,
			Type:    "tcp",
			Timeout: 5,
		}
		e, err := CheckTcp(e, false)
		require.Nil(t, err)
		assert.True(t, e.Online)
		assert.False(t, e.LastCheck.IsZero())
		assert.NotEqual(t, 0, e.PingTime)
		assert.NotEqual(t, 0, e.Latency)
	})

	t.Run("Test TLS TCP Check", func(t *testing.T) {
		e := &Service{
			Name:       "Example TLS TCP",
			Domain:     "localhost",
			Port:       15003,
			Type:       "tcp",
			Timeout:    15,
			TLSCert:    null.NewNullString(tlsCert),
			TLSCertKey: null.NewNullString(tlsCertKey),
		}
		e, err := CheckTcp(e, false)
		require.Nil(t, err)
		assert.True(t, e.Online)
		assert.False(t, e.LastCheck.IsZero())
		assert.NotEqual(t, 0, e.PingTime)
		assert.NotEqual(t, 0, e.Latency)
	})

	t.Run("Test UDP Check", func(t *testing.T) {
		e := &Service{
			Name:    "Example UDP",
			Domain:  "localhost",
			Port:    15003,
			Type:    "udp",
			Timeout: 5,
		}
		e, err := CheckTcp(e, false)
		require.Nil(t, err)
		assert.True(t, e.Online)
		assert.False(t, e.LastCheck.IsZero())
		assert.NotEqual(t, 0, e.PingTime)
		assert.NotEqual(t, 0, e.Latency)
	})

	t.Run("Test gRPC Check", func(t *testing.T) {
		e := &Service{
			Name:    "Example gRPC",
			Domain:  "localhost",
			Port:    15004,
			Type:    "grpc",
			Timeout: 5,
		}
		e, err := CheckGrpc(e, false)
		require.Nil(t, err)
		assert.True(t, e.Online)
		assert.False(t, e.LastCheck.IsZero())
		assert.NotEqual(t, 0, e.PingTime)
		assert.NotEqual(t, 0, e.Latency)
	})

	t.Run("Test ICMP Check", func(t *testing.T) {
		e := &Service{
			Name:    "Example ICMP",
			Domain:  "localhost",
			Type:    "icmp",
			Timeout: 5,
		}
		e, err := CheckIcmp(e, false)
		require.Nil(t, err)
		assert.True(t, e.Online)
		assert.False(t, e.LastCheck.IsZero())
		assert.NotEqual(t, 0, e.PingTime)
		assert.NotEqual(t, 0, e.Latency)
	})

	t.Run("Test All", func(t *testing.T) {
		items := All()
		assert.Len(t, items, 1)
	})

	t.Run("Test Checkins", func(t *testing.T) {
		item, err := Find(1)
		require.Nil(t, err)
		assert.Len(t, item.Checkins, 1)
	})

	t.Run("Test All Hits", func(t *testing.T) {
		item, err := Find(1)
		require.Nil(t, err)
		assert.Len(t, item.AllHits().List(), 3)
		assert.Equal(t, 3, item.AllHits().Count())
	})

	t.Run("Test All Failures", func(t *testing.T) {
		item, err := Find(1)
		require.Nil(t, err)
		assert.Len(t, item.AllFailures().List(), 2)
		assert.Equal(t, 2, item.AllFailures().Count())
	})

	t.Run("Test First Hit", func(t *testing.T) {
		item, err := Find(1)
		require.Nil(t, err)
		hit := item.FirstHit()
		assert.Equal(t, int64(1), hit.Id)
	})

	t.Run("Test Last Hit", func(t *testing.T) {
		item, err := Find(1)
		require.Nil(t, err)
		hit := item.AllHits().Last()
		assert.Equal(t, int64(3), hit.Id)
	})

	t.Run("Test Last Failure", func(t *testing.T) {
		item, err := Find(1)
		require.Nil(t, err)
		fail := item.AllFailures().Last()
		assert.Equal(t, int64(2), fail.Id)
	})

	t.Run("Test Duration", func(t *testing.T) {
		item, err := Find(1)
		require.Nil(t, err)
		assert.Equal(t, float64(30), item.Duration().Seconds())
	})

	t.Run("Test Count Hits", func(t *testing.T) {
		item, err := Find(1)
		require.Nil(t, err)
		count := item.AllHits().Count()
		assert.NotZero(t, count)
	})

	t.Run("Test Average Time", func(t *testing.T) {
		item, err := Find(1)
		require.Nil(t, err)

		assert.Equal(t, int64(123456), item.AvgTime())
	})

	t.Run("Test Hits Since", func(t *testing.T) {
		item, err := Find(1)
		require.Nil(t, err)

		count := item.HitsSince(utils.Now().Add(-60 * time.Second))
		assert.Equal(t, 1, count.Count())

		count = item.HitsSince(utils.Now().Add(-360 * time.Second))
		assert.Equal(t, 3, count.Count())
	})

	t.Run("Test Service Running", func(t *testing.T) {
		item, err := Find(1)
		require.Nil(t, err)
		assert.False(t, item.IsRunning())
	})

	t.Run("Test Online Percent", func(t *testing.T) {
		item, err := Find(1)
		require.Nil(t, err)

		amount := item.OnlineDaysPercent(1)

		assert.Equal(t, float32(33.33), amount)
	})

	t.Run("Test Downtime", func(t *testing.T) {
		item, err := Find(1)
		require.Nil(t, err)
		amount := item.Downtime().Seconds()
		assert.GreaterOrEqual(t, int64(amount), int64(75))
	})

	t.Run("Test Failures Since", func(t *testing.T) {
		item, err := Find(1)
		require.Nil(t, err)

		count := item.FailuresSince(utils.Now().Add(-30 * time.Second))
		assert.Equal(t, 1, count.Count())

		count = item.FailuresSince(utils.Now().Add(-180 * time.Second))
		assert.Equal(t, 2, count.Count())
	})

	t.Run("Test Create", func(t *testing.T) {
		example := &Service{
			Name:           "Example Service 2",
			Domain:         "https://slack.statping.com",
			ExpectedStatus: &[]int{200}[0],
			Interval:       10,
			Type:           "http",
			Method:         "GET",
			Timeout:        5,
			Order:          3,
			VerifySSL:      null.NewNullBool(true),
			Public:         null.NewNullBool(false),
			GroupId:        1,
			Permalink:      null.NewNullString("statping2"),
		}
		err := example.Create()
		require.Nil(t, err)
		assert.NotZero(t, example.Id)
		assert.Equal(t, "Example Service 2", example.Name)
		assert.False(t, example.Public.Bool)
		assert.NotZero(t, example.CreatedAt)
		assert.Equal(t, int64(2), example.Id)
		assert.Len(t, allServices, 2)
	})

	t.Run("Test Update Service", func(t *testing.T) {
		item, err := Find(1)
		require.Nil(t, err)
		item.Name = "Updated Service"
		item.Order = 1
		err = item.Update()
		require.Nil(t, err)
		assert.Equal(t, int64(1), item.Id)
		assert.Equal(t, "Updated Service", item.Name)
	})

	t.Run("Test In Order", func(t *testing.T) {
		inOrder := AllInOrder()
		assert.Len(t, inOrder, 2)
		assert.Equal(t, "Updated Service", inOrder[0].Name)
		assert.Equal(t, "Example Service 2", inOrder[1].Name)
	})

	t.Run("Test Delete", func(t *testing.T) {
		all := All()
		assert.Len(t, all, 2)

		item, err := Find(1)
		require.Nil(t, err)
		assert.Equal(t, int64(1), item.Id)

		err = item.Delete()
		require.Nil(t, err)

		// after deleted service, make sure checkins, failures, hits, and incidents are also delete
		assert.Len(t, item.AllFailures().List(), 0)
		assert.Len(t, item.AllHits().List(), 0)

		checkin := item.Checkins
		assert.Len(t, checkin, 0)
		for _, c := range checkin {
			assert.Len(t, c.Failures().List(), 0)
			assert.Len(t, c.Hits(), 0)
			assert.False(t, c.IsRunning())
		}

		inc := item.Incidents
		assert.Len(t, inc, 0)
		for _, i := range inc {
			assert.Len(t, i.Updates, 0)
		}

		all = All()
		assert.Len(t, all, 1)
	})

	t.Run("Test Load services.yml", func(t *testing.T) {

		file := `x-tcpservice: &tcpservice
  type: tcp
  check_interval: 60
  timeout: 15
  allow_notifications: true
  notify_after: 0
  notify_all_changes: true
  public: true
  redirect: true

x-httpservice: &httpservice
  type: http
  method: GET
  check_interval: 45
  timeout: 10
  expected_status: 200
  allow_notifications: true
  notify_after: 2
  notify_all_changes: true
  public: true
  redirect: true

services:

  - name: Statping Demo
    domain: https://demo.statping.com
    <<: *httpservice

  - name: Portainer
    domain: portainer
    port: 9000
    <<: *tcpservice

  - name: Statping Github
    domain: https://github.com/statping/statping
    <<: *httpservice`

		err := utils.SaveFile(utils.Directory+"/services.yml", []byte(file))
		require.Nil(t, err)

		assert.FileExists(t, utils.Directory+"/services.yml")

		srvs, err := LoadServicesYaml()
		require.Nil(t, err)
		require.Equal(t, 3, len(srvs.Services))

		assert.Equal(t, "Statping Demo", srvs.Services[0].Name)
		assert.Equal(t, 45, srvs.Services[0].Interval)
		assert.Equal(t, "https://demo.statping.com", srvs.Services[0].Domain)

		err = utils.DeleteFile(utils.Directory + "/services.yml")
		require.Nil(t, err)
		assert.NoFileExists(t, utils.Directory+"/services.yml")
	})
}

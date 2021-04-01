package services

import (
	"fmt"
	"net"
	"strconv"
	"strings"
	"testing"

	"github.com/statping/statping/types/null"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
)

// grpcServerDef is function type.
// Consumed by Test data.
type grpcServerDef func(int, bool) *grpc.Server

// Test Data: Simulates testing scenarios
var testdata = []struct {
	grpcService   grpcServerDef
	clientChecker *Service
}{
	{
		grpcService: func(port int, enableHealthCheck bool) *grpc.Server {
			return grpcServer(port, enableHealthCheck)
		},
		clientChecker: &Service{
			Name:            "GRPC Server with Health check",
			Domain:          "localhost",
			Port:            50053,
			Expected:        null.NewNullString("status:SERVING"),
			ExpectedStatus:  &[]int{1}[0],
			Type:            "grpc",
			Timeout:         3,
			VerifySSL:       null.NewNullBool(false),
			GrpcHealthCheck: null.NewNullBool(true),
		},
	},
	{
		grpcService: func(port int, enableHealthCheck bool) *grpc.Server {
			return grpcServer(port, enableHealthCheck)
		},
		clientChecker: &Service{
			Name:            "Check TLS endpoint on GRPC Server with TLS disabled",
			Domain:          "localhost",
			Port:            50054,
			Expected:        null.NewNullString(""),
			ExpectedStatus:  &[]int{0}[0],
			Type:            "grpc",
			Timeout:         1,
			VerifySSL:       null.NewNullBool(true),
			GrpcHealthCheck: null.NewNullBool(true),
		},
	},
	{
		grpcService: func(port int, enableHealthCheck bool) *grpc.Server {
			return grpcServer(port, enableHealthCheck)
		},
		clientChecker: &Service{
			Name:           "Check GRPC Server without Health check endpoint",
			Domain:         "localhost",
			Port:           50055,
			Expected:       null.NewNullString(""),
			ExpectedStatus: &[]int{0}[0],
			Type:           "grpc",
			Timeout:        1,
			VerifySSL:      null.NewNullBool(false),
		},
	},
	{
		grpcService: func(port int, enableHealthCheck bool) *grpc.Server {
			return grpcServer(50056, enableHealthCheck)
		},
		clientChecker: &Service{
			Name:            "Check where no GRPC Server exists",
			Domain:          "localhost",
			Port:            1000,
			Expected:        null.NewNullString(""),
			ExpectedStatus:  &[]int{0}[0],
			Type:            "grpc",
			Timeout:         1,
			VerifySSL:       null.NewNullBool(false),
			GrpcHealthCheck: null.NewNullBool(true),
		},
	},
	{
		grpcService: func(port int, enableHealthCheck bool) *grpc.Server {
			return grpcServer(50057, enableHealthCheck)
		},
		clientChecker: &Service{
			Name:            "Check where no GRPC Server exists (Verify TLS)",
			Domain:          "localhost",
			Port:            1000,
			Expected:        null.NewNullString(""),
			ExpectedStatus:  &[]int{0}[0],
			Type:            "grpc",
			Timeout:         1,
			VerifySSL:       null.NewNullBool(true),
			GrpcHealthCheck: null.NewNullBool(true),
		},
	},
	{
		grpcService: func(port int, enableHealthCheck bool) *grpc.Server {
			return grpcServer(port, enableHealthCheck)
		},
		clientChecker: &Service{
			Name:            "Check GRPC Server with url",
			Domain:          "http://localhost",
			Port:            50058,
			Expected:        null.NewNullString("status:SERVING"),
			ExpectedStatus:  &[]int{1}[0],
			Type:            "grpc",
			Timeout:         1,
			VerifySSL:       null.NewNullBool(false),
			GrpcHealthCheck: null.NewNullBool(true),
		},
	},
	{
		grpcService: func(port int, enableHealthCheck bool) *grpc.Server {
			return grpcServer(port, enableHealthCheck)
		},
		clientChecker: &Service{
			Name:            "Unparseable Url Error",
			Domain:          "http://local//host",
			Port:            50059,
			Expected:        null.NewNullString(""),
			ExpectedStatus:  &[]int{0}[0],
			Type:            "grpc",
			Timeout:         1,
			VerifySSL:       null.NewNullBool(false),
			GrpcHealthCheck: null.NewNullBool(true),
		},
	},
	{
		grpcService: func(port int, enableHealthCheck bool) *grpc.Server {
			return grpcServer(50060, enableHealthCheck)
		},
		clientChecker: &Service{
			Name:           "Check GRPC on HTTP server",
			Domain:         "https://google.com",
			Port:           443,
			Expected:       null.NewNullString(""),
			ExpectedStatus: &[]int{0}[0],
			Type:           "grpc",
			Timeout:        1,
			VerifySSL:      null.NewNullBool(false),
		},
	},
	{
		grpcService: func(port int, enableHealthCheck bool) *grpc.Server {
			return grpcServer(port, true)
		},
		clientChecker: &Service{
			Name:            "GRPC HealthCheck where health check endpoint is not implemented",
			Domain:          "http://localhost",
			Port:            50061,
			Expected:        null.NewNullString(""),
			ExpectedStatus:  &[]int{0}[0],
			Type:            "grpc",
			Timeout:         1,
			VerifySSL:       null.NewNullBool(false),
			GrpcHealthCheck: null.NewNullBool(false),
		},
	},
}

// grpcServer creates grpc Service with optional parameters.
func grpcServer(port int, enableHealthCheck bool) *grpc.Server {
	portString := strconv.Itoa(port)
	server := grpc.NewServer()
	lis, err := net.Listen("tcp", "localhost:"+portString)
	if err != nil {
		fmt.Println(err)
	}

	if enableHealthCheck {
		healthServer := health.NewServer()
		healthServer.SetServingStatus("Test GRPC Service", healthpb.HealthCheckResponse_SERVING)
		healthpb.RegisterHealthServer(server, healthServer)
		go server.Serve(lis)
	}
	return server
}

// TestCheckGrpc ranges over the testdata struct.
// Examines checkGrpc() function
func TestCheckGrpc(t *testing.T) {
	for _, testscenario := range testdata {
		v := testscenario
		t.Run(v.clientChecker.Name, func(t *testing.T) {
			t.Parallel()
			server := v.grpcService(v.clientChecker.Port, v.clientChecker.GrpcHealthCheck.Bool)
			defer server.Stop()
			v.clientChecker.CheckService(false)
			if v.clientChecker.LastStatusCode != *v.clientChecker.ExpectedStatus || strings.TrimSpace(v.clientChecker.LastResponse) != v.clientChecker.Expected.String {
				t.Errorf("Expected message: '%v', Got message: '%v' , Expected Status: '%v', Got Status: '%v'", v.clientChecker.Expected.String, v.clientChecker.LastResponse, v.clientChecker.ExpectedStatus, v.clientChecker.LastStatusCode)
			}
		})
	}
}

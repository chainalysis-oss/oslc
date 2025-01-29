package main

import (
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/health"
	healthgrpc "google.golang.org/grpc/health/grpc_health_v1"
	"net"
	"strings"
	"testing"
)

func TestHealthcheckAction(t *testing.T) {
	// check the health of the grpc server pointed to by config.Grpc.Interface and config.Grpc.Port.
	// The service checked should be the "" service.

	tlsCreds, err := credentials.NewServerTLSFromFile("../../build/tls/oslc-request-server.internal.crt", "../../build/tls/oslc-request-server.internal.key")
	require.NoError(t, err)
	grpcServer := grpc.NewServer(grpc.Creds(tlsCreds))
	healthcheck := health.NewServer()
	healthgrpc.RegisterHealthServer(grpcServer, healthcheck)
	lis, err := net.Listen("tcp", "127.0.0.1:")
	require.NoError(t, err)
	go grpcServer.Serve(lis)
	defer grpcServer.Stop()
	p := strings.Split(lis.Addr().String(), ":")[1]
	t.Setenv("OSLC_GRPC_INTERFACE", "127.0.0.1")
	t.Setenv("OSLC_GRPC_PORT", p)

	require.NoError(t, healthcheckAction(nil))
	healthcheck.SetServingStatus("", healthgrpc.HealthCheckResponse_NOT_SERVING)
	require.Error(t, healthcheckAction(nil))
}

package grpc

import (
	"bytes"
	"context"
	oslcv1 "github.com/chainalysis-oss/oslc/gen/oslc/v1"
	grpcmock "github.com/chainalysis-oss/oslc/mocks/oslc/grpc"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/require"
	"io"
	"log/slog"
	"net"
	"testing"
)

func TestServer_Serve(t *testing.T) {
	listener, err := net.Listen("tcp", ":0")
	require.NoError(t, err)
	mock := grpcmock.NewMockgrpcServer(t)
	mock.EXPECT().Serve(listener).Return(nil)

	s := &Server{
		options: &serverOptions{
			Logger: slog.New(slog.NewTextHandler(io.Discard, nil)),
		},
		gprcServer: mock,
	}

	_ = s.Serve(listener)
}

func TestServer_GracefulStop(t *testing.T) {
	var logs bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&logs, nil))
	mock := grpcmock.NewMockgrpcServer(t)
	mock.EXPECT().GracefulStop().Return()

	s := &Server{
		options: &serverOptions{
			Logger: logger,
		},
		gprcServer: mock,
	}
	s.GracefulStop()
	require.Contains(t, logs.String(), "stopping grpc server")
	require.Contains(t, logs.String(), "grpc server stopped")
}

func TestServer_RegisterService(t *testing.T) {
	svc := &oslcv1.UnimplementedOslcServiceServer{}
	svcDesc := oslcv1.OslcService_ServiceDesc
	mock := grpcmock.NewMockgrpcServer(t)
	mock.EXPECT().RegisterService(&svcDesc, svc)

	s := &Server{
		options:    &serverOptions{},
		gprcServer: mock,
	}
	s.RegisterService(&svcDesc, svc)
}

func TestServer_GetServiceInfo(t *testing.T) {
	mock := grpcmock.NewMockgrpcServer(t)
	mock.EXPECT().GetServiceInfo().Return(nil)

	s := &Server{
		options:    &serverOptions{},
		gprcServer: mock,
	}
	s.GetServiceInfo()
}

func TestInterceptorLogger(t *testing.T) {
	cases := []struct {
		level    logging.Level
		levelStr string
	}{
		{logging.LevelDebug, "DEBUG"},
		{logging.LevelInfo, "INFO"},
		{logging.LevelWarn, "WARN"},
		{logging.LevelError, "ERROR"},
	}
	for _, c := range cases {
		var logs bytes.Buffer
		logger := slog.New(slog.NewTextHandler(&logs, &slog.HandlerOptions{Level: slog.LevelDebug}))
		log := interceptorLogger(logger)
		require.NotNil(t, log)
		log.Log(context.Background(), c.level, "test", "key", "value")
		require.Contains(t, logs.String(), "msg=test")
		require.Contains(t, logs.String(), "key=value")
		require.Contains(t, logs.String(), "level="+c.levelStr)
	}
}

func TestInterceptorLogger_Panics_unknown_level(t *testing.T) {
	var logs bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&logs, &slog.HandlerOptions{Level: slog.LevelDebug}))
	log := interceptorLogger(logger)
	require.NotNil(t, log)
	require.Panics(t, func() {
		log.Log(context.Background(), 100, "test")
	})
}

func TestNewGrpcRecoveryHandler(t *testing.T) {
	counter := prometheus.NewCounter(prometheus.CounterOpts{})

	var logs bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&logs, nil))
	handler := newGrpcRecoveryHandler(logger, counter)
	require.NotNil(t, handler)
	require.NotPanics(t, func() {
		_ = handler(nil)
	})
	require.Contains(t, logs.String(), "recovered from panic")
	// Canonical way of retrieving the value from a prometheus.Counter: https://github.com/prometheus/client_golang/issues/486
	require.Equal(t, 1, int(testutil.ToFloat64(counter)))
}

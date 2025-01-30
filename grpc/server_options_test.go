package grpc

import (
	"buf.build/gen/go/chainalysis-oss/oslc/grpc/go/chainalysis_oss/oslc/v1alpha/oslcv1alphagrpc"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/require"
	"io"
	"log/slog"
	"os"
	"testing"
)

func TestNewServer(t *testing.T) {
	server, err := NewServer(WithLogger(slog.New(slog.NewTextHandler(io.Discard, nil))))
	require.NoError(t, err)
	require.NotNil(t, server)
	require.Equal(t, slog.New(slog.NewTextHandler(io.Discard, nil)), server.options.Logger)
}

func TestNewServer_globalOptionsAreApplied(t *testing.T) {
	optCopy := make([]ServerOption, len(globalServerOptions))
	copy(optCopy, globalServerOptions)
	defer func() {
		globalServerOptions = optCopy
	}()

	globalServerOptions = append(globalServerOptions, WithLogger(slog.New(slog.NewTextHandler(io.Discard, nil))))
	server, err := NewServer()
	require.NoError(t, err)
	require.Equal(t, slog.New(slog.NewTextHandler(io.Discard, nil)), server.options.Logger)
}

func TestFuncClientOption_apply(t *testing.T) {
	opts := serverOptions{}
	fdo := newFuncClientOption(func(o *serverOptions) {
		o.Logger = slog.New(slog.NewTextHandler(io.Discard, nil))
	})
	fdo.apply(&opts)
	require.Equal(t, slog.New(slog.NewTextHandler(io.Discard, nil)), opts.Logger)
}

func TestNewFuncClientOption(t *testing.T) {
	fdo := newFuncClientOption(func(o *serverOptions) {
		o.Logger = slog.New(slog.NewTextHandler(io.Discard, nil))
	})
	require.NotNil(t, fdo)
}

func TestWithLogger(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	require.NotNil(t, logger)
	opts := serverOptions{}
	f := WithLogger(logger)
	f.apply(&opts)
	require.Equal(t, logger, opts.Logger)
}

func TestWithPanicsTotalCounter(t *testing.T) {
	counter := prometheus.NewCounter(prometheus.CounterOpts{})
	require.NotNil(t, counter)
	opts := serverOptions{}
	f := WithPanicsTotalCounter(counter)
	f.apply(&opts)
	require.Equal(t, counter, opts.PanicsTotalCounter)
}

func TestWithPrometheusRegistry(t *testing.T) {
	registry := prometheus.NewRegistry()
	require.NotNil(t, registry)
	opts := serverOptions{}
	f := WithPrometheusRegistry(registry)
	f.apply(&opts)
	require.Equal(t, registry, opts.PrometheusRegistry)
}

func TestWithOslcServiceServer(t *testing.T) {
	thing := oslcv1alphagrpc.UnimplementedOslcServiceServer{}
	require.NotNil(t, thing)
	opts := serverOptions{}
	f := WithOslcServiceServer(thing)
	f.apply(&opts)
	require.Equal(t, thing, opts.oslcv1alphagrpc)
}

func TestWithTLS(t *testing.T) {
	opts := serverOptions{}
	f := WithTLS("certFile", "keyFile")
	f.apply(&opts)
	require.Equal(t, "certFile", opts.CertFile)
	require.Equal(t, "keyFile", opts.KeyFile)
}

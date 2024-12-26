package grpc

import (
	oslcv1alpha "github.com/chainalysis-oss/oslc/gen/oslc/oslc/v1alpha"
	grpcprom "github.com/grpc-ecosystem/go-grpc-middleware/providers/prometheus"
	"github.com/prometheus/client_golang/prometheus"
	"log/slog"
)

type serverOptions struct {
	Logger             *slog.Logger
	Metrics            *grpcprom.ServerMetrics
	PanicsTotalCounter prometheus.Counter
	PrometheusRegistry *prometheus.Registry
	oslcv1alpha        oslcv1alpha.OslcServiceServer
	CertFile           string
	KeyFile            string
}

var defaultServerOptions = serverOptions{
	Logger: slog.Default(),
	Metrics: grpcprom.NewServerMetrics(
		grpcprom.WithServerHandlingTimeHistogram(
			grpcprom.WithHistogramBuckets([]float64{0.001, 0.01, 0.1, 0.3, 0.6, 1, 3, 6, 9, 20, 30, 60, 90, 120}),
		),
	),
}

var globalServerOptions []ServerOption

// ServerOption is an option for configuring a Client.
type ServerOption interface {
	apply(*serverOptions)
}

// funcServerOption is a ServerOption that calls a function.
// It is used to wrap a function, so it satisfies the ServerOption interface.
type funcServerOption struct {
	f func(*serverOptions)
}

func (fdo *funcServerOption) apply(opts *serverOptions) {
	fdo.f(opts)
}

func newFuncClientOption(f func(*serverOptions)) *funcServerOption {
	return &funcServerOption{
		f: f,
	}
}

// WithLogger returns a ServerOption that uses the provided logger.
func WithLogger(logger *slog.Logger) ServerOption {
	return newFuncClientOption(func(opts *serverOptions) {
		opts.Logger = logger
	})
}

// WithPanicsTotalCounter returns a ServerOption that uses the provided counter.
func WithPanicsTotalCounter(counter prometheus.Counter) ServerOption {
	return newFuncClientOption(func(opts *serverOptions) {
		opts.PanicsTotalCounter = counter
	})
}

// WithPrometheusRegistry returns a ServerOption that uses the provided registry.
func WithPrometheusRegistry(registry *prometheus.Registry) ServerOption {
	return newFuncClientOption(func(opts *serverOptions) {
		opts.PrometheusRegistry = registry
	})
}

// WithOslcServiceServer returns a ServerOption that uses the provided OslcServiceServer.
func WithOslcServiceServer(oslcv1alpha oslcv1alpha.OslcServiceServer) ServerOption {
	return newFuncClientOption(func(opts *serverOptions) {
		opts.oslcv1alpha = oslcv1alpha
	})
}

// WithTLS returns a ServerOption that uses the provided TLS configuration.
func WithTLS(certFile, keyFile string) ServerOption {
	return newFuncClientOption(func(opts *serverOptions) {
		opts.CertFile = certFile
		opts.KeyFile = keyFile
	})
}

package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"log/slog"
)

type serverOptions struct {
	Logger             *slog.Logger
	PrometheusRegistry *prometheus.Registry
}

var defaultServerOptions = serverOptions{
	Logger: slog.Default(),
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

// WithPrometheusRegistry returns a ServerOption that uses the provided registry.
func WithPrometheusRegistry(registry *prometheus.Registry) ServerOption {
	return newFuncClientOption(func(opts *serverOptions) {
		opts.PrometheusRegistry = registry
	})
}

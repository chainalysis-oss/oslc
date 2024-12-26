package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"log/slog"
	"net"
	"net/http"
)

// httpServer is an interface designed to allow easy testing of the [http.Server].
type httpServer interface {
	Serve(l net.Listener) error
	Close() error
}

type Server struct {
	options    *serverOptions
	httpServer httpServer
}

func NewServer(options ...ServerOption) (*Server, error) {
	opts := defaultServerOptions
	for _, opt := range globalServerOptions {
		opt.apply(&opts)
	}
	for _, opt := range options {
		opt.apply(&opts)
	}

	// We cannot create a default PrometheusRegistry in the defaultServerOptions, as it would be shared between
	// multiple instances of the Server. This would cause a panic when registering the same collector multiple times.
	if opts.PrometheusRegistry == nil {
		opts.PrometheusRegistry = prometheus.NewRegistry()
	}

	opts.PrometheusRegistry.MustRegister(collectors.NewGoCollector())
	opts.PrometheusRegistry.MustRegister(collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}))
	opts.PrometheusRegistry.MustRegister(collectors.NewBuildInfoCollector())

	handler := http.NewServeMux()
	handler.Handle("/metrics", httpLoggerMiddleware(opts.Logger, promhttp.HandlerFor(opts.PrometheusRegistry, promhttp.HandlerOpts{
		EnableOpenMetrics: true,
	})))

	s := &Server{
		options: &opts,
		httpServer: &http.Server{
			Handler: handler,
		},
	}

	return s, nil
}

func (s *Server) Serve(l net.Listener) error {
	s.options.Logger.Info("starting metrics server", slog.String("address", l.Addr().String()))
	return s.httpServer.Serve(l)
}

func (s *Server) Close() error {
	s.options.Logger.Info("stopping metrics server")
	err := s.httpServer.Close()
	if err != nil {
		s.options.Logger.Error("failed to close metrics server", slog.String("error", err.Error()))
	} else {
		s.options.Logger.Info("metrics server stopped")
	}
	return err
}

func (s *Server) GetPrometheusRegistry() *prometheus.Registry {
	return s.options.PrometheusRegistry
}

func httpLoggerMiddleware(l *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r)
		l.Info("http request", slog.String("method", r.Method), slog.String("url", r.URL.String()), slog.String("remote", r.RemoteAddr))
	})
}

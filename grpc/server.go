package grpc

import (
	"buf.build/gen/go/chainalysis-oss/oslc/grpc/go/chainalysis_oss/oslc/v1alpha/oslcv1alphagrpc"
	"context"
	"fmt"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/recovery"
	"github.com/prometheus/client_golang/prometheus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/health"
	healthgrpc "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
	"log/slog"
	"net"
	"runtime/debug"
)

// grpcServer is an interface designed to allow easy testing of the [grpc.Server].
type grpcServer interface {
	Serve(l net.Listener) error
	GracefulStop()
	RegisterService(sd *grpc.ServiceDesc, ss any)
	GetServiceInfo() map[string]grpc.ServiceInfo
}

type Server struct {
	options    *serverOptions
	gprcServer grpcServer
}

func NewServer(options ...ServerOption) (*Server, error) {
	opts := defaultServerOptions
	for _, opt := range globalServerOptions {
		opt.apply(&opts)
	}
	for _, opt := range options {
		opt.apply(&opts)
	}

	unaryInterceptors := make([]grpc.UnaryServerInterceptor, 0)
	if opts.Metrics != nil {
		unaryInterceptors = append(unaryInterceptors, opts.Metrics.UnaryServerInterceptor())
	}
	unaryInterceptors = append(unaryInterceptors, logging.UnaryServerInterceptor(interceptorLogger(opts.Logger)))
	unaryInterceptors = append(unaryInterceptors, newGrpcErrorHandler(opts.Logger))
	unaryInterceptors = append(unaryInterceptors, recovery.UnaryServerInterceptor(recovery.WithRecoveryHandler(newGrpcRecoveryHandler(opts.Logger, opts.PanicsTotalCounter))))

	grpcOpts := make([]grpc.ServerOption, 0)
	grpcOpts = append(grpcOpts, grpc.ChainUnaryInterceptor(unaryInterceptors...))

	if opts.CertFile != "" || opts.KeyFile != "" {
		creds, err := credentials.NewServerTLSFromFile(opts.CertFile, opts.KeyFile)
		if err != nil {
			return nil, fmt.Errorf("failed to create TLS credentials: %w", err)
		}
		grpcOpts = append(grpcOpts, grpc.Creds(creds))
	}

	s := &Server{
		options:    &opts,
		gprcServer: grpc.NewServer(grpcOpts...),
	}

	healthcheck := health.NewServer()
	healthgrpc.RegisterHealthServer(s.gprcServer, healthcheck)
	oslcv1alphagrpc.RegisterOslcServiceServer(s.gprcServer, opts.oslcv1alphagrpc)
	reflection.Register(s.gprcServer)
	if opts.Metrics != nil {
		opts.Metrics.InitializeMetrics(s.gprcServer)
		if opts.PrometheusRegistry != nil {
			opts.PrometheusRegistry.MustRegister(opts.Metrics)
		}
	}
	return s, nil
}

func (s *Server) Serve(l net.Listener) error {
	s.options.Logger.Info("starting grpc server", slog.String("address", l.Addr().String()))
	return s.gprcServer.Serve(l)
}

func (s *Server) GracefulStop() {
	s.options.Logger.Info("stopping grpc server")
	s.gprcServer.GracefulStop()
	s.options.Logger.Info("grpc server stopped")
}

func (s *Server) RegisterService(sd *grpc.ServiceDesc, ss any) {
	s.gprcServer.RegisterService(sd, ss)
}

func (s *Server) GetServiceInfo() map[string]grpc.ServiceInfo {
	return s.gprcServer.GetServiceInfo()
}

func interceptorLogger(l *slog.Logger) logging.Logger {
	return logging.LoggerFunc(func(ctx context.Context, lvl logging.Level, msg string, fields ...any) {
		switch lvl {
		case logging.LevelDebug:
			l.DebugContext(ctx, msg, fields...)
		case logging.LevelInfo:
			l.InfoContext(ctx, msg, fields...)
		case logging.LevelWarn:
			l.WarnContext(ctx, msg, fields...)
		case logging.LevelError:
			l.ErrorContext(ctx, msg, fields...)
		default:
			panic(fmt.Sprintf("unknown level %v", lvl))
		}
	})
}

func newGrpcRecoveryHandler(logger *slog.Logger, panicsTotal prometheus.Counter) recovery.RecoveryHandlerFunc {
	return func(p any) (err error) {
		if panicsTotal != nil {
			panicsTotal.Inc()
		}
		logger.Error("recovered from panic", slog.String("panic", fmt.Sprintf("%v", p)), "stack", fmt.Sprintf("%v", debug.Stack()))
		return status.Errorf(codes.Internal, "%s", p)
	}
}

func newGrpcErrorHandler(logger *slog.Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		m, err := handler(ctx, req)
		if err == nil {
			return m, err
		}

		_, ok := status.FromError(err)
		if ok {
			return m, err
		}
		logger.Error("non-grpc error encountered, returning internal server error instead", slog.String("error", err.Error()))
		return m, InternalServerError
	}
}

// InternalServerError is a status error that represents an internal server error. This is a fairly non-descriptive
// error intended to hide the details of the error from the client. It should rarely be used, and should generally be
// indicative of a bug in error handling somewhere in the call chain.
var InternalServerError = status.Error(codes.Internal, "internal server error")

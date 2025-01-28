package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/chainalysis-oss/oslc/cratesio"
	"github.com/chainalysis-oss/oslc/goproxy"
	"github.com/chainalysis-oss/oslc/grpc"
	"github.com/chainalysis-oss/oslc/maven"
	"github.com/chainalysis-oss/oslc/metrics"
	"github.com/chainalysis-oss/oslc/npm"
	"github.com/chainalysis-oss/oslc/oslc"
	"github.com/chainalysis-oss/oslc/postgres"
	"github.com/chainalysis-oss/oslc/pypi"
	"github.com/chainalysis-oss/oslc/sll"
	"github.com/chainalysis-oss/oslc/spdxnormalizer"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"log/slog"
	"net"
	"net/url"
	"os"
	"strconv"
	"syscall"

	"github.com/oklog/run"
	"github.com/prometheus/client_golang/prometheus"
)

var Version = "0.0.0"

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	config, err := createConfiguration("config.json")
	if err != nil {
		logger.Error("failed to create configuration", slog.String("error", translateValidationError(err).Error()))
		os.Exit(1)
	}
	logger = getLogger(config.Log.Level, config.Log.Kind)
	logger.Info("starting oslc-request-server", slog.String("version", Version))

	pypiClient, _ := pypi.NewClient(pypi.WithLogger(logger))
	npmClient, _ := npm.NewClient(npm.WithLogger(logger))
	mavenClient, _ := maven.NewClient(maven.WithLogger(logger))
	cratesioClient, err := cratesio.NewClient(cratesio.WithLogger(logger))
	goClient, err := goproxy.NewClient(goproxy.WithLogger(logger))

	dbPool, err := postgres.NewPool(context.Background(), fmt.Sprintf("postgres://%s:%s@%s:%d/%s", url.QueryEscape(config.Datastore.Username), url.QueryEscape(config.Datastore.Password), config.Datastore.Host, config.Datastore.Port, config.Datastore.Database))
	if err != nil {
		logger.Error("failed to create database pool", slog.String("error", err.Error()))
		os.Exit(1)
	}

	datastore, err := postgres.NewDatastore(
		postgres.WithLogger(logger),
		postgres.WithPool(dbPool))
	if err != nil {
		logger.Error("failed to create datastore", slog.String("error", err.Error()))
		os.Exit(1)
	}

	normalizer, err := spdxnormalizer.NewNormalizer(
		spdxnormalizer.WithLogger(logger),
		spdxnormalizer.WithLicenseRetriever(sll.AsLicenseRetriever()),
	)
	if err != nil {
		logger.Error("failed to create SPDX normalizer", slog.String("error", err.Error()))
		os.Exit(1)
	}

	oslcSrv, err := oslc.NewServer(
		oslc.WithLogger(logger),
		oslc.WithPypiClient(pypiClient),
		oslc.WithNpmClient(npmClient),
		oslc.WithMavenClient(mavenClient),
		oslc.WithCratesIoClient(cratesioClient),
		oslc.WithGoClient(goClient),
		oslc.WithDatastore(datastore),
		oslc.WithLicenseIDNormalizer(normalizer),
	)
	if err != nil {
		logger.Error("failed to create oslc server", slog.String("error", err.Error()))
		os.Exit(1)
	}

	var metricsServer *metrics.Server
	var optionalGrpcServerOptions []grpc.ServerOption

	rpcLogger := logger.With(slog.String("service", "gRPC/server"))
	metricsLogger := logger.With(slog.String("service", "metrics/server"))

	if config.Metrics.Enabled {
		metricsServer, err = metrics.NewServer(
			metrics.WithLogger(metricsLogger),
		)
		if err != nil {
			logger.Error("failed to create metrics server", slog.String("error", err.Error()))
			os.Exit(1)
		}
		optionalGrpcServerOptions = append(optionalGrpcServerOptions, grpc.WithPrometheusRegistry(metricsServer.GetPrometheusRegistry()))
		optionalGrpcServerOptions = append(optionalGrpcServerOptions, grpc.WithPanicsTotalCounter(promauto.With(metricsServer.GetPrometheusRegistry()).NewCounter(prometheus.CounterOpts{
			Name: "grpc_req_panics_recovered_total",
			Help: "Total number of gRPC requests recovered from internal panic.",
		})))
	}

	if !config.Grpc.NoTLS {
		optionalGrpcServerOptions = append(optionalGrpcServerOptions, grpc.WithTLS(config.TLS.CertFile, config.TLS.KeyFile))
	} else {
		logger.Info("gRPC server will not use TLS")
	}

	grpcServerOptions := []grpc.ServerOption{
		grpc.WithLogger(rpcLogger),
		grpc.WithOslcServiceServer(oslcSrv),
	}
	grpcServerOptions = append(grpcServerOptions, optionalGrpcServerOptions...)

	grpcServer, err := grpc.NewServer(grpcServerOptions...)
	if err != nil {
		logger.Error("failed to create grpc server", slog.String("error", err.Error()))
		os.Exit(1)
	}

	listeners, err := NewListeners(config)
	if err != nil {
		logger.Error("failed to create listeners", slog.String("error", err.Error()))
		os.Exit(1)
	}

	g := &run.Group{}

	runGrpcServer(g, grpcServer, listeners.Grpc)

	if config.Metrics.Enabled {
		if metricsServer == nil {
			logger.Error("metrics server is nil - this is almost certainly a bug")
			os.Exit(1)
		}
		runMetricsServer(g, metricsServer, listeners.Metrics)
	}

	g.Add(run.SignalHandler(context.Background(), syscall.SIGINT, syscall.SIGTERM))
	if err := g.Run(); err != nil {
		var sigErr run.SignalError
		if errors.As(err, &sigErr) {
			logger.Debug("received signal", slog.String("signal", sigErr.Signal.String()))
			logger.Info("oslc-request-server has shut down")
			os.Exit(0)
		}
		logger.Error("server error", slog.String("error", err.Error()))
		os.Exit(1)
	}
}

type Listeners struct {
	Grpc    net.Listener
	Metrics net.Listener
}

func NewListeners(config *cfg) (*Listeners, error) {
	listeners := &Listeners{}
	var err error
	listeners.Grpc, err = net.Listen("tcp", net.JoinHostPort(config.Grpc.Interface, strconv.Itoa(config.Grpc.Port)))
	if err != nil {
		return nil, fmt.Errorf("failed to listen on gRPC port: %w", err)
	}

	if config.Metrics.Enabled {
		listeners.Metrics, err = net.Listen("tcp", net.JoinHostPort(config.Metrics.Interface, strconv.Itoa(config.Metrics.Port)))
		if err != nil {
			return nil, fmt.Errorf("failed to listen on metrics port: %w", err)
		}
	}

	return listeners, nil
}

func runGrpcServer(g *run.Group, grpcServer *grpc.Server, listener net.Listener) {
	g.Add(func() error {
		return grpcServer.Serve(listener)
	}, func(error) {
		grpcServer.GracefulStop()
	})
}

func runMetricsServer(g *run.Group, metricsServer *metrics.Server, listener net.Listener) {
	g.Add(func() error {
		return metricsServer.Serve(listener)
	}, func(error) {
		// metricsServer.Close() already logs the error, which is all we'd do here anyway.
		// This is why we ignore the return value.
		_ = metricsServer.Close()
	})
}

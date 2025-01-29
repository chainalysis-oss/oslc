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
	"github.com/urfave/cli/v2"
)

var Version = "0.0.0"

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	app := &cli.App{
		Name:    "OSLC Request Server",
		Usage:   "Run the OSLC request server",
		Action:  rootAction,
		Version: Version,
		Commands: []*cli.Command{
			healthCheckCommand,
		},
	}

	if err := app.Run(os.Args); err != nil {
		logger.Error("failed to run app", slog.String("error", err.Error()))
		os.Exit(1)
	}
}

func rootAction(cCtx *cli.Context) error {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	config, err := createConfiguration("config.json")
	if err != nil {
		return fmt.Errorf("failed to create configuration: %w", err)
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
		return fmt.Errorf("failed to create database pool: %w", err)
	}

	datastore, err := postgres.NewDatastore(
		postgres.WithLogger(logger),
		postgres.WithPool(dbPool))
	if err != nil {
		return fmt.Errorf("failed to create datastore: %w", err)
	}

	normalizer, err := spdxnormalizer.NewNormalizer(
		spdxnormalizer.WithLogger(logger),
		spdxnormalizer.WithLicenseRetriever(sll.AsLicenseRetriever()),
	)
	if err != nil {
		return fmt.Errorf("failed to create SPDX normalizer: %w", err)
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
		return fmt.Errorf("failed to create oslc server: %w", err)
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
			return fmt.Errorf("failed to create metrics server: %w", err)
		}
		optionalGrpcServerOptions = append(optionalGrpcServerOptions, grpc.WithPrometheusRegistry(metricsServer.GetPrometheusRegistry()))
		optionalGrpcServerOptions = append(optionalGrpcServerOptions, grpc.WithPanicsTotalCounter(promauto.With(metricsServer.GetPrometheusRegistry()).NewCounter(prometheus.CounterOpts{
			Name: "grpc_req_panics_recovered_total",
			Help: "Total number of gRPC requests recovered from internal panic.",
		})))
	}

	optionalGrpcServerOptions = append(optionalGrpcServerOptions, grpc.WithTLS(config.TLS.CertFile, config.TLS.KeyFile))

	grpcServerOptions := []grpc.ServerOption{
		grpc.WithLogger(rpcLogger),
		grpc.WithOslcServiceServer(oslcSrv),
	}
	grpcServerOptions = append(grpcServerOptions, optionalGrpcServerOptions...)

	grpcServer, err := grpc.NewServer(grpcServerOptions...)
	if err != nil {
		return fmt.Errorf("failed to create grpc server: %w", err)
	}

	listeners, err := NewListeners(config)
	if err != nil {
		return fmt.Errorf("failed to create listeners: %w", err)
	}

	g := &run.Group{}

	runGrpcServer(g, grpcServer, listeners.Grpc)

	if config.Metrics.Enabled {
		if metricsServer == nil {
			return fmt.Errorf("metrics server is nil - this is almost certainly a bug")
		}
		runMetricsServer(g, metricsServer, listeners.Metrics)
	}

	g.Add(run.SignalHandler(context.Background(), syscall.SIGINT, syscall.SIGTERM))
	if err := g.Run(); err != nil {
		var sigErr run.SignalError
		if errors.As(err, &sigErr) {
			logger.Debug("received signal", slog.String("signal", sigErr.Signal.String()))
			logger.Info("oslc-request-server has shut down")
			return nil
		}
		return fmt.Errorf("server error: %w", err)
	}

	return nil
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

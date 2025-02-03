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
	"strings"
	"syscall"

	_ "embed"
	"github.com/oklog/run"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/urfave/cli/v2"
	"github.com/urfave/cli/v2/altsrc"
)

var Version = "0.0.0"

func main() {
	app := &cli.App{
		Name:  "oslc-request-server",
		Usage: `Serving license information since 2024`,
		Description: `The OSLC Request Server is a gRPC server that provides a unified interface for querying package license information from various package managers.

The application supports configuration in several ways (listed in order of precedence):

1. Command-line flags
2. Environment variables
3. Filesystem paths
4. Configuration file (YAML)
5. Default values

Configuration values are read from the sources by referencing the configuration keys. The configuration keys are
identical to the long flag names used in the command-line interface, without the leading double-dash. For example, the
configuration key for the '--datastore.username' flag is 'datastore.username'.

When reading values from a configuration file, the period (.) in the key denotes a nested structure. For example, the
key 'datastore.username' is used to retrieve the value of the username field in the datastore structure in a YAML
configuration file.

All configuration values can be read from individual files, where the file name is used as the configuration key and
the raw value is used as the configuration value. For filesystem paths, the configuration key has periods (.) replaced
with underscores (_). For example, the configuration key for the '--datastore.username' flag is 'datastore_username'.
The files are read from the filesystem and the raw value is used as the configuration value. The files are expected to
be in the "/run/secrets" directory, however this can be overwritten by setting the "OSLC_CONFIG_FILE_PREFIX" environment
variable to the desired directory. As an example of this, "datastore.username" configuration key can be set by creating
a file at "/run/secrets/datastore.username" with the desired value.

This is useful for providing sensitive information like passwords and API keys without exposing them in the command-line
interface or environment variables.

Additionally, the application supports reading configuration values from environment variables. The environment variable
names are derived from the configuration keys by converting the keys to uppercase and replacing periods (.) with
underscores (_). For example, the environment variable for the 'datastore.username' configuration key is
'OSLC_DATASTORE_USERNAME'.

Finally, the application supports setting configuration values using flags, where the flag names are derived from the
configuration keys by prepending two dashes (--) and replacing periods (.) with hyphens (-). For example, the flag for
the 'datastore.username' configuration key is '--datastore.username'. These are documented further below.
`,
		Action:  rootAction,
		Version: Version,
		Commands: []*cli.Command{
			healthCheckCommand,
			asMarkdownCmd,
		},
		Flags: flags,
		Before: func(cCtx *cli.Context) error {
			err := altsrc.InitInputSourceWithContext(flags, altsrc.NewYamlSourceFromFlagFunc("config"))(cCtx)
			if err != nil {
				// We're forced to use [strings.Contains] here, since urfave/cli doesn't have error types for these
				// cases.
				if strings.Contains(err.Error(), "because it does not exist") {
					return nil
				}
				if strings.Contains(err.Error(), "is a directory") {
					return fmt.Errorf("config path '%s' is a directory", cCtx.String("config"))
				}
				return err
			}
			return nil
		},
		Writer: os.Stdout,
	}

	logger := slog.New(slog.NewJSONHandler(app.Writer, nil))
	if err := app.Run(os.Args); err != nil {
		logger.Error("failed to run app", slog.String("error", err.Error()))
		os.Exit(1)
	}
}

// The asMarkdownCmd command is hidden from the help output, but can be used to generate markdown documentation for the
// application to which it is attached.
var asMarkdownCmd = &cli.Command{
	Name:   "asMarkdown",
	Hidden: true,
	Action: asMarkdownAction,
}

func asMarkdownAction(cCtx *cli.Context) error {
	md, err := cCtx.App.ToMarkdown()
	if err != nil {
		return err
	}
	_, err = fmt.Fprint(cCtx.App.Writer, md)
	if err != nil {
		return err
	}

	if md[len(md)-1] != '\n' {
		_, err = fmt.Fprint(cCtx.App.Writer, "\n")
		if err != nil {
			return err
		}
	}
	return nil
}

func rootAction(cCtx *cli.Context) error {
	logger := slog.New(slog.NewJSONHandler(cCtx.App.Writer, nil))
	logger = getLogger(cCtx.String(configLogLevelKey), cCtx.String(configLogKindKey), cCtx.App.Writer)
	logger.Info("starting oslc-request-server", slog.String("version", Version))

	pypiClient, _ := pypi.NewClient(pypi.WithLogger(logger))
	npmClient, _ := npm.NewClient(npm.WithLogger(logger))
	mavenClient, _ := maven.NewClient(maven.WithLogger(logger))
	cratesioClient, err := cratesio.NewClient(cratesio.WithLogger(logger))
	goClient, err := goproxy.NewClient(goproxy.WithLogger(logger))

	dbPool, err := postgres.NewPool(context.Background(), fmt.Sprintf("postgres://%s:%s@%s:%d/%s", url.QueryEscape(cCtx.String(configDatastoreUsernameKey)), url.QueryEscape(cCtx.String(configDatastorePasswordKey)), cCtx.String(configDatastoreHostKey), cCtx.Int(configDatastorePortKey), cCtx.String(configDatastoreDatabaseKey)))
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

	if cCtx.Bool(configMetricsEnabledKey) {
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

	optionalGrpcServerOptions = append(optionalGrpcServerOptions, grpc.WithTLS(cCtx.String(configTlsCertFilePathKey), cCtx.String(configTlsKeyFilePathKey)))

	grpcServerOptions := []grpc.ServerOption{
		grpc.WithLogger(rpcLogger),
		grpc.WithOslcServiceServer(oslcSrv),
	}
	grpcServerOptions = append(grpcServerOptions, optionalGrpcServerOptions...)

	grpcServer, err := grpc.NewServer(grpcServerOptions...)
	if err != nil {
		return fmt.Errorf("failed to create grpc server: %w", err)
	}

	listeners, err := NewListeners(cCtx)
	if err != nil {
		return fmt.Errorf("failed to create listeners: %w", err)
	}

	g := &run.Group{}

	runGrpcServer(g, grpcServer, listeners.Grpc)

	if cCtx.Bool(configMetricsEnabledKey) {
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

func NewListeners(cCtx *cli.Context) (*Listeners, error) {
	listeners := &Listeners{}
	var err error
	listeners.Grpc, err = net.Listen("tcp", net.JoinHostPort(cCtx.String(configGrpcInterfaceKey), cCtx.String(configGrpcPortKey)))
	if err != nil {
		return nil, fmt.Errorf("failed to listen on gRPC port: %w", err)
	}

	if cCtx.Bool(configMetricsEnabledKey) {
		listeners.Metrics, err = net.Listen("tcp", net.JoinHostPort(cCtx.String(configMetricsInterfaceKey), cCtx.String(configMetricsPortKey)))
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

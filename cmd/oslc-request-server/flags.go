package main

import (
	"fmt"
	"github.com/urfave/cli/v2"
	"github.com/urfave/cli/v2/altsrc"
	"os"
	"path"
	"strings"
)

// configValidationError is an error type for config validation errors.
type configValidationError struct {
	key    string // key is the configuration key that caused the error.
	value  string // value is the value that caused the error.
	detail string // detail is a detailed description of the error for providing extra context.
}

func (e configValidationError) Error() string {
	return fmt.Sprintf("config validation failure: invalid value for key '%s': '%s': %s", e.key, e.value, e.detail)
}

func (e configValidationError) ExitCoder() int {
	return 42
}

const valNotBeEmpty = "value must not be empty"

// The following constants are used to define the configuration keys for the application. These keys are used to
// retrieve the configuration values from the configuration files, and command-line flags.
// When used in a configuration file, the period (.) in the key denotes a nested structure. For example, the key
// "datastore.username" is used to retrieve the value of the username field in the datastore structure for JSON and
// YAML configuration files.
const (
	configDatastoreUsernameKey string = "datastore.username"
	configDatastorePasswordKey string = "datastore.password"
	configDatastoreHostKey     string = "datastore.host"
	configDatastorePortKey     string = "datastore.port"
	configDatastoreDatabaseKey string = "datastore.database"
	configGrpcInterfaceKey     string = "grpc.interface"
	configGrpcPortKey          string = "grpc.port"
	configMetricsEnabledKey    string = "metrics.enabled"
	configMetricsInterfaceKey  string = "metrics.interface"
	configMetricsPortKey       string = "metrics.port"
	configLogLevelKey          string = "log.level"
	configLogKindKey           string = "log.kind"
	configTlsCertFilePathKey   string = "tls.cert_file_path"
	configTlsKeyFilePathKey    string = "tls.key_file_path"
)

// The following constants are used to define the environment variables that can be used to set the configuration
// values for the application.
const (
	configDatastoreUsernameEnv string = "OSLC_DATASTORE_USERNAME"
	configDatastorePasswordEnv string = "OSLC_DATASTORE_PASSWORD"
	configDatastoreHostEnv     string = "OSLC_DATASTORE_HOST"
	configDatastorePortEnv     string = "OSLC_DATASTORE_PORT"
	configDatastoreDatabaseEnv string = "OSLC_DATASTORE_DB"
	configGrpcInterfaceEnv     string = "OSLC_GRPC_INTERFACE"
	configGrpcPortEnv          string = "OSLC_GRPC_PORT"
	configMetricsEnabledEnv    string = "OSLC_METRICS_ENABLED"
	configMetricsInterfaceEnv  string = "OSLC_METRICS_INTERFACE"
	configMetricsPortEnv       string = "OSLC_METRICS_PORT"
	configLogLevelEnv          string = "OSLC_LOG_LEVEL"
	configLogKindEnv           string = "OSLC_LOG_KIND"
	configTlsCertFilePathEnv   string = "OSLC_TLS_CERT_FILE_PATH"
	configTlsKeyFilePathEnv    string = "OSLC_TLS_KEY_FILE_PATH"
)

const filePrefixFallback = "/run/secrets"
const filePrefixEnv = "OSLC_CONFIG_FILE_PREFIX"

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func getFilePathWithPrefix(s string) string {
	return path.Join(getEnv(filePrefixEnv, filePrefixFallback), s)
}

// The following variables are used to define the file paths for the configuration files. These file paths are used to
// load the configuration values from files, where the raw configuration values are stored.
// These are not to be confused with the singular configuration file that is passed in as a command-line flag, which
// can be used to store several configuration values.
//
// These file paths are mostly used to store sensitive configuration values that one does not want to expose in
// environment variables or configuration files.
var (
	configDatastoreUsernameFile = getFilePathWithPrefix(strings.ToLower(configDatastoreUsernameEnv))
	configDatastorePasswordFile = getFilePathWithPrefix(strings.ToLower(configDatastorePasswordEnv))
	configDatastoreHostFile     = getFilePathWithPrefix(strings.ToLower(configDatastoreHostEnv))
	configDatastorePortFile     = getFilePathWithPrefix(strings.ToLower(configDatastorePortEnv))
	configDatastoreDatabaseFile = getFilePathWithPrefix(strings.ToLower(configDatastoreDatabaseEnv))
	configGrpcInterfaceFile     = getFilePathWithPrefix(strings.ToLower(configGrpcInterfaceEnv))
	configGrpcPortFile          = getFilePathWithPrefix(strings.ToLower(configGrpcPortEnv))
	configMetricsEnabledFile    = getFilePathWithPrefix(strings.ToLower(configMetricsEnabledEnv))
	configMetricsInterfaceFile  = getFilePathWithPrefix(strings.ToLower(configMetricsInterfaceEnv))
	configMetricsPortFile       = getFilePathWithPrefix(strings.ToLower(configMetricsPortEnv))
	configLogLevelFile          = getFilePathWithPrefix(strings.ToLower(configLogLevelEnv))
	configLogKindFile           = getFilePathWithPrefix(strings.ToLower(configLogKindEnv))
	configTlsCertFilePathFile   = getFilePathWithPrefix(strings.ToLower(configTlsCertFilePathEnv))
	configTlsKeyFilePathFile    = getFilePathWithPrefix(strings.ToLower(configTlsKeyFilePathEnv))
)

func cfgStringMustNotBeEmpty(key string) func(cCtx *cli.Context, s string) error {
	return func(cCtx *cli.Context, s string) error {
		if s == "" {
			return &configValidationError{key: key, value: s, detail: valNotBeEmpty}
		}
		return nil
	}
}

func cfgIntMustBeValidPort(key string) func(cCtx *cli.Context, i int) error {
	return func(cCtx *cli.Context, i int) error {
		if i < 1 || i > 65535 {
			return &configValidationError{key: key, value: fmt.Sprintf("%d", i), detail: "value must be between 1 and 65535"}
		}
		return nil
	}
}

func cfgStringMustBeValidLoggingLevel(key string) func(cCtx *cli.Context, s string) error {
	return func(cCtx *cli.Context, s string) error {
		switch s {
		case "debug", "info", "warn", "error":
			return nil
		default:
			return &configValidationError{key: key, value: s, detail: "value must be one of debug, info, warn, error"}
		}
	}
}

func cfgStringMustBeValidLoggingKind(key string) func(cCtx *cli.Context, s string) error {
	return func(cCtx *cli.Context, s string) error {
		switch s {
		case "text", "json", "discard":
			return nil
		default:
			return &configValidationError{key: key, value: s, detail: "value must be one of text, json or discard"}
		}
	}
}

var flags = []cli.Flag{
	&cli.StringFlag{
		Name:    "config",
		Aliases: []string{"c"},
		Value:   "config.yaml",
	},
	altsrc.NewStringFlag(&cli.StringFlag{
		Name:     configDatastoreUsernameKey,
		Value:    "postgres",
		Usage:    "Username for OSLC's datastore",
		EnvVars:  []string{configDatastoreUsernameEnv},
		FilePath: configDatastoreUsernameFile,
		Action:   cfgStringMustNotBeEmpty(configDatastoreUsernameKey),
	}),
	altsrc.NewStringFlag(&cli.StringFlag{
		Name:     configDatastorePasswordKey,
		Value:    "postgres",
		Usage:    "Password for OSLC's datastore",
		EnvVars:  []string{configDatastorePasswordEnv},
		FilePath: configDatastorePasswordFile,
		Action:   cfgStringMustNotBeEmpty(configDatastorePasswordKey),
	}),
	altsrc.NewStringFlag(&cli.StringFlag{
		Name:     configDatastoreHostKey,
		Value:    "localhost",
		Usage:    "Host for OSLC's datastore",
		EnvVars:  []string{configDatastoreHostEnv},
		FilePath: configDatastoreHostFile,
		Action:   cfgStringMustNotBeEmpty(configDatastoreHostKey),
	}),
	altsrc.NewIntFlag(&cli.IntFlag{
		Name:     configDatastorePortKey,
		Value:    5432,
		Usage:    "Port for OSLC's datastore",
		EnvVars:  []string{configDatastorePortEnv},
		FilePath: configDatastorePortFile,
		Action:   cfgIntMustBeValidPort(configDatastorePortKey),
	}),
	altsrc.NewStringFlag(&cli.StringFlag{
		Name:     configDatastoreDatabaseKey,
		Value:    "postgres",
		Usage:    "Database for OSLC's datastore",
		EnvVars:  []string{configDatastoreDatabaseEnv},
		FilePath: configDatastoreDatabaseFile,
		Action:   cfgStringMustNotBeEmpty(configDatastoreDatabaseKey),
	}),
	altsrc.NewStringFlag(&cli.StringFlag{
		Name:     configGrpcInterfaceKey,
		Value:    "0.0.0.0",
		Usage:    "Interface for OSLC's gRPC server",
		EnvVars:  []string{configGrpcInterfaceEnv},
		FilePath: configGrpcInterfaceFile,
		Action:   cfgStringMustNotBeEmpty(configGrpcInterfaceKey),
	}),
	altsrc.NewIntFlag(&cli.IntFlag{
		Name:     configGrpcPortKey,
		Value:    8080,
		Usage:    "Port for OSLC's gRPC server",
		EnvVars:  []string{configGrpcPortEnv},
		FilePath: configGrpcPortFile,
		Action:   cfgIntMustBeValidPort(configGrpcPortKey),
	}),
	altsrc.NewBoolFlag(&cli.BoolFlag{
		Name:     configMetricsEnabledKey,
		Value:    false,
		Usage:    fmt.Sprintf("Enable metrics server - If enabled, a Prometheus metrics server will be started on %s:%s and served via HTTP", configMetricsInterfaceKey, configMetricsPortKey),
		EnvVars:  []string{configMetricsEnabledEnv},
		FilePath: configMetricsEnabledFile,
	}),
	altsrc.NewStringFlag(&cli.StringFlag{
		Name:     configMetricsInterfaceKey,
		Value:    "0.0.0.0",
		Usage:    "Interface for OSLC's metrics server",
		EnvVars:  []string{configMetricsInterfaceEnv},
		FilePath: configMetricsInterfaceFile,
		Action:   cfgStringMustNotBeEmpty(configMetricsInterfaceKey),
	}),
	altsrc.NewIntFlag(&cli.IntFlag{
		Name:     configMetricsPortKey,
		Value:    9090,
		Usage:    "Port for OSLC's metrics server",
		EnvVars:  []string{configMetricsPortEnv},
		FilePath: configMetricsPortFile,
		Action:   cfgIntMustBeValidPort(configMetricsPortKey),
	}),
	altsrc.NewStringFlag(&cli.StringFlag{
		Name:     configLogLevelKey,
		Value:    "info",
		Usage:    "Log level for OSLC - valid values are debug, info, warn, error",
		EnvVars:  []string{configLogLevelEnv},
		FilePath: configLogLevelFile,
		Action:   cfgStringMustBeValidLoggingLevel(configLogLevelKey),
	}),
	altsrc.NewStringFlag(&cli.StringFlag{
		Name:     configLogKindKey,
		Value:    "json",
		Usage:    "Log kind for OSLC - valid values are text, json and discard. Setting the logger to discard will discard all logs",
		EnvVars:  []string{configLogKindEnv},
		FilePath: configLogKindFile,
		Action:   cfgStringMustBeValidLoggingKind(configLogKindKey),
	}),
	altsrc.NewStringFlag(&cli.StringFlag{
		Name:     configTlsCertFilePathKey,
		Value:    getFilePathWithPrefix("oslc_tls_cert_file"),
		Usage:    "Path to the TLS certificate file",
		EnvVars:  []string{configTlsCertFilePathEnv},
		FilePath: configTlsCertFilePathFile,
		Action:   cfgStringMustNotBeEmpty(configTlsCertFilePathKey),
	}),
	altsrc.NewStringFlag(&cli.StringFlag{
		Name:     configTlsKeyFilePathKey,
		Value:    getFilePathWithPrefix("oslc_tls_key_file"),
		Usage:    "Path to the TLS key file",
		EnvVars:  []string{configTlsKeyFilePathEnv},
		FilePath: configTlsKeyFilePathFile,
		Action:   cfgStringMustNotBeEmpty(configTlsKeyFilePathKey),
	}),
}

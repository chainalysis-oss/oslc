package main

import (
	"github.com/BoRuDar/configuration/v4"
	"github.com/go-playground/validator/v10"
	"strings"
)

type cfgDatastore struct {
	Username string `default:"postgres" env:"OSLC_DATASTORE_USER" file_json:"datastore.username" validate:"required"`
	Password string `default:"postgres" env:"OSLC_DATASTORE_PASSWORD" file_json:"datastore.password" validate:"required"`
	Host     string `default:"localhost" env:"OSLC_DATASTORE_HOST" file_json:"datastore.host" validate:"required"`
	Port     int    `default:"5432" env:"OSLC_DATASTORE_PORT" file_json:"datastore.port" validate:"min=1,max=65535"`
	Database string `default:"postgres" env:"OSLC_DATASTORE_DB" file_json:"datastore.database" validate:"required"`
}
type cfgGrpc struct {
	Interface string `default:"0.0.0.0" env:"OSLC_GRPC_INTERFACE" file_json:"grpc.interface" validate:"ip"`
	Port      int    `default:"8080" env:"OSLC_GRPC_PORT" file_json:"grpc.port" validate:"min=1,max=65535"`
	NoTLS     bool   `default:"false" env:"OSLC_GRPC_NO_TLS" file_json:"grpc.no_tls"`
}

type cfgMetrics struct {
	Enabled   bool   `default:"true" env:"OSLC_METRICS_ENABLED" file_json:"metrics.enabled"`
	Interface string `default:"0.0.0.0" env:"OSLC_METRICS_INTERFACE" file_json:"metrics.interface" validate:"ip"`
	Port      int    `default:"9090" env:"OSLC_METRICS_PORT" file_json:"metrics.port" validate:"min=1,max=65535"`
}
type cfgLog struct {
	Level string `default:"info" env:"OSLC_LOG_LEVEL" file_json:"log.level" validate:"oneof=debug info warn error"`
	Kind  string `default:"json" env:"OSLC_LOG_KIND" file_json:"log.kind" validate:"oneof=text json"`
}

type cfg struct {
	Datastore cfgDatastore `validate:"required"`
	Grpc      cfgGrpc      `validate:"required"`
	Metrics   cfgMetrics   `validate:"required"`
	Log       cfgLog       `validate:"required"`
	TLS       cfgTLS       `validate:"required"`
}

type cfgTLS struct {
	CertFile string `default:"/run/secrets/oslc_tls_cert_file" env:"OSLC_TLS_CERT_FILE" file_json:"tls.cert_file"`
	KeyFile  string `default:"/run/secrets/oslc_tls_key_file" env:"OSLC_TLS_KEY_FILE" file_json:"tls.key_file"`
}

func createConfiguration(filepath string) (*cfg, error) {
	cfgObj := cfg{}
	err := configuration.New(
		&cfgObj,
		configuration.NewFlagProvider(),
		configuration.NewEnvProvider(),
		configuration.NewJSONFileProvider(filepath),
		configuration.NewDefaultProvider(),
	).InitValues()

	if err != nil {
		if strings.Contains(err.Error(), "JSONFileProvider") && strings.Contains(err.Error(), "no such file or directory") {
			err = configuration.New(
				&cfgObj,
				configuration.NewFlagProvider(),
				configuration.NewEnvProvider(),
				configuration.NewDefaultProvider(),
			).InitValues()
		}
		if err != nil {
			return nil, err
		}
	}

	return &cfgObj, validateConfig(&cfgObj)
}

func validateConfig(c *cfg) error {
	validate := validator.New(validator.WithRequiredStructEnabled())
	return validate.Struct(c)
}

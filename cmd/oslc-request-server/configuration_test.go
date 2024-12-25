package main

import (
	"github.com/BoRuDar/configuration/v4"
	"github.com/stretchr/testify/require"
	"reflect"
	"testing"
)

var defaultCfg cfg = cfg{
	Grpc: cfgGrpc{
		Interface: "0.0.0.0",
		Port:      8080,
	},
	Datastore: cfgDatastore{
		Host:     "localhost",
		Port:     5432,
		Database: "postgres",
		Username: "postgres",
		Password: "postgres",
	},
	Metrics: cfgMetrics{
		Enabled:   true,
		Interface: "0.0.0.0",
		Port:      9090,
	},
	Log: cfgLog{
		Level: "info",
		Kind:  "json",
	},
	TLS: cfgTLS{
		CertFile: "/run/secrets/oslc_tls_cert_file",
		KeyFile:  "/run/secrets/oslc_tls_key_file",
	},
}

// This test is to ensure that the configuration derived using the default provider is correct. This should catch
// instances where default values are changed in the [cfg] struct, but tests haven't been updated. We only use the
// [defaultCfg] struct for testing, because it prevents us from having to type out the default config for every test.
func Test_defaultCfgIsCorrect(t *testing.T) {
	cfgObj := cfg{}
	err := configuration.New(
		&cfgObj,
		configuration.NewDefaultProvider(),
	).InitValues()

	require.NoError(t, err)

	require.Equal(t, defaultCfg, cfgObj)
}

func Test_validateConfig(t *testing.T) {
	cfg := cfg{}
	require.Error(t, validateConfig(&cfg))

	cfg = defaultCfg
	require.NoError(t, validateConfig(&cfg))
}

func Test_createConfiguration(t *testing.T) {
	type args struct {
		filepath string
	}
	tests := []struct {
		name    string
		args    args
		want    *cfg
		wantErr bool
	}{
		{
			name: "no_config_file",
			args: args{
				filepath: "",
			},
			want:    &defaultCfg,
			wantErr: false,
		},
		{
			name: "config_file_present",
			args: args{
				filepath: "./testdata/config.json",
			},
			want: func() *cfg {
				c := defaultCfg
				c.Log.Kind = "text"
				return &c
			}(),
			wantErr: false,
		},
		{
			name: "bad_config",
			args: args{
				filepath: "./testdata/unparseable_config_invalid_json.json",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := createConfiguration(tt.args.filepath)
			if (err != nil) != tt.wantErr {
				t.Errorf("createConfiguration() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("createConfiguration() got = %v, want %v", got, tt.want)
			}
		})
	}
}

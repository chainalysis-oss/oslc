package main

import (
	"fmt"
	"github.com/stretchr/testify/require"
	"testing"
)

func Test_cfgLog_Level_invalid(t *testing.T) {
	c := defaultCfg
	c.Log.Level = "invalid"
	require.Error(t, validateConfig(&c))
}

func Test_cfgLog_Level(t *testing.T) {
	cases := []string{
		"debug",
		"info",
		"warn",
		"error",
	}
	for _, c := range cases {
		t.Run(c, func(t *testing.T) {
			cfg := defaultCfg
			cfg.Log.Level = c
			require.NoError(t, validateConfig(&cfg))
		})
	}
}

func Test_cfgLog_Kind_invalid(t *testing.T) {
	c := defaultCfg
	c.Log.Kind = "invalid"
	require.Error(t, validateConfig(&c))
}

func Test_cfgLog_Kind(t *testing.T) {
	cases := []string{
		"text",
		"json",
	}
	for _, c := range cases {
		t.Run(c, func(t *testing.T) {
			cfg := defaultCfg
			cfg.Log.Kind = c
			require.NoError(t, validateConfig(&cfg))
		})
	}
}

func Test_cfgGrpc_Interface_invalid(t *testing.T) {
	cases := []string{
		"invalid",
		"127.0.0.1.1",
		"localhost",
	}
	for _, c := range cases {
		t.Run(c, func(t *testing.T) {
			cfg := defaultCfg
			cfg.Grpc.Interface = c
			require.Error(t, validateConfig(&cfg))
		})
	}
}

func Test_cfgGrpc_Interface(t *testing.T) {
	cases := []string{
		"0.0.0.0",
		"127.0.0.1",
	}
	for _, c := range cases {
		t.Run(c, func(t *testing.T) {
			cfg := defaultCfg
			cfg.Grpc.Interface = c
			require.NoError(t, validateConfig(&cfg))
		})
	}
}

func Test_cfgGrpc_Port_invalid(t *testing.T) {
	cases := []int{
		-1,
		0,
		65536,
	}
	for _, c := range cases {
		t.Run(fmt.Sprintf("%d", c), func(t *testing.T) {
			cfg := defaultCfg
			cfg.Grpc.Port = c
			require.Error(t, validateConfig(&cfg))
		})
	}
}

func Test_cfgGrpc_Port(t *testing.T) {
	cases := []int{
		1,
		65535,
	}
	for _, c := range cases {
		t.Run(fmt.Sprintf("%d", c), func(t *testing.T) {
			cfg := defaultCfg
			cfg.Grpc.Port = c
			require.NoError(t, validateConfig(&cfg))
		})
	}
}

func Test_cfgDatastore_Database_invalid(t *testing.T) {
	cfg := defaultCfg
	cfg.Datastore.Database = ""
	require.Error(t, validateConfig(&cfg))
}

func Test_cfgDatastore_Database(t *testing.T) {
	cfg := defaultCfg
	cfg.Datastore.Database = "test"
	require.NoError(t, validateConfig(&cfg))
}

func Test_cfgDatastore_Port_invalid(t *testing.T) {
	cases := []int{
		-1,
		0,
		65536,
	}
	for _, c := range cases {
		t.Run(fmt.Sprintf("%d", c), func(t *testing.T) {
			cfg := defaultCfg
			cfg.Datastore.Port = c
			require.Error(t, validateConfig(&cfg))
		})
	}
}

func Test_cfgDatastore_Port(t *testing.T) {
	cases := []int{
		1,
		65535,
	}
	for _, c := range cases {
		t.Run(fmt.Sprintf("%d", c), func(t *testing.T) {
			cfg := defaultCfg
			cfg.Datastore.Port = c
			require.NoError(t, validateConfig(&cfg))
		})
	}
}

func Test_cfgDatastore_Host_invalid(t *testing.T) {
	cfg := defaultCfg
	cfg.Datastore.Host = ""
	require.Error(t, validateConfig(&cfg))
}

func Test_cfgDatastore_Host(t *testing.T) {
	cfg := defaultCfg
	cfg.Datastore.Host = "localhost"
	require.NoError(t, validateConfig(&cfg))
}

func Test_cfgDatastore_Password_invalid(t *testing.T) {
	cfg := defaultCfg
	cfg.Datastore.Password = ""
	require.Error(t, validateConfig(&cfg))
}

func Test_cfgDatastore_Password(t *testing.T) {
	cfg := defaultCfg
	cfg.Datastore.Password = "test"
	require.NoError(t, validateConfig(&cfg))
}

func Test_cfgDatastore_Username_invalid(t *testing.T) {
	cfg := defaultCfg
	cfg.Datastore.Username = ""
	require.Error(t, validateConfig(&cfg))
}

func Test_cfgDatastore_Username(t *testing.T) {
	cfg := defaultCfg
	cfg.Datastore.Username = "test"
	require.NoError(t, validateConfig(&cfg))
}

func Test_cfgMetrics_Interface_invalid(t *testing.T) {
	cases := []string{
		"invalid",
	}
	for _, c := range cases {
		t.Run(c, func(t *testing.T) {
			cfg := defaultCfg
			cfg.Metrics.Interface = c
			require.Error(t, validateConfig(&cfg))
		})
	}
}

func Test_cfgMetrics_Interface(t *testing.T) {
	cases := []string{
		"127.0.0.1",
	}
	for _, c := range cases {
		t.Run(c, func(t *testing.T) {
			cfg := defaultCfg
			cfg.Metrics.Interface = c
			require.NoError(t, validateConfig(&cfg))
		})
	}
}

func Test_cfgMetrics_Port_invalid(t *testing.T) {
	cases := []int{
		-1,
		0,
		65536,
	}
	for _, c := range cases {
		t.Run(fmt.Sprintf("%d", c), func(t *testing.T) {
			cfg := defaultCfg
			cfg.Metrics.Port = c
			require.Error(t, validateConfig(&cfg))
		})
	}
}

func Test_cfgMetrics_Port(t *testing.T) {
	cases := []int{
		1,
		65535,
	}
	for _, c := range cases {
		t.Run(fmt.Sprintf("%d", c), func(t *testing.T) {
			cfg := defaultCfg
			cfg.Metrics.Port = c
			require.NoError(t, validateConfig(&cfg))
		})
	}
}

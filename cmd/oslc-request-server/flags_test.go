package main

import (
	"flag"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v2"
	"strconv"
	"testing"
)

func TestGetFilePathWithPrefix(t *testing.T) {
	tests := []struct {
		name string
		path string
		env  string
		want string
	}{
		{
			name: "empty path",
			path: "",
			want: filePrefixFallback,
		},
		{
			name: "none-empty path",
			path: "/path/to/file",
			want: filePrefixFallback + "/path/to/file",
		},
		{
			name: "empty path with env",
			path: "",
			env:  "/env/path",
			want: "/env/path",
		},
		{
			name: "none-empty path with env",
			path: "/path/to/file",
			env:  "/env/path",
			want: "/env/path/path/to/file",
		},
		{
			name: "empty path with env and trailing slash",
			path: "",
			env:  "/env/path/",
			want: "/env/path",
		},
		{
			name: "none-empty path with env and trailing slash",
			path: "/path/to/file",
			env:  "/env/path/",
			want: "/env/path/path/to/file",
		},
		{
			name: "path with env and leading slash",
			path: "/path/to/file",
			env:  "/env/path/",
			want: "/env/path/path/to/file",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.env != "" {
				t.Setenv(filePrefixEnv, tt.env)
				defer t.Setenv(filePrefixEnv, "")
			}
			require.Equal(t, tt.want, getFilePathWithPrefix(tt.path))
		})
	}
}

func TestGetEnv(t *testing.T) {
	t.Setenv("KEY_EXISTS", "value")
	cases := []struct {
		name     string
		key      string
		fallback string
		want     string
	}{
		{
			name:     "key exists",
			key:      "KEY_EXISTS",
			fallback: "fallback",
			want:     "value",
		},
		{
			name:     "key does not exist",
			key:      "KEY_DOES_NOT_EXIST",
			fallback: "fallback",
			want:     "fallback",
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.want, getEnv(tt.key, tt.fallback))
		})
	}
}

func createContextWithStringFlag(t *testing.T, name string, value string) *cli.Context {
	t.Helper()
	fs := flag.NewFlagSet("", flag.ExitOnError)
	app := cli.NewApp()
	fs.String(name, "default", "")
	err := fs.Set(name, value)
	require.NoError(t, err)

	return cli.NewContext(app, fs, nil)
}

func createContextWithIntFlag(t *testing.T, name string, value int) *cli.Context {
	t.Helper()
	fs := flag.NewFlagSet("", flag.ExitOnError)
	app := cli.NewApp()
	fs.Int(name, 0, "")
	err := fs.Set(name, strconv.Itoa(value))
	require.NoError(t, err)

	return cli.NewContext(app, fs, nil)
}

func TestCfgStringMustNotBeEmpty(t *testing.T) {
	cCtx := createContextWithStringFlag(t, "key", "")
	err := cfgStringMustNotBeEmpty("key")(cCtx, "")
	var cfgValErr *configValidationError
	require.ErrorAs(t, err, &cfgValErr)

	cCtx = createContextWithStringFlag(t, "key", "value")
	err = cfgStringMustNotBeEmpty("key")(cCtx, "value")
	require.NoError(t, err)
}

func TestCfgIntMustBeValidPort(t *testing.T) {
	cCtx := createContextWithIntFlag(t, "key", 0)
	err := cfgIntMustBeValidPort("key")(cCtx, 0)
	var cfgValErr *configValidationError
	require.ErrorAs(t, err, &cfgValErr)

	cCtx = createContextWithIntFlag(t, "key", 65536)
	err = cfgIntMustBeValidPort("key")(cCtx, 65536)
	require.Error(t, err)

	cCtx = createContextWithIntFlag(t, "key", 65535)
	err = cfgIntMustBeValidPort("key")(cCtx, 65535)
	require.NoError(t, err)
}

func TestCfgStringMustBeValidLoggingLevel(t *testing.T) {
	cCtx := createContextWithStringFlag(t, "key", "invalid")
	err := cfgStringMustBeValidLoggingLevel("key")(cCtx, "invalid")
	var cfgValErr *configValidationError
	require.ErrorAs(t, err, &cfgValErr)

	cases := []struct {
		value string
	}{
		{"debug"},
		{"info"},
		{"warn"},
		{"error"},
	}

	for _, tt := range cases {
		t.Run(tt.value, func(t *testing.T) {
			cCtx = createContextWithStringFlag(t, "key", tt.value)
			err = cfgStringMustBeValidLoggingLevel("key")(cCtx, tt.value)
			require.NoError(t, err)
		})
	}
}

func TestCfgStringMustBeValidLoggingKind(t *testing.T) {
	cCtx := createContextWithStringFlag(t, "key", "invalid")
	err := cfgStringMustBeValidLoggingKind("key")(cCtx, "invalid")
	var cfgValErr *configValidationError
	require.ErrorAs(t, err, &cfgValErr)

	cases := []struct {
		value string
	}{
		{"text"},
		{"json"},
	}

	for _, tt := range cases {
		t.Run(tt.value, func(t *testing.T) {
			cCtx = createContextWithStringFlag(t, "key", tt.value)
			err = cfgStringMustBeValidLoggingKind("key")(cCtx, tt.value)
			require.NoError(t, err)
		})
	}
}

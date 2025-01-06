package oslc

import (
	"context"
	"github.com/chainalysis-oss/oslc"
	"github.com/chainalysis-oss/oslc/maven"
	oslcmocks "github.com/chainalysis-oss/oslc/mocks/oslc"
	"github.com/chainalysis-oss/oslc/npm"
	"github.com/chainalysis-oss/oslc/pypi"
	"github.com/stretchr/testify/require"
	"io"
	"log/slog"
	"os"
	"testing"
)

func TestNewServer(t *testing.T) {
	server, err := NewServer(WithLogger(slog.New(slog.NewTextHandler(io.Discard, nil))))
	require.NoError(t, err)
	require.NotNil(t, server)
	require.Equal(t, slog.New(slog.NewTextHandler(io.Discard, nil)), server.options.Logger)
}

func TestNewServer_globalOptionsAreApplied(t *testing.T) {
	optCopy := make([]ServerOption, len(globalServerOptions))
	copy(optCopy, globalServerOptions)
	defer func() {
		globalServerOptions = optCopy
	}()

	globalServerOptions = append(globalServerOptions, WithLogger(slog.New(slog.NewTextHandler(io.Discard, nil))))
	server, err := NewServer()
	require.NoError(t, err)
	require.Equal(t, slog.New(slog.NewTextHandler(io.Discard, nil)), server.options.Logger)
}

func TestFuncClientOption_apply(t *testing.T) {
	opts := serverOptions{}
	fdo := newFuncClientOption(func(o *serverOptions) {
		o.Logger = slog.New(slog.NewTextHandler(io.Discard, nil))
	})
	fdo.apply(&opts)
	require.Equal(t, slog.New(slog.NewTextHandler(io.Discard, nil)), opts.Logger)
}

func TestNewFuncClientOption(t *testing.T) {
	fdo := newFuncClientOption(func(o *serverOptions) {
		o.Logger = slog.New(slog.NewTextHandler(io.Discard, nil))
	})
	require.NotNil(t, fdo)
}

func TestWithLogger(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	require.NotNil(t, logger)
	opts := serverOptions{}
	f := WithLogger(logger)
	f.apply(&opts)
	require.Equal(t, logger, opts.Logger)
}

func TestWithPypiClient(t *testing.T) {
	client, err := pypi.NewClient()
	require.NoError(t, err)
	require.NotNil(t, client)
	opts := serverOptions{}
	f := WithPypiClient(client)
	f.apply(&opts)
	require.Equal(t, client, opts.PypiClient)
}

func TestWithNpmClient(t *testing.T) {
	client, err := npm.NewClient()
	require.NoError(t, err)
	require.NotNil(t, client)
	opts := serverOptions{}
	f := WithNpmClient(client)
	f.apply(&opts)
	require.Equal(t, client, opts.NpmClient)
}

func TestWithMavenClient(t *testing.T) {
	client, err := maven.NewClient()
	require.NoError(t, err)
	require.NotNil(t, client)
	opts := serverOptions{}
	f := WithMavenClient(client)
	f.apply(&opts)
	require.Equal(t, client, opts.MavenClient)
}

type mockDatastore struct{}

func (m mockDatastore) Save(ctx context.Context, entry oslc.Entry) error {
	return nil
}

func (m mockDatastore) Retrieve(ctx context.Context, name, version, distributor string) (oslc.Entry, error) {
	return oslc.Entry{}, nil
}

func TestWithDatastore(t *testing.T) {
	ds := mockDatastore{}
	opts := serverOptions{}
	f := WithDatastore(ds)
	f.apply(&opts)
	require.Equal(t, ds, opts.Datastore)
}

func TestWithLicenseIDNormalizer(t *testing.T) {
	mock := oslcmocks.NewMockLicenseIDNormalizer(t)
	opts := serverOptions{}
	f := WithLicenseIDNormalizer(mock)
	f.apply(&opts)
	require.Equal(t, mock, opts.LicenseIDNormalizer)
}

func TestWithCratesIoClient(t *testing.T) {
	client := oslcmocks.NewMockDistributorClient(t)
	opts := serverOptions{}
	f := WithCratesIoClient(client)
	f.apply(&opts)
	require.Equal(t, client, opts.CratesIoClient)
}

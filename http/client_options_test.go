package http

import (
	"github.com/stretchr/testify/require"
	"log/slog"
	"net/http"
	"os"
	"testing"
)

func TestNewClient(t *testing.T) {
	client, err := NewClient(WithLogger(slog.Default()))
	require.NoError(t, err)
	require.NotNil(t, client)
	require.Equal(t, slog.Default(), client.options.Logger)
	require.NotNil(t, client.options.HttpClient)
}

func TestNewClient_globalOptionsAreApplied(t *testing.T) {
	optCopy := make([]ClientOption, len(globalClientOptions))
	copy(optCopy, globalClientOptions)
	defer func() {
		globalClientOptions = optCopy
	}()

	globalClientOptions = append(globalClientOptions, WithLogger(slog.Default()))
	client, err := NewClient()
	require.NoError(t, err)
	require.Equal(t, slog.Default(), client.options.Logger)
}

func TestFuncClientOption_apply(t *testing.T) {
	opts := clientOptions{}
	fdo := newFuncClientOption(func(o *clientOptions) {
		o.Logger = slog.Default()
	})
	fdo.apply(&opts)
	require.Equal(t, slog.Default(), opts.Logger)
}

func TestNewFuncClientOption(t *testing.T) {
	fdo := newFuncClientOption(func(o *clientOptions) {
		o.Logger = slog.Default()
	})
	require.NotNil(t, fdo)
}

func TestWithLogger(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	require.NotNil(t, logger)
	opts := clientOptions{}
	f := WithLogger(logger)
	f.apply(&opts)
	require.Equal(t, logger, opts.Logger)
}

func TestWithHTTPClient(t *testing.T) {
	client := &http.Client{}
	opts := clientOptions{}
	f := WithHTTPClient(client)
	f.apply(&opts)
	require.Equal(t, client, opts.HttpClient)
}

func TestWithUserAgent(t *testing.T) {
	opts := clientOptions{}
	f := WithUserAgent("test")
	f.apply(&opts)
	require.Equal(t, "test", opts.UserAgent)
}

func TestWithHeaders(t *testing.T) {
	opts := clientOptions{}
	f := WithHeaders(http.Header{
		"Accept": {"test"},
	})
	f.apply(&opts)
	require.Equal(t, "test", opts.Headers.Get("Accept"))
}

func TestWithReaderLimit(t *testing.T) {
	opts := clientOptions{}
	f := WithReaderLimit(10)
	f.apply(&opts)
	require.Equal(t, int64(10), opts.ReaderLimit)
}

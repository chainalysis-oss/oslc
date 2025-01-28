package goproxy

import (
	"github.com/chainalysis-oss/oslc/http"
	"log/slog"
	"os"
)

type clientOptions struct {
	HttpClient *http.Client
	BaseURL    string
	Logger     *slog.Logger
	TempDir    string
}

var defaultClientOptions = clientOptions{
	BaseURL: goProxyBaseURL,
	Logger:  slog.Default(),
	TempDir: os.TempDir(),
}

var GlobalClientOptions []ClientOption

// ClientOption is an option for configuring a Client.
type ClientOption interface {
	apply(*clientOptions)
}

// funcClientOption is a ClientOption that calls a function.
// It is used to wrap a function, so it satisfies the ClientOption interface.
type funcClientOption struct {
	f func(*clientOptions)
}

func (fdo *funcClientOption) apply(opts *clientOptions) {
	fdo.f(opts)
}

func newFuncClientOption(f func(*clientOptions)) *funcClientOption {
	return &funcClientOption{
		f: f,
	}
}

// WithHTTPClient returns a ClientOption that uses the provided http.Client.
func WithHTTPClient(c *http.Client) ClientOption {
	return newFuncClientOption(func(opts *clientOptions) {
		opts.HttpClient = c
	})
}

// WithLogger returns a ClientOption that uses the provided logger.
func WithLogger(logger *slog.Logger) ClientOption {
	return newFuncClientOption(func(opts *clientOptions) {
		opts.Logger = logger
	})
}

// WithTempDir returns a ClientOption that uses the provided temp dir.
func WithTempDir(tempDir string) ClientOption {
	return newFuncClientOption(func(opts *clientOptions) {
		opts.TempDir = tempDir
	})
}

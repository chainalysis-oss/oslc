package pypi

import (
	"github.com/chainalysis-oss/oslc/http"
	"log/slog"
)

type clientOptions struct {
	HttpClient *http.Client
	BaseURL    string
	UserAgent  string
	Logger     *slog.Logger
}

var defaultClientOptions = clientOptions{
	BaseURL: "https://pypi.org",
	Logger:  slog.Default(),
}

var globalClientOptions []ClientOption

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

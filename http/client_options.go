package http

import (
	"log/slog"
	"net/http"
	"time"
)

type clientOptions struct {
	Logger     *slog.Logger
	HttpClient *http.Client
	UserAgent  string
	Headers    http.Header
	// ReaderLimit is the maximum number of bytes to read from the response body.
	ReaderLimit int64
}

var defaultClientOptions = clientOptions{
	Logger: slog.Default(),
	HttpClient: &http.Client{
		Timeout: 10 * time.Second,
	},
	UserAgent:   "Open Software License Catalogue (github.com/chainalysis-oss/oslc)",
	ReaderLimit: 20 * 1024 * 1024,
}

var globalClientOptions []ClientOption

type ClientOption interface {
	apply(*clientOptions)
}

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

func WithLogger(logger *slog.Logger) ClientOption {
	return newFuncClientOption(func(opts *clientOptions) {
		opts.Logger = logger
	})
}

func WithHTTPClient(c *http.Client) ClientOption {
	return newFuncClientOption(func(opts *clientOptions) {
		opts.HttpClient = c
	})
}

func WithUserAgent(ua string) ClientOption {
	return newFuncClientOption(func(opts *clientOptions) {
		opts.UserAgent = ua
	})
}

func WithHeaders(headers http.Header) ClientOption {
	return newFuncClientOption(func(opts *clientOptions) {
		opts.Headers = headers
	})
}

func WithReaderLimit(limit int64) ClientOption {
	return newFuncClientOption(func(opts *clientOptions) {
		opts.ReaderLimit = limit
	})
}

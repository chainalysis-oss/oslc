package http

import (
	"bytes"
	"context"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"
)

var DefaultClient *Client

func init() {
	DefaultClient, _ = NewClient(
		WithLogger(slog.Default()),
		WithUserAgent("oslc-go"),
		WithHTTPClient(&http.Client{
			Timeout: 10 * time.Second,
		}),
	)
}

type Client struct {
	options *clientOptions
}

func NewClient(options ...ClientOption) (*Client, error) {
	opts := defaultClientOptions
	for _, opt := range globalClientOptions {
		opt.apply(&opts)
	}
	for _, opt := range options {
		opt.apply(&opts)
	}

	if opts.Headers == nil {
		opts.Headers = make(http.Header)
	}

	return &Client{
		options: &opts,
	}, nil
}

// Query executes a GET request against the given url and returns the response and any errors associated with it.
//
// The HTTP Headers defined under the [clientOptions] struct are added to the request. If the User-Agent header is
// not set, it is set to the value of the UserAgent field in the [clientOptions] struct.
//
// The response body is limited to the value of the ReaderLimit field in the [clientOptions] struct and an error
// is returned if the limit is exceeded. Additionally, the response body will be read by this function to facilitate
// logging, yet returned to the caller as a ReadCloser to be handled like any other response body.
func (c *Client) Query(url string) (*http.Response, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header = c.options.Headers
	if req.Header.Get("User-Agent") == "" {
		req.Header.Set("User-Agent", c.options.UserAgent)
	}

	logHeader := make([]any, 0)
	for header := range req.Header {
		logHeader = append(logHeader, slog.String(strings.ToLower(header), req.Header.Get(header)))
	}
	c.options.Logger.LogAttrs(req.Context(), slog.LevelDebug, "outgoing request", slog.String("path", req.Method), slog.String("url", req.URL.String()), slog.Group("headers", logHeader...))

	resp, err := c.options.HttpClient.Do(req)
	if err != nil {
		return nil, err
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, c.options.ReaderLimit))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Reset the body so it can be read again. The body is a ReadCloser, but [bytes.NewBuffer] does not implement
	// ReadCloser, so we need to use [io.NopCloser] to wrap it.
	resp.Body = io.NopCloser(bytes.NewBuffer(body))
	c.options.Logger.LogAttrs(context.Background(), slog.LevelDebug, "response", slog.Int("status", resp.StatusCode), slog.String("body", string(body)))
	return resp, err
}

type RoundTripFunc func(req *http.Request) (*http.Response, error)

func (f RoundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

// NewTestHTTPClient returns a new [http.Client] that uses the given function as its RoundTripper.
// This allows for easy mocking of an HTTP client for test purposes. Usually, this function is used in conjunction
// with the [WithHTTPClient] function to configure a [Client] with a mock HTTP client.
func NewTestHTTPClient(fn RoundTripFunc) *http.Client {
	return &http.Client{
		Transport: fn,
	}
}

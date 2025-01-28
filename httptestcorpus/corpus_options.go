package httptestcorpus

import (
	"net/http"
	"path"
	"testing"
)

type clientOptions struct {
	t          *testing.T
	corpusDir  string
	httpClient *http.Client
}

var defaultClientOptions = clientOptions{
	corpusDir:  path.Join("testdata", "httptestcorpus"),
	httpClient: http.DefaultClient,
}

// ClientOption is an option for configuring a [Client].
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

// WithTest returns a ClientOption that instructs the client to use the provided testing.T. This is required for the
// client to function correctly.
func WithTest(t *testing.T) ClientOption {
	return newFuncClientOption(func(opts *clientOptions) {
		opts.t = t
	})
}

// WithCorpusDir returns a ClientOption that, when provided to a [Client], instructs the client to use the provided
// directory as the corpus directory. If not provided, the default directory is "testdata/httptestcorpus".
func WithCorpusDir(dir string) ClientOption {
	return newFuncClientOption(func(opts *clientOptions) {
		opts.corpusDir = dir
	})
}

// WithHTTPClient returns a ClientOption that, when provided to a [Client], instructs the client to use the provided
// [http.Client] for making requests. If not provided, the default client is [http.DefaultClient].
func WithHTTPClient(client *http.Client) ClientOption {
	return newFuncClientOption(func(opts *clientOptions) {
		opts.httpClient = client
	})
}

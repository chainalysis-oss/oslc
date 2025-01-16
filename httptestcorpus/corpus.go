// Package httptestcorpus makes testing HTTP calls easier by caching responses in a corpus directory.
//
// Usage example:
//
//	 package awesomeproject
//
//   import (
//	     "net/http"
//	     "testing"
//	     "github.com/chainalysis-oss/oslc/httptestcorpus"
//   )
//
//   func TestAwesomeProject(t *testing.T) {
//	     httpClient := httptestcorpus.Embed(&http.Client{}, httptestcorpus.WithTest(t))
//	     resp, err := httpClient.Get("https://example.com")
//	     if err != nil {
//	       t.Fatal(err)
//	     }
//	     if resp.StatusCode != 200 {
//	       t.Fatalf("unexpected status code: %d", resp.StatusCode)
//	     }
//   }
//
// You will notice that on the initial run, the response will be fetched from the server and a directory will be created
// next to your test file. The directory will be called `testdata/httptestcorpus` and will contain a file with the name
// of the test and a fingerprint of the request. The file will contain the response body, status code, and headers.
//
// Subsequent runs will use the cached response instead of making the request to the server.

package httptestcorpus

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
)

var (
	ErrNoTestProvided = errors.New("no test provided when creating a new client - did you forget to use WithTest?")
)

// Client is a client that can make HTTP requests via its [Client.do] method. For every request processed, the client
// will fingerprint the request (see [RequestFingerprint]) and if the clients corpus contains a response for that
// fingerprint, it will return the cached response. If the response is not in the corpus, the client will make the
// request and install the response in the corpus before returning it.
//
// The corpus directory defaults to `testdata/httptestcorpus` and can be changed with the [WithCorpusDir] option. If the
// corpus directory path is relative, it will be relative to the current working directory.
//
// The client requires a testing.T instance to be provided via the [WithTest] option. This is used to name the corpus
// entries and to ensure that the client is used in the context of a test.
//
// The client should not be created directly, but instead via the [NewClient] function.
type Client struct {
	options *clientOptions
}

// NewClient creates a new client with the provided options. The client will return an error if [WithTest] is not
// provided.
//
// Default options are:
// - [WithCorpusDir] with the default corpus directory path `testdata/httptestcorpus`
func NewClient(options ...ClientOption) (*Client, error) {
	opts := defaultClientOptions
	for _, opt := range options {
		opt.apply(&opts)
	}

	if opts.t == nil {
		return nil, ErrNoTestProvided
	}

	return &Client{
		options: &opts,
	}, nil
}

// RoundTrip implements the http.RoundTripper interface.
// This method can be used to create a custom http.Client with the client as the transport.
func (c *Client) RoundTrip(req *http.Request) (*http.Response, error) {
	return c.do(req)
}

func (c *Client) do(req *http.Request) (*http.Response, error) {
	fp, err := NewRequestFingerprint(req)
	if err != nil {
		return nil, err
	}
	corpusEntry := path.Join(c.options.corpusDir, c.options.t.Name()+"."+fp.Fingerprint()+".json")
	_, err = os.Stat(corpusEntry)
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return nil, err
		}
	}

	if errors.Is(err, os.ErrNotExist) {
		resp, err := c.options.httpClient.Do(req)
		if err != nil {
			return nil, err
		}

		err = writeResponseToCache(resp, corpusEntry)
		if err != nil {
			return nil, err
		}

		return resp, err
	}

	rec, err := readResponseFromCache(corpusEntry)
	if err != nil {
		return nil, err
	}
	return &http.Response{
		StatusCode: rec.StatusCode,
		Body:       io.NopCloser(bytes.NewReader(rec.Body)),
		Header:     rec.Header,
	}, nil
}

func writeResponseToCache(resp *http.Response, p string) error {
	err := os.MkdirAll(path.Dir(p), 0755)
	if err != nil {
		if !errors.Is(err, os.ErrExist) {
			return err
		}
	}
	f, err := os.Create(p)
	if err != nil {
		return err
	}
	defer f.Close()
	rec, err := recordResponse(resp)
	if err != nil {
		return err
	}

	err = json.NewEncoder(f).Encode(rec)
	if err != nil {
		return err
	}

	return nil
}

func readResponseFromCache(path string) (recordedResponse, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return recordedResponse{}, err
	}
	var rec recordedResponse
	err = json.Unmarshal(b, &rec)
	if err != nil {
		return recordedResponse{}, err
	}
	return rec, nil
}

type recordedResponse struct {
	StatusCode int
	Body       []byte
	Header     map[string][]string
}

func recordResponse(resp *http.Response) (recordedResponse, error) {
	if resp == nil {
		return recordedResponse{}, nil
	}

	b, err := io.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		return recordedResponse{}, err
	}
	// Reset the body so it can be read again.
	resp.Body = io.NopCloser(bytes.NewBuffer(b))
	return recordedResponse{
		StatusCode: resp.StatusCode,
		Body:       b,
		Header:     resp.Header,
	}, nil
}

type RequestFingerprint struct {
	Method string
	URL    string
	Body   []byte
	Header map[string][]string
}

// NewRequestFingerprint creates a new request fingerprint from the provided request.
//
// If the request body is not nil, it will be read and stored in the fingerprint.
func NewRequestFingerprint(req *http.Request) (RequestFingerprint, error) {
	var b []byte
	var err error
	if req.Body != nil {
		b, err = io.ReadAll(req.Body)
		if err != nil {
			return RequestFingerprint{}, err
		}
	}
	return RequestFingerprint{
		Method: req.Method,
		URL:    req.URL.String(),
		Body:   b,
		Header: req.Header,
	}, nil
}

// Fingerprint returns the fingerprint of the request.
// The fingerprint is a SHA256 hash of the JSON representation of the [RequestFingerprint] struct.
func (rf RequestFingerprint) Fingerprint() string {
	d := bytes.Buffer{}
	_ = json.NewEncoder(&d).Encode(rf)
	sum := sha256.Sum256(d.Bytes())
	return fmt.Sprintf("%x", sum)
}

// Embed installs a [Client] as the transport of the provided http.Client. The [Client] will be created using the
// [NewClient] function with the provided options. This is a shorthand for creating a new client and setting it as the
// transport of a http.Client.
//
// If an error occurs during the creation of the client, and options contain a [testing.T] instance, the error will be
// reported using the [testing.T.Error] method and nil is returned. If no [testing.T] instance is provided, the error will
// cause a panic.
//
// The returned [http.Client] is the same as the one provided as an argument. Either can be used after the call.
func Embed(httpClient *http.Client, options ...ClientOption) *http.Client {
	opts := clientOptions{}
	for _, opt := range options {
		opt.apply(&opts)
	}

	c, err := NewClient(options...)
	if err != nil {
		if opts.t != nil {
			opts.t.Error(err)
			return nil
		}
		panic(err)
	}
	httpClient.Transport = c
	return httpClient
}

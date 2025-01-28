package httptestcorpus

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path"
	"strings"
	"testing"
)

type testHttpServer struct {
	httpTestServer *httptest.Server
	ReqCounter     int
}

func (t *testHttpServer) Close() {
	t.httpTestServer.Close()
}

func (t *testHttpServer) URL() string {
	return t.httpTestServer.URL
}

func newTestServer(statuscode int, body []byte, header map[string]string) *testHttpServer {
	srv := testHttpServer{}
	wrapCalled := func(srv *testHttpServer) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			srv.ReqCounter++
			for k, v := range header {
				w.Header().Set(k, v)
			}
			w.WriteHeader(statuscode)
			w.Write(body)
			return
		}
	}
	srv.httpTestServer = httptest.NewServer(wrapCalled(&srv))

	return &srv
}

func TestClient(t *testing.T) {
	corpusDir := t.TempDir()
	ts := newTestServer(
		200,
		[]byte(`{"Version":"v0.3.0","Time":"2025-01-06T15:15:29Z","Origin":{"VCS":"git","URL":"https://github.com/chainalysis-oss/oslc","Hash":"755c6565c94d5ff6fd1fbaab923e36be424360f0","Ref":"refs/tags/v0.3.0"}}`),
		map[string]string{
			"Accept-Ranges":                       "bytes",
			"Access-Control-Allow-Origin":         "*",
			"Age":                                 "2859",
			"Alt-Svc":                             "h3=\":443\"; ma=2592000,h3-29=\":443\"; ma=2592000",
			"Cache-Control":                       "public, max-age=10800",
			"Content-Length":                      "196",
			"Content-Security-Policy-Report-Only": "script-src 'none'; form-action 'none'; frame-src 'none'; report-uri https://csp.withgoogle.com/csp/goa-fa2dfb7c_2",
			"Content-Type":                        "application/json",
			"Cross-Origin-Opener-Policy":          "same-origin",
			"Date":                                "Tue, 14 Jan 2025 12:48:44 GMT",
			"Expires":                             "Tue, 14 Jan 2025 15:48:44 GMT",
			"Vary":                                "Sec-Fetch-Site,Sec-Fetch-Mode,Sec-Fetch-Dest",
			"X-Content-Type-Options":              "nosniff",
			"X-Frame-Options":                     "SAMEORIGIN",
			"X-Xss-Protection":                    "0",
		},
	)
	defer ts.Close()
	c, err := NewClient(WithTest(t), WithCorpusDir(corpusDir), WithHTTPClient(&http.Client{}))
	require.NoError(t, err)
	require.NotNil(t, c)

	client := &http.Client{
		Transport: c,
	}

	req, err := http.NewRequest("POST", ts.URL(), nil)
	require.NoError(t, err)

	resp, err := client.Do(req)
	require.NoError(t, err)
	require.NotNil(t, resp)

	info, err := os.Stat(path.Join(corpusDir))
	require.NoError(t, err)
	require.True(t, info.IsDir())

	dir, err := os.ReadDir(corpusDir)
	require.NoError(t, err)
	require.Len(t, dir, 1)
	nameparts := strings.Split(dir[0].Name(), ".")
	require.Len(t, nameparts, 3)
	require.Equal(t, t.Name(), nameparts[0])
	require.Len(t, nameparts[1], 64)
	require.Equal(t, "json", nameparts[2])

	require.Equal(t, 1, ts.ReqCounter, "First request should have been served by the test server, req counter indicates that it was not")

	resp, err = client.Do(req)
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, 1, ts.ReqCounter, "Second request should be served from cache, req counter indicates that it was not")

	// Make a request to a new endpoint to ensure that the cache is not shared between requests to different endpoints
	otherTs := newTestServer(500, []byte("Internal Server Error"), map[string]string{"Content-Type": "text/plain"})
	defer otherTs.Close()
	req, err = http.NewRequest("GET", otherTs.URL(), nil)
	require.NoError(t, err)
	require.NotNil(t, resp)

	resp, err = client.Do(req)
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, 1, otherTs.ReqCounter, "Third request should not be served from cache, req counter indicates that it was")

	resp, err = client.Do(req)
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, 1, otherTs.ReqCounter, "Fourth request should be served from cache, req counter indicates that it was not")
}

func TestClient_no_test_provided(t *testing.T) {
	client, err := NewClient()
	require.Nil(t, client)
	require.ErrorIs(t, err, ErrNoTestProvided)
}

func TestClient_has_default_corpus_dir(t *testing.T) {
	client, err := NewClient(WithTest(t))
	require.NoError(t, err)
	require.NotNil(t, client)
	require.Equal(t, defaultClientOptions.corpusDir, client.options.corpusDir)
}

func TestRecordedResponse_to_and_from_json(t *testing.T) {
	cases := []struct {
		name  string
		input recordedResponse
	}{
		{
			name:  "basic request",
			input: recordedResponse{StatusCode: 200, Body: []byte("Hello, world!"), Header: map[string][]string{"Content-Type": {"text/plain"}}},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			b := bytes.Buffer{}
			err := json.NewEncoder(&b).Encode(tc.input)
			require.NoError(t, err)

			b64 := base64.StdEncoding.EncodeToString(tc.input.Body)

			raw := map[string]json.RawMessage{}
			err = json.Unmarshal(b.Bytes(), &raw)
			require.NoError(t, err)
			require.Equal(t, b64, string([]byte(raw["Body"])[1:len(raw["Body"])-1])) // skip first and last byte as those are the doublequotes

			output := recordedResponse{}
			err = json.NewDecoder(bytes.NewBuffer(b.Bytes())).Decode(&output)
			require.NoError(t, err)

			require.Equal(t, tc.input, output)
		})
	}
}

func TestFingerprintRequest(t *testing.T) {
	cases := []struct {
		name     string
		method   string
		url      string
		body     io.Reader
		header   map[string][]string
		expected string
	}{
		{
			name:     "basic request",
			method:   "GET",
			url:      "https://example.com",
			body:     bytes.NewReader([]byte("Hello, world!")),
			header:   map[string][]string{"Content-Type": {"text/plain"}},
			expected: "871c0c4df03bff56a0ae78a61f12f33259470b95d23f8ed55f0fb06ad8c9775e",
		},
		{
			name:     "nil body",
			method:   "GET",
			url:      "https://example.com",
			body:     nil,
			header:   map[string][]string{"Content-Type": {"text/plain"}},
			expected: "0b3ca9c3d9f8c49ee3f54618858f9bdca9c7eebd7a7b9f38e2c4a2a6445c8d62",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			req, err := http.NewRequest(tc.method, tc.url, tc.body)
			require.NoError(t, err)
			req.Header = tc.header

			fp, err := NewRequestFingerprint(req)
			require.NoError(t, err)
			require.NotEmpty(t, fp)
			require.Equal(t, tc.expected, fp.Fingerprint())
		})
	}
}

func TestEmbed(t *testing.T) {
	client := Embed(&http.Client{}, WithTest(t))
	require.NotNil(t, client.Transport)
}

func TestEmbedPanic(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("The code did not panic")
		}
	}()
	client := &http.Client{}
	client = Embed(client)
}

package goproxy

import (
	"bytes"
	"github.com/chainalysis-oss/oslc"
	ownHTTP "github.com/chainalysis-oss/oslc/http"
	"github.com/chainalysis-oss/oslc/httptestcorpus"
	"github.com/stretchr/testify/require"
	"io"
	"log/slog"
	"net/http"
	"testing"
)

func TestClient(t *testing.T) {
	t.Run("Create new client", func(t *testing.T) {
		client, err := NewClient()
		require.NoError(t, err)
		require.NotNil(t, client)

		t.Run("with default logger", func(t *testing.T) {
			require.Equal(t, slog.Default(), client.options.Logger)
		})
		t.Run("with default HTTP client", func(t *testing.T) {
			require.NotNil(t, client.options.HttpClient)
		})
		t.Run("with default base URL", func(t *testing.T) {
			require.Equal(t, client.options.BaseURL, client.options.BaseURL)
		})
	})
	t.Run("Get new client with logger", func(t *testing.T) {
		customLogger := slog.New(slog.NewTextHandler(io.Discard, nil))
		client, err := NewClient(WithLogger(customLogger))
		require.NoError(t, err)
		require.Equal(t, customLogger, client.options.Logger)
	})
	t.Run("Get new client with custom HTTP client", func(t *testing.T) {
		t.Skip("Not implemented")
	})
}

func getHttpClient(t *testing.T) *ownHTTP.Client {
	t.Helper()
	client, err := ownHTTP.NewClient(ownHTTP.WithHTTPClient(httptestcorpus.Embed(&http.Client{}, httptestcorpus.WithTest(t))))
	require.NoError(t, err)
	return client
}

func TestGetInfo_for_specific_version(t *testing.T) {
	client, err := NewClient(WithHTTPClient(getHttpClient(t)))
	require.NoError(t, err)
	require.NotNil(t, client)
	testcases := []struct {
		version  string
		expected versionInfo
	}{
		{"v0.3.0", versionInfo{
			Version: "v0.3.0",
			Time:    "2025-01-06T15:15:29Z",
			Origin: versionOrigin{
				VCS:  "git",
				URL:  "https://github.com/chainalysis-oss/oslc",
				Hash: "755c6565c94d5ff6fd1fbaab923e36be424360f0",
				Ref:  "refs/tags/v0.3.0",
			},
		}},
		{"v0.2.0", versionInfo{
			Version: "v0.2.0",
			Time:    "2024-12-26T10:43:42Z",
			Origin: versionOrigin{
				VCS:  "git",
				URL:  "https://github.com/chainalysis-oss/oslc",
				Hash: "c6fd791e9ea1adf481f1feef237c142b975e9e1c",
				Ref:  "refs/tags/v0.2.0",
			},
		}},
	}
	for _, tc := range testcases {
		t.Run(tc.version, func(t *testing.T) {
			resp, err := client.getInfo("github.com/chainalysis-oss/oslc", tc.version)
			require.NoError(t, err)
			require.Equal(t, tc.expected, resp)
		})
	}
}

func TestGetInfo_for_latest(t *testing.T) {
	client, err := NewClient(WithHTTPClient(getHttpClient(t)))
	require.NoError(t, err)
	require.NotNil(t, client)
	testcases := []struct {
		module   string
		expected versionInfo
	}{
		{"github.com/chainalysis-oss/oslc", versionInfo{
			Version: "v0.3.0",
			Time:    "2025-01-06T15:15:29Z",
			Origin: versionOrigin{
				VCS:  "git",
				URL:  "https://github.com/chainalysis-oss/oslc",
				Hash: "755c6565c94d5ff6fd1fbaab923e36be424360f0",
				Ref:  "refs/tags/v0.3.0",
			},
		}},
	}
	for _, tc := range testcases {
		t.Run(tc.module, func(t *testing.T) {
			resp, err := client.getInfo(tc.module, "")
			require.NoError(t, err)
			require.Equal(t, tc.expected, resp)
		})
	}
}

func TestGetInfo_for_bad_module(t *testing.T) {
	client, err := NewClient(WithHTTPClient(getHttpClient(t)))
	require.NoError(t, err)
	require.NotNil(t, client)
	testcases := []struct {
		module string
	}{
		{"invalid"},
		{"github.com/chainalysis-oss/oslcinvalid"},
		{"github.com/chainalsyis-oss"},
	}
	for _, tc := range testcases {
		t.Run(tc.module, func(t *testing.T) {
			resp, err := client.getInfo(tc.module, "")
			require.ErrorIs(t, err, oslc.ErrNoSuchPackage)
			require.Empty(t, resp)
		})
	}
}

func TestGetInfo_for_bad_version(t *testing.T) {
	client, err := NewClient(WithHTTPClient(getHttpClient(t)))
	require.NoError(t, err)
	require.NotNil(t, client)
	testcases := []struct {
		version string
	}{
		{"invalid"},
	}
	for _, tc := range testcases {
		t.Run(tc.version, func(t *testing.T) {
			resp, err := client.getInfo("github.com/chainalysis-oss/oslc", tc.version)
			require.ErrorIs(t, err, oslc.ErrVersionNotFound)
			require.Empty(t, resp)
		})
	}
}

func setupHttpClientWithStatusAndBody(t *testing.T, status int, body string) *ownHTTP.Client {
	t.Helper()
	mock := ownHTTP.NewTestHTTPClient(func(req *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: status,
			Body:       io.NopCloser(bytes.NewBuffer([]byte(body))),
		}, nil
	})
	client, err := ownHTTP.NewClient(ownHTTP.WithHTTPClient(mock))
	require.NoError(t, err)
	return client
}

func TestGetInfo_upstream_error(t *testing.T) {
	cases := []struct {
		name          string
		code          int
		body          string
		expectedError error
	}{
		{"not found", 404, "not found", &oslc.DistributorError{}},
		{"internal server error", 500, "internal server error", &oslc.DistributorError{}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			client, err := NewClient(WithHTTPClient(setupHttpClientWithStatusAndBody(t, tc.code, tc.body)))
			require.NoError(t, err)
			require.NotNil(t, client)

			resp, err := client.getInfo("thisdoesnotmatter", "alsodoesnotmatter")
			require.Empty(t, resp)
			require.ErrorAs(t, err, tc.expectedError)
		})
	}
}

func TestClient_implements_oslc_distributor_client(t *testing.T) {
	var _ oslc.DistributorClient = &Client{}
}

func TestClient_GetPackageVersion(t *testing.T) {
	client, err := NewClient(WithHTTPClient(getHttpClient(t)), WithTempDir(t.TempDir()))
	require.NoError(t, err)
	require.NotNil(t, client)

	cases := []struct {
		name     string
		module   string
		version  string
		expected oslc.Entry
	}{
		{
			name:    "valid version",
			module:  "github.com/keltia/leftpad",
			version: "v0.1.0",
			expected: oslc.Entry{
				Name:    "github.com/keltia/leftpad",
				Version: "v0.1.0",
				License: "BSD-2-Clause",
				DistributionPoints: []oslc.DistributionPoint{
					{
						Name:        "github.com/keltia/leftpad",
						URL:         "https://proxy.golang.org/github.com/keltia/leftpad/@v/v0.1.0.zip",
						Distributor: oslc.DistributorGo,
					},
				},
			},
		},
		{
			name:    "valid latest version",
			module:  "github.com/keltia/leftpad",
			version: "",
			expected: oslc.Entry{
				Name:    "github.com/keltia/leftpad",
				Version: "v0.1.0",
				License: "BSD-2-Clause",
				DistributionPoints: []oslc.DistributionPoint{
					{
						Name:        "github.com/keltia/leftpad",
						URL:         "https://proxy.golang.org/github.com/keltia/leftpad/@v/v0.1.0.zip",
						Distributor: oslc.DistributorGo,
					},
				},
			},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			resp, err := client.GetPackageVersion(tc.module, tc.version)
			require.NoError(t, err)
			require.Equal(t, tc.expected, resp)
		})
	}
}

func TestClient_GetPackage(t *testing.T) {
	client, err := NewClient(WithHTTPClient(getHttpClient(t)), WithTempDir(t.TempDir()))
	require.NoError(t, err)
	require.NotNil(t, client)

	cases := []struct {
		name     string
		module   string
		expected oslc.Entry
	}{
		{
			name:   "valid",
			module: "github.com/keltia/leftpad",
			expected: oslc.Entry{
				Name:    "github.com/keltia/leftpad",
				Version: "v0.1.0",
				License: "BSD-2-Clause",
				DistributionPoints: []oslc.DistributionPoint{
					{
						Name:        "github.com/keltia/leftpad",
						URL:         "https://proxy.golang.org/github.com/keltia/leftpad/@v/v0.1.0.zip",
						Distributor: oslc.DistributorGo,
					},
				},
			},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			resp, err := client.GetPackage(tc.module)
			require.NoError(t, err)
			require.Equal(t, tc.expected, resp)
		})
	}
}

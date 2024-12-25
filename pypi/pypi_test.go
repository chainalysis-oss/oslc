package pypi

import (
	"bytes"
	"github.com/chainalysis-oss/oslc"
	ownHTTP "github.com/chainalysis-oss/oslc/http"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"testing"
)

func TestPypiPackageResponse_AsEntry(t *testing.T) {
	tests := []struct {
		name string
		pkg  pypiPackageResponse
		want oslc.Entry
	}{
		{
			name: "test",
			pkg: pypiPackageResponse{
				Releases: map[string][]struct{}{
					"test3": {{}, {}},
				},
				Info: struct {
					Name        string `json:"name"`
					License     string `json:"license"`
					PackageURL  string `json:"package_url"`
					ProjectURLs struct {
						Source string `json:"Source"`
					} `json:"project_urls"`
					Version string `json:"version"`
				}{
					Name:       "test",
					License:    "test2",
					PackageURL: "https://example.com",
					Version:    "test3",
				},
			},
			want: oslc.Entry{
				Name: "test",
				DistributionPoints: []oslc.DistributionPoint{{
					Name:        "test",
					URL:         "https://example.com",
					Distributor: oslc.DistributorPypi,
				}},
				License: "test2",
				Version: "test3",
			},
		},
		{
			name: "test2",
			pkg: pypiPackageResponse{
				Releases: map[string][]struct{}{
					"test3": {{}, {}},
				},
				Info: struct {
					Name        string `json:"name"`
					License     string `json:"license"`
					PackageURL  string `json:"package_url"`
					ProjectURLs struct {
						Source string `json:"Source"`
					} `json:"project_urls"`
					Version string `json:"version"`
				}{
					Name:    "test",
					Version: "test3",
				},
			},
			want: oslc.Entry{
				Name:    "test",
				License: "Unknown",
				Version: "test3",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.want, tt.pkg.AsEntry())
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

func setupHttpClientWithBody(t *testing.T, body string) *ownHTTP.Client {
	t.Helper()
	return setupHttpClientWithStatusAndBody(t, http.StatusOK, body)
}

func setupClient(t *testing.T, httpClient *ownHTTP.Client) *Client {
	t.Helper()
	c, err := NewClient(WithHTTPClient(httpClient))
	require.NoError(t, err)
	return c
}

func TestClient_GetPackageVersion(t *testing.T) {
	testcases := []struct {
		name       string
		pkgName    string
		pkgVersion string
		body       string
		expected   oslc.Entry
	}{
		{
			name:       "correct response",
			pkgName:    "test",
			pkgVersion: "test2",
			body:       `{"info":{"name":"test","version":"test2","license":"test3"}}`,
			expected: oslc.Entry{
				Name:    "test",
				Version: "test2",
				License: "test3",
			},
		},
		{
			name:       "no license info in response",
			pkgName:    "test",
			pkgVersion: "test2",
			body:       `{"info":{"name":"test","version":"test2"}}`,
			expected: oslc.Entry{
				Name:    "test",
				Version: "test2",
				License: "Unknown",
			},
		},
	}
	for _, tt := range testcases {
		t.Run(tt.name, func(t *testing.T) {
			c := setupClient(t, setupHttpClientWithBody(t, tt.body))
			out, err := c.GetPackageVersion(tt.pkgName, tt.pkgVersion)
			require.NoError(t, err)
			require.Equal(t, tt.expected, out)
		})
	}
}

func TestClient_GetPackageVersion_version_path(t *testing.T) {
	mock := ownHTTP.NewTestHTTPClient(func(req *http.Request) (*http.Response, error) {
		require.Equal(t, "https://pypi.org/pypi/test/1.0.0/json", req.URL.String())
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(bytes.NewBuffer([]byte(`{}`))),
		}, nil
	})
	httpClient, err := ownHTTP.NewClient(ownHTTP.WithHTTPClient(mock))
	require.NoError(t, err)
	c := setupClient(t, httpClient)
	_, err = c.GetPackageVersion("test", "1.0.0")
	require.NoError(t, err)
}

func TestClient_GetPackageVersion_no_version_path(t *testing.T) {
	mock := ownHTTP.NewTestHTTPClient(func(req *http.Request) (*http.Response, error) {
		require.Equal(t, "https://pypi.org/pypi/test/json", req.URL.String())
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(bytes.NewBuffer([]byte(`{}`))),
		}, nil
	})
	httpClient, err := ownHTTP.NewClient(ownHTTP.WithHTTPClient(mock))
	require.NoError(t, err)
	c := setupClient(t, httpClient)
	_, err = c.GetPackageVersion("test", "")
	require.NoError(t, err)
}

func TestClient_GetPackageVersion_http_client_error(t *testing.T) {
	mock := ownHTTP.NewTestHTTPClient(func(req *http.Request) (*http.Response, error) {
		return nil, assert.AnError
	})
	httpClient, err := ownHTTP.NewClient(ownHTTP.WithHTTPClient(mock))
	require.NoError(t, err)
	c := setupClient(t, httpClient)
	_, err = c.GetPackageVersion("test", "")
	assert.Error(t, err)
}

func TestClient_GetPackageVersion_http_client_status_error(t *testing.T) {
	c := setupClient(t, setupHttpClientWithStatusAndBody(t, http.StatusNotFound, ""))
	_, err := c.GetPackageVersion("test", "")
	assert.Error(t, err)
}

func TestClient_GetPackageVersion_json_decode_error(t *testing.T) {
	c := setupClient(t, setupHttpClientWithBody(t, "test"))
	_, err := c.GetPackageVersion("test", "")
	assert.Error(t, err)
}

func TestClient_GetPackage(t *testing.T) {
	mock := ownHTTP.NewTestHTTPClient(func(req *http.Request) (*http.Response, error) {
		require.Equal(t, "https://pypi.org/pypi/test/json", req.URL.String())
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(bytes.NewBuffer([]byte(`{}`))),
		}, nil
	})
	httpClient, err := ownHTTP.NewClient(ownHTTP.WithHTTPClient(mock))
	require.NoError(t, err)
	c := setupClient(t, httpClient)
	_, err = c.GetPackage("test")
	require.NoError(t, err)
}

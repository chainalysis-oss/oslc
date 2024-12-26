package npm

import (
	"bytes"
	"encoding/json"
	"github.com/chainalysis-oss/oslc"
	ownHTTP "github.com/chainalysis-oss/oslc/http"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"os"
	"testing"
)

func TestNpmPackageResponse_AsEntry(t *testing.T) {
	tests := []struct {
		name string
		pkg  npmPackageResponse
		want oslc.Entry
	}{
		{
			name: "includes_top_level_version",
			pkg: npmPackageResponse{
				Name:    "test",
				Version: "test3",
			},
			want: oslc.Entry{
				Name:    "test",
				Version: "test3",
				License: "Unknown",
				DistributionPoints: []oslc.DistributionPoint{{
					Name:        "test",
					URL:         "https://www.npmjs.com/package/test",
					Distributor: oslc.DistributorNpm,
				}},
			},
		},
		{
			name: "includes_latest_version_from_versions",
			pkg: npmPackageResponse{
				Name: "test",
				Versions: map[string]struct {
					Version    string `json:"version"`
					License    string `json:"license"`
					Repository struct {
						Type string `json:"type"`
						URL  string `json:"url"`
					} `json:"repository"`
				}{
					"test2": {
						Version: "test2",
					},
					"test3": {
						Version: "test3",
					},
				},
				DistTags: struct {
					Latest string `json:"latest"`
				}{
					Latest: "test3",
				},
			},
			want: oslc.Entry{
				Name:    "test",
				Version: "test3",
				License: "Unknown",
				DistributionPoints: []oslc.DistributionPoint{{
					Name:        "test",
					URL:         "https://www.npmjs.com/package/test",
					Distributor: oslc.DistributorNpm,
				}},
			},
		},
		{
			name: "unknown_version_and_license",
			pkg: npmPackageResponse{
				Name: "test",
			},
			want: oslc.Entry{
				Name:    "test",
				Version: "Unknown",
				License: "Unknown",
				DistributionPoints: []oslc.DistributionPoint{{
					Name:        "test",
					URL:         "https://www.npmjs.com/package/test",
					Distributor: oslc.DistributorNpm,
				}},
			},
		},
		{
			name: "include_top_level_license",
			pkg: npmPackageResponse{
				Name:    "test",
				Version: "test2",
				License: "test3",
			},
			want: oslc.Entry{
				Name:    "test",
				Version: "test2",
				License: "test3",
				DistributionPoints: []oslc.DistributionPoint{{
					Name:        "test",
					URL:         "https://www.npmjs.com/package/test",
					Distributor: oslc.DistributorNpm,
				}},
			},
		},
		{
			name: "include_latest_license_from_versions",
			pkg: npmPackageResponse{
				Name: "test",
				Versions: map[string]struct {
					Version    string `json:"version"`
					License    string `json:"license"`
					Repository struct {
						Type string `json:"type"`
						URL  string `json:"url"`
					} `json:"repository"`
				}{
					"test2": {
						Version: "test2",
					},
					"test3": {
						Version: "test3",
						License: "test4",
					},
				},
				DistTags: struct {
					Latest string `json:"latest"`
				}{
					Latest: "test3",
				},
			},
			want: oslc.Entry{
				Name:    "test",
				Version: "test3",
				License: "test4",
				DistributionPoints: []oslc.DistributionPoint{{
					Name:        "test",
					URL:         "https://www.npmjs.com/package/test",
					Distributor: oslc.DistributorNpm,
				}},
			},
		},
		{
			name: "npm_latest",
			pkg:  readJson(t, "testdata/npm_latest.json"),
			want: oslc.Entry{
				Name: "npm",
				DistributionPoints: []oslc.DistributionPoint{{
					Name:        "npm",
					URL:         "https://www.npmjs.com/package/npm",
					Distributor: oslc.DistributorNpm,
				}},
				License: "Artistic-2.0",
				Version: "10.8.3",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.want, tt.pkg.AsEntry())
		})
	}
}

func readJson(t *testing.T, path string) npmPackageResponse {
	var pkg npmPackageResponse
	jsonFile, err := os.Open(path)
	require.NoError(t, err)
	defer jsonFile.Close()
	require.NoError(t, json.NewDecoder(jsonFile).Decode(&pkg))
	return pkg
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
			body:       `{"name":"test","version":"test2","license":"test3"}`,
			expected: oslc.Entry{
				Name:    "test",
				Version: "test2",
				License: "test3",
				DistributionPoints: []oslc.DistributionPoint{{
					Name:        "test",
					URL:         "https://www.npmjs.com/package/test",
					Distributor: oslc.DistributorNpm,
				}},
			},
		},
		{
			name:       "no license info in response",
			pkgName:    "test",
			pkgVersion: "test2",
			body:       `{"name":"test","version":"test2"}`,
			expected: oslc.Entry{
				Name:    "test",
				Version: "test2",
				License: "Unknown",
				DistributionPoints: []oslc.DistributionPoint{{
					Name:        "test",
					URL:         "https://www.npmjs.com/package/test",
					Distributor: oslc.DistributorNpm,
				}},
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
		require.Equal(t, "https://registry.npmjs.org/test/1.0.0", req.URL.String())
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
		require.Equal(t, "https://registry.npmjs.org/test", req.URL.String())
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
		require.Equal(t, "https://registry.npmjs.org/test", req.URL.String())
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

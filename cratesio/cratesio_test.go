package cratesio

import (
	"bytes"
	"fmt"
	"github.com/chainalysis-oss/oslc"
	ownHTTP "github.com/chainalysis-oss/oslc/http"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"os"
	"testing"
)

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

func TestCrateResponse_newestVersion(t *testing.T) {
	testcases := []struct {
		name          string
		c             crateResponse
		expected      crateVersion
		expectedError error
	}{
		{
			name: "correct response",
			c: crateResponse{
				Versions: []crateVersion{
					{
						Num: "test2",
					},
				},
				Crate: crate{
					NewestVersion: "test2",
				},
			},
			expected: crateVersion{
				Num: "test2",
			},
		},
		{
			name: "no versions in crateResponse",
			c: crateResponse{
				Versions: []crateVersion{},
			},
			expectedError: oslc.ErrVersionNotFound,
		},
		{
			name: "crateResponse has no newest version",
			c: crateResponse{
				Versions: []crateVersion{
					{
						Num: "test2",
					},
				},
			},
			expectedError: oslc.ErrVersionNotFound,
		},
		{
			name: "crateResponse specifies a version not included in the crateResponse",
			c: crateResponse{
				Versions: []crateVersion{
					{
						Num: "test3",
					},
				},
				Crate: crate{
					NewestVersion: "test2",
				},
			},
			expectedError: oslc.ErrVersionNotFound,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := tc.c.newestVersion()
			if tc.expectedError == nil {
				require.NoError(t, err)
			}
			require.Equal(t, tc.expected, result)
			require.ErrorIs(t, err, tc.expectedError)
		})
	}
}

func TestCrateVersion_AsEntry(t *testing.T) {
	testcases := []struct {
		name     string
		p        crateVersion
		expected oslc.Entry
	}{
		{
			name: "correct response",
			p: crateVersion{
				Crate:   "test",
				Num:     "test2",
				License: "test3",
				Links: struct {
					VersionDownloads string `json:"version_downloads"`
				}{
					VersionDownloads: "somewhere",
				},
			},
			expected: oslc.Entry{
				Name:    "test",
				Version: "test2",
				License: "test3",
				DistributionPoints: []oslc.DistributionPoint{{
					Name:        "test",
					URL:         fmt.Sprintf("%ssomewhere", cratesIOBaseURL),
					Distributor: oslc.DistributorCratesIo,
				}},
			},
		},
		{
			name: "no license info in response",
			p: crateVersion{
				Crate: "test",
				Num:   "test2",
				Links: struct {
					VersionDownloads string `json:"version_downloads"`
				}{
					VersionDownloads: "somewhere",
				},
			},
			expected: oslc.Entry{
				Name:    "test",
				Version: "test2",
				License: "Unknown",
				DistributionPoints: []oslc.DistributionPoint{{
					Name:        "test",
					URL:         fmt.Sprintf("%ssomewhere", cratesIOBaseURL),
					Distributor: oslc.DistributorCratesIo,
				}},
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			result := tc.p.AsEntry()
			require.Equal(t, tc.expected, result)
		})
	}
}

func getTestData(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	require.NoError(t, err)
	return string(data)
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
			name:       "package and version correct",
			pkgName:    "snarkvm-marlin",
			pkgVersion: "0.8.0",
			body:       getTestData(t, "testdata/crateversion.json"),
			expected: oslc.Entry{
				Name:    "snarkvm-marlin",
				Version: "0.8.0",
				License: "GPL-3.0",
				DistributionPoints: []oslc.DistributionPoint{{
					Name:        "snarkvm-marlin",
					URL:         fmt.Sprintf("%s/api/v1/crates/snarkvm-marlin/0.8.0/downloads", cratesIOBaseURL),
					Distributor: oslc.DistributorCratesIo,
				}},
			},
		},
		{
			name:       "package and latest version correct",
			pkgName:    "snarkvm-marlin",
			pkgVersion: "",
			body:       getTestData(t, "testdata/crate.json"),
			expected: oslc.Entry{
				Name:    "snarkvm-marlin",
				Version: "0.8.0",
				License: "GPL-3.0",
				DistributionPoints: []oslc.DistributionPoint{{
					Name:        "snarkvm-marlin",
					URL:         fmt.Sprintf("%s/api/v1/crates/snarkvm-marlin/0.8.0/downloads", cratesIOBaseURL),
					Distributor: oslc.DistributorCratesIo,
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

func TestClient_GetPackageVersion_http_client_error(t *testing.T) {
	mock := ownHTTP.NewTestHTTPClient(func(req *http.Request) (*http.Response, error) {
		return nil, assert.AnError
	})
	httpClient, err := ownHTTP.NewClient(ownHTTP.WithHTTPClient(mock))
	require.NoError(t, err)
	c := setupClient(t, httpClient)
	_, err = c.GetPackageVersion("test", "")
	require.Error(t, err)
}

func TestClient_GetPackageVersion_http_client_status_error(t *testing.T) {
	c := setupClient(t, setupHttpClientWithStatusAndBody(t, http.StatusNotFound, ""))
	_, err := c.GetPackageVersion("test", "")
	require.Error(t, err)
}

func TestClient_GetPackageVersion_latest_version_json_decode_error(t *testing.T) {
	c := setupClient(t, setupHttpClientWithBody(t, "test"))
	_, err := c.GetPackageVersion("test", "")
	require.Error(t, err)
}

func TestClient_GetPackageVersion_specific_version_json_decode_error(t *testing.T) {
	c := setupClient(t, setupHttpClientWithBody(t, "test"))
	_, err := c.GetPackageVersion("test", "0.0.1")
	require.Error(t, err)
}

func TestClient_GetPackageVersion_latest_version_no_version(t *testing.T) {
	c := setupClient(t, setupHttpClientWithBody(t, "{}"))
	_, err := c.GetPackageVersion("test", "")
	require.Error(t, err)
}

func TestClient_GetPackage(t *testing.T) {
	c := setupClient(t, setupHttpClientWithBody(t, getTestData(t, "testdata/crate.json")))
	out, err := c.GetPackage("snarkvm-marlin")
	require.NoError(t, err)
	require.NotEmpty(t, out)
}

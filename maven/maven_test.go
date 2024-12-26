package maven

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

func TestMavenPOM_AsEntry(t *testing.T) {
	tests := []struct {
		name string
		pom  mavenPOM
		want oslc.Entry
	}{
		{
			name: "testArtId",
			pom: mavenPOM{
				Licenses: []struct {
					License struct {
						Name string `xml:"name"`
					} `xml:"license"`
				}{
					{
						License: struct {
							Name string `xml:"name"`
						}{
							Name: "test",
						},
					},
				},
				GroupId:    "testGroupId",
				ArtifactId: "testArtId",
				Version:    "test",
			},
			want: oslc.Entry{
				Name:    "testGroupId:testArtId",
				Version: "test",
				License: "test",
				DistributionPoints: []oslc.DistributionPoint{{
					Name:        "testGroupId:testArtId",
					URL:         "https://central.sonatype.com/artifact/testGroupId/testArtId",
					Distributor: oslc.DistributorMaven,
				}},
			},
		},
		{
			name: "missing_license_info",
			pom: mavenPOM{
				GroupId:    "testGroupId",
				ArtifactId: "testArtId",
				Version:    "test",
			},
			want: oslc.Entry{
				Name:    "testGroupId:testArtId",
				Version: "test",
				License: "Unknown",
				DistributionPoints: []oslc.DistributionPoint{{
					Name:        "testGroupId:testArtId",
					URL:         "https://central.sonatype.com/artifact/testGroupId/testArtId",
					Distributor: oslc.DistributorMaven,
				}},
			},
		},
		{
			name: "missing_version_info",
			pom: mavenPOM{
				Licenses: []struct {
					License struct {
						Name string `xml:"name"`
					} `xml:"license"`
				}{
					{
						License: struct {
							Name string `xml:"name"`
						}{
							Name: "test",
						},
					},
				},
				GroupId:    "testGroupId",
				ArtifactId: "testArtId",
			},
			want: oslc.Entry{
				Name:    "testGroupId:testArtId",
				Version: "Unknown",
				License: "test",
				DistributionPoints: []oslc.DistributionPoint{{
					Name:        "testGroupId:testArtId",
					URL:         "https://central.sonatype.com/artifact/testGroupId/testArtId",
					Distributor: oslc.DistributorMaven,
				}},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.want, tt.pom.AsEntry())
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
			pkgName:    "testGroupId:testArtifactId",
			pkgVersion: "test2",
			body: `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<project xmlns="http://maven.apache.org/POM/4.0.0" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:schemaLocation="http://maven.apache.org/POM/4.0.0 http://maven.apache.org/xsd/maven-4.0.0.xsd">
<licenses>
<license>
<name>test3</name>
</license>
</licenses>
<groupId>testGroupId</groupId>
<artifactId>testArtifactId</artifactId>
<version>test2</version>
</project>`,
			expected: oslc.Entry{
				Name:    "testGroupId:testArtifactId",
				Version: "test2",
				License: "test3",
				DistributionPoints: []oslc.DistributionPoint{{
					Name:        "testGroupId:testArtifactId",
					URL:         "https://central.sonatype.com/artifact/testGroupId/testArtifactId",
					Distributor: oslc.DistributorMaven,
				}},
			},
		},
		{
			name:       "no license info in response",
			pkgName:    "testGroupId:testArtifactId",
			pkgVersion: "test2",
			body: `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<project xmlns="http://maven.apache.org/POM/4.0.0" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:schemaLocation="http://maven.apache.org/POM/4.0.0 http://maven.apache.org/xsd/maven-4.0.0.xsd">
<groupId>testGroupId</groupId>
<artifactId>testArtifactId</artifactId>
<version>test2</version>
</project>`,
			expected: oslc.Entry{
				Name:    "testGroupId:testArtifactId",
				Version: "test2",
				License: "Unknown",
				DistributionPoints: []oslc.DistributionPoint{{
					Name:        "testGroupId:testArtifactId",
					URL:         "https://central.sonatype.com/artifact/testGroupId/testArtifactId",
					Distributor: oslc.DistributorMaven,
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

func TestClient_GetPackageVersion_latest(t *testing.T) {
	mock := ownHTTP.NewTestHTTPClient(func(req *http.Request) (*http.Response, error) {
		require.Equal(t, "https://search.maven.org/solrsearch/select?q=g:testGroupId+AND+a:testArtifactId&rows=1&wt=json", req.URL.String())
		return &http.Response{
			StatusCode: http.StatusNotFound,
			Body:       http.NoBody,
		}, nil
	})
	httpClient, err := ownHTTP.NewClient(ownHTTP.WithHTTPClient(mock))
	require.NoError(t, err)
	c := setupClient(t, httpClient)
	_, err = c.GetPackageVersion("testGroupId:testArtifactId", "latest")
	assert.Error(t, err)
}

func TestClient_GetPackageVersion_version_path(t *testing.T) {
	mock := ownHTTP.NewTestHTTPClient(func(req *http.Request) (*http.Response, error) {
		require.Equal(t, "https://search.maven.org/remotecontent?filepath=testGroupId/testArtifactId/1.0.0/testArtifactId-1.0.0.pom", req.URL.String())
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(bytes.NewBuffer([]byte(`<project></project>`))),
		}, nil
	})
	httpClient, err := ownHTTP.NewClient(ownHTTP.WithHTTPClient(mock))
	require.NoError(t, err)
	c := setupClient(t, httpClient)
	_, err = c.GetPackageVersion("testGroupId:testArtifactId", "1.0.0")
	require.NoError(t, err)
}

func TestClient_GetPackageVersion_latest_version_error(t *testing.T) {
	mock := ownHTTP.NewTestHTTPClient(func(req *http.Request) (*http.Response, error) {
		return nil, assert.AnError
	})
	httpClient, err := ownHTTP.NewClient(ownHTTP.WithHTTPClient(mock))
	require.NoError(t, err)
	c := setupClient(t, httpClient)
	_, err = c.GetPackageVersion("testGroupId:testArtifactId", "")
	assert.Error(t, err)
}

func TestClient_GetPackageVersion_http_client_error(t *testing.T) {
	mock := ownHTTP.NewTestHTTPClient(func(req *http.Request) (*http.Response, error) {
		return nil, assert.AnError
	})
	httpClient, err := ownHTTP.NewClient(ownHTTP.WithHTTPClient(mock))
	require.NoError(t, err)
	c := setupClient(t, httpClient)
	_, err = c.GetPackageVersion("testGroupId:testArtifactId", "1.0.0")
	assert.Error(t, err)
}

func TestClient_GetPackageVersion_http_client_status_error(t *testing.T) {
	c := setupClient(t, setupHttpClientWithStatusAndBody(t, http.StatusNotFound, ""))
	_, err := c.GetPackageVersion("testGroupId:testArtifactId", "1.0.0")
	assert.Error(t, err)
}

func TestClient_GetPackageVersion_xml_decode_error(t *testing.T) {
	c := setupClient(t, setupHttpClientWithBody(t, "test"))
	_, err := c.GetPackageVersion("testGroupId:testArtifactId", "1.0.0")
	assert.Error(t, err)
}

func TestClient_GetPackage(t *testing.T) {
	var count int
	mock := ownHTTP.NewTestHTTPClient(func(req *http.Request) (*http.Response, error) {
		if count == 0 {
			count++
			require.Equal(t, "https://search.maven.org/solrsearch/select?q=g:testGroupId+AND+a:testArtifactId&rows=1&wt=json", req.URL.String())
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(bytes.NewBuffer([]byte(`{"response":{"numFound":1,"docs":[{"id":"testGroupId:testArtifactId","latestVersion":"test2"}]}}`))),
			}, nil
		}
		require.Equal(t, "https://search.maven.org/remotecontent?filepath=testGroupId/testArtifactId/test2/testArtifactId-test2.pom", req.URL.String())
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(bytes.NewBuffer([]byte(`<project></project>`))),
		}, nil
	})
	httpClient, err := ownHTTP.NewClient(ownHTTP.WithHTTPClient(mock))
	require.NoError(t, err)
	c := setupClient(t, httpClient)
	_, err = c.GetPackage("testGroupId:testArtifactId")
	require.NoError(t, err)
}

func TestClient_GetLatestVersion_error_json_decode(t *testing.T) {
	mock := ownHTTP.NewTestHTTPClient(func(req *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(bytes.NewBuffer([]byte(`test`))),
		}, nil
	})
	httpClient, err := ownHTTP.NewClient(ownHTTP.WithHTTPClient(mock))
	require.NoError(t, err)
	c := setupClient(t, httpClient)
	_, err = c.GetLatestVersion("testGroupId", "testArtifactId")
	assert.Error(t, err)
}

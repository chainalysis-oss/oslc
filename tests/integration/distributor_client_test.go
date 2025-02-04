package integration

import (
	"github.com/chainalysis-oss/oslc"
	"github.com/chainalysis-oss/oslc/cratesio"
	"github.com/chainalysis-oss/oslc/goproxy"
	ownHTTP "github.com/chainalysis-oss/oslc/http"
	"github.com/chainalysis-oss/oslc/httptestcorpus"
	"github.com/chainalysis-oss/oslc/maven"
	"github.com/chainalysis-oss/oslc/npm"
	"github.com/chainalysis-oss/oslc/pypi"
	"github.com/stretchr/testify/require"
	"net/http"
	"testing"
)

func TestImplementations_of_DistributorClient(t *testing.T) {
	type inputVariables struct {
		versionNotFound           []string
		packageNotFound           []string
		packageAndVersionNotFound []string
	}
	cases := []struct {
		distributor             string
		iv                      inputVariables
		createDistributorClient func(t *testing.T) oslc.DistributorClient
	}{
		{
			distributor: oslc.DistributorPypi,
			iv: inputVariables{
				versionNotFound:           []string{"requests", "thisisnotaversion"},
				packageNotFound:           []string{"thisisnotapackage", ""},
				packageAndVersionNotFound: []string{"thisisnotapackage", "thisisnotaversion"},
			},
			createDistributorClient: func(t *testing.T) oslc.DistributorClient {
				httpClient, err := ownHTTP.NewClient(ownHTTP.WithHTTPClient(httptestcorpus.Embed(&http.Client{}, httptestcorpus.WithTest(t))))
				require.NoError(t, err)
				client, err := pypi.NewClient(pypi.WithHTTPClient(httpClient))
				require.NoError(t, err)
				return client
			},
		},
		{
			distributor: oslc.DistributorMaven,
			iv: inputVariables{
				versionNotFound:           []string{"org.apache.commons:commons-parent", "thisisnotaversion"},
				packageNotFound:           []string{"thisisnot:apackage", ""},
				packageAndVersionNotFound: []string{"thisisnot:apackage", "thisisnotaversion"},
			},
			createDistributorClient: func(t *testing.T) oslc.DistributorClient {
				httpClient, err := ownHTTP.NewClient(ownHTTP.WithHTTPClient(httptestcorpus.Embed(&http.Client{}, httptestcorpus.WithTest(t))))
				require.NoError(t, err)
				client, err := maven.NewClient(maven.WithHTTPClient(httpClient))
				require.NoError(t, err)
				return client
			},
		},
		{
			distributor: oslc.DistributorGo,
			iv: inputVariables{
				versionNotFound:           []string{"github.com/chainalysis-oss/oslc", "thisisnotaversion"},
				packageNotFound:           []string{"thisisnotapackage", ""},
				packageAndVersionNotFound: []string{"thisisnotapackage", "thisisnotaversion"},
			},
			createDistributorClient: func(t *testing.T) oslc.DistributorClient {
				httpClient, err := ownHTTP.NewClient(ownHTTP.WithHTTPClient(httptestcorpus.Embed(&http.Client{}, httptestcorpus.WithTest(t))))
				require.NoError(t, err)
				client, err := goproxy.NewClient(goproxy.WithHTTPClient(httpClient))
				require.NoError(t, err)
				return client
			},
		},
		{
			distributor: oslc.DistributorNpm,
			iv: inputVariables{
				versionNotFound:           []string{"@types/node", "thisisnotaversion"},
				packageNotFound:           []string{"thisisnotapackage", ""},
				packageAndVersionNotFound: []string{"thisisnotapackage", "thisisnotaversion"},
			},
			createDistributorClient: func(t *testing.T) oslc.DistributorClient {
				httpClient, err := ownHTTP.NewClient(ownHTTP.WithHTTPClient(httptestcorpus.Embed(&http.Client{}, httptestcorpus.WithTest(t))))
				require.NoError(t, err)
				client, err := npm.NewClient(npm.WithHTTPClient(httpClient))
				require.NoError(t, err)
				return client
			},
		},
		{
			distributor: oslc.DistributorCratesIo,
			iv: inputVariables{
				versionNotFound:           []string{"just-for-test", "999.999.999"},
				packageNotFound:           []string{"thisisnotapackage", ""},
				packageAndVersionNotFound: []string{"thisisnotapackage", "999.999.999"},
			},
			createDistributorClient: func(t *testing.T) oslc.DistributorClient {
				httpClient, err := ownHTTP.NewClient(ownHTTP.WithHTTPClient(httptestcorpus.Embed(&http.Client{}, httptestcorpus.WithTest(t))))
				require.NoError(t, err)
				client, err := cratesio.NewClient(cratesio.WithHTTPClient(httpClient))
				require.NoError(t, err)
				return client
			},
		},
	}
	for _, tc := range cases {
		t.Run(tc.distributor, func(t *testing.T) {
			require.Len(t, tc.iv.versionNotFound, 2)
			require.Len(t, tc.iv.packageNotFound, 2)
			require.Len(t, tc.iv.packageAndVersionNotFound, 2)
			client := tc.createDistributorClient(t)
			t.Run("version_not_found", func(t *testing.T) {
				resp, err := client.GetPackageVersion(tc.iv.versionNotFound[0], tc.iv.versionNotFound[1])
				require.Empty(t, resp)
				require.ErrorIs(t, err, oslc.ErrVersionNotFound)
			})
			t.Run("package_not_found", func(t *testing.T) {
				resp, err := client.GetPackageVersion(tc.iv.packageNotFound[0], tc.iv.packageNotFound[1])
				require.Empty(t, resp)
				require.ErrorIs(t, err, oslc.ErrNoSuchPackage)
			})
			t.Run("package_and_version_not_found", func(t *testing.T) {
				resp, err := client.GetPackageVersion(tc.iv.packageAndVersionNotFound[0], tc.iv.packageAndVersionNotFound[1])
				require.Empty(t, resp)
				require.ErrorIs(t, err, oslc.ErrNoSuchPackage)
			})
			t.Run("package_is_empty", func(t *testing.T) {
				resp, err := client.GetPackageVersion("", "")
				require.Empty(t, resp)
				require.ErrorIs(t, err, oslc.ErrNoSuchPackage)
			})
		})
	}
}

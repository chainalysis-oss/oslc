package oslc

import (
	ownHTTP "github.com/chainalysis-oss/oslc/http"
	"github.com/chainalysis-oss/oslc/httptestcorpus"
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
		createDistributorClient func(t *testing.T) DistributorClient
	}{
		{
			distributor: DistributorPypi,
			iv: inputVariables{
				versionNotFound:           []string{"requests", "thisisnotaversion"},
				packageNotFound:           []string{"thisisnotapackage", ""},
				packageAndVersionNotFound: []string{"thisisnotapackage", "thisisnotaversion"},
			},
			createDistributorClient: func(t *testing.T) DistributorClient {
				httpClient, err := ownHTTP.NewClient(ownHTTP.WithHTTPClient(httptestcorpus.Embed(&http.Client{}, httptestcorpus.WithTest(t))))
				require.NoError(t, err)
				client, err := pypi.NewClient(pypi.WithHTTPClient(httpClient))
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
				resp, err := client.GetPackageVersion(tc.iv.packageNotFound[0], tc.iv.packageNotFound[1])
				require.Empty(t, resp)
				require.ErrorIs(t, err, ErrNoSuchPackage)
			})
			t.Run("package_not_found", func(t *testing.T) {
				resp, err := client.GetPackageVersion(tc.iv.packageNotFound[0], tc.iv.packageNotFound[1])
				require.Empty(t, resp)
				require.ErrorIs(t, err, ErrNoSuchPackage)
			})
			t.Run("package_and_version_not_found", func(t *testing.T) {
				resp, err := client.GetPackageVersion(tc.iv.packageNotFound[0], tc.iv.packageNotFound[1])
				require.Empty(t, resp)
				require.ErrorIs(t, err, ErrNoSuchPackage)
			})
			t.Run("package_is_empty", func(t *testing.T) {
				resp, err := client.GetPackageVersion("", "")
				require.Empty(t, resp)
				require.ErrorIs(t, err, ErrNoSuchPackage)
			})
		})
	}
}

package sll

import (
	"github.com/chainalysis-oss/oslc"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestLr_Source(t *testing.T) {
	lr := &lr{}
	require.Equal(t, LicenseListSource, lr.Source())
}

func TestLr_Version(t *testing.T) {
	lr := &lr{}
	require.Equal(t, ll.lg.LicenseListVersion, lr.Version())
}

func TestLr_ReleaseDate(t *testing.T) {
	lr := &lr{}
	require.Equal(t, ll.lg.ReleaseDate, lr.ReleaseDate())
}

func TestLr_Lookup(t *testing.T) {
	t.Run("found", func(t *testing.T) {
		lr := &lr{}
		lic := Lookup("mit")
		require.Equal(t, oslc.License{
			Name: lic.Name,
			ID:   lic.LicenseID,
		}, lr.Lookup("mit"))
	})

	t.Run("not found", func(t *testing.T) {
		lr := &lr{}
		require.Equal(t, oslc.License{}, lr.Lookup("not-found"))
	})
}

func TestLr_Licenses(t *testing.T) {
	lr := &lr{}
	require.Equal(t, Licenses(), lr.Licenses())
}

func TestAsLicenseRetriever(t *testing.T) {
	require.Implements(t, (*oslc.LicenseRetriever)(nil), AsLicenseRetriever())
}

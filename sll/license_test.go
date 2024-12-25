package sll

import (
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/require"
	"reflect"
	"testing"
)

// Ensure that changes to the returned [License] object do not affect the original.
func TestLookup_immutability(t *testing.T) {
	out := Lookup("MIT")
	out.Name = "changed"
	require.NotEqual(t, "changed", Lookup("MIT").Name)
}

func ExampleLookup() {
	lic := Lookup("MIT")
	if reflect.DeepEqual(NoLicenseMatch, lic) {
		fmt.Println("License not found")
	}

	fmt.Println(lic.Reference)
	fmt.Println(lic.IsDeprecatedLicenseID)
	fmt.Println(lic.DetailsURL)
	fmt.Println(lic.ReferenceNumber)
	fmt.Println(lic.Name)
	fmt.Println(lic.LicenseID)
	fmt.Println(lic.SeeAlso)
	fmt.Println(lic.IsOSIApproved)
	fmt.Println(*lic.IsFSFLibre)
	// Output:
	// https://spdx.org/licenses/MIT.html
	// false
	// https://spdx.org/licenses/MIT.json
	// 601
	// MIT License
	// MIT
	// [https://opensource.org/license/mit/]
	// true
	// true
}

func TestSource(t *testing.T) {
	require.Equal(t, LicenseListSource, Source())
}

func TestVersion(t *testing.T) {
	type llv struct {
		LicenseListVersion string `json:"licenseListVersion"`
	}
	var r llv
	_ = json.Unmarshal(licensesJSON, &r)
	require.Equal(t, r.LicenseListVersion, Version())
	require.NotEmpty(t, Version())
}

func TestReleaseDate(t *testing.T) {
	type rdt struct {
		ReleaseDate string `json:"releaseDate"`
	}
	var r rdt
	_ = json.Unmarshal(licensesJSON, &r)
	require.Equal(t, r.ReleaseDate, ReleaseDate())
	require.NotEmpty(t, ReleaseDate())
}

func TestLookup(t *testing.T) {
	cases := []struct {
		spdxID string
	}{
		{"0BSD"},
		{"mIT"},
		{"Apache-2.0"},
		{"GPL-3.0-or-later"},
	}
	for _, c := range cases {
		t.Run(c.spdxID, func(t *testing.T) {
			lic := Lookup(c.spdxID)
			require.NotEqual(t, NoLicenseMatch, lic, "license not found - ensure the JSON file is up-to-date")
		})
	}
	t.Run("empty", func(t *testing.T) {
		require.Equal(t, NoLicenseMatch, Lookup(""))
	})
	t.Run("no match", func(t *testing.T) {
		require.Equal(t, NoLicenseMatch, Lookup("ThisLicenseDoesNotExistAndNeverWill87132d885b8426bc30278036249171deaaa510ad7101608f77dc499179b0fc0d"))
	})
}

func TestLicenses(t *testing.T) {
	lic := Licenses()
	require.NotEmpty(t, lic)
	require.Equal(t, len(lic), len(ll.lg.Licenses))

	dumpedKeys := make([]string, 0, len(ll.lg.Licenses))
	for _, l := range ll.lg.Licenses {
		dumpedKeys = append(dumpedKeys, l.LicenseID)
	}
	for _, l := range lic {
		require.Contains(t, dumpedKeys, l)
	}
}

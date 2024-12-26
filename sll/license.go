// SPDX License List (sll) provides a simple way to lookup SPDX License Identifiers and their details.
//
// The SPDX License List is a list of commonly found licenses and exceptions used in free and open-source software and
// can be found at https://spdx.org/licenses/.
//
// This package bundles the SPDX License List in JSON format and provides a simple way to lookup licenses by their
// SPDX License Identifier.
//
// Upon initialization, the package reads the SPDX License List JSON file and stores it in memory. The package provides
// a Lookup function to search for a license by its SPDX License Identifier.
//
// The current version of the SPDX License List included in this package can be found by referencing the included
// licenses.json file. At runtime, the SPDX License List version and release date can be retrieved using the [Version]
// and [ReleaseDate] functions.
//
// # License
//
// While the code in this package is licensed as per the LICENSE file, the SPDX License List JSON file is licensed
// under the Creative Commons Attribution 3.0 Unported License (CC-BY-3.0 - https://creativecommons.org/licenses/by/3.0/).
// The SPDX License List JSON file is was obtained from https://github.com/spdx/license-list-data and is maintained by
// the SPDX Legal Team (https://spdx.dev/engage/participate/legal-team/). The included SPDX License List JSON file is
// the copyright of The Linux Foundation.
//
// No warranties are provided by the SPDX Legal Team or the Linux Foundation.
//
// # Usage
//
// The package is kept simple and provides a single function to lookup licenses by their SPDX License Identifier.
//
// As a user of this package, all you have to do is to import the package and call the [Lookup] function with the SPDX
// License Identifier you want to lookup. The license list is initialized at package load time and is stored in memory.
//
// Please see the prvoided examples for more information.
//
// # Limitations
//
// This package does not provide the ability to lookup exceptions at this time.
package sll

import (
	_ "embed"
	"encoding/json"
	"strings"
)

//go:embed licenses.json
var licensesJSON []byte
var ll licenseListContainer
var NoLicenseMatch = License{}

const (
	LicenseListSource = "https://spdx.org/licenses/"
)

func init() {
	ll = licenseListContainer{
		source: LicenseListSource,
	}

	var r licenseList
	// error is ignored here, because if there's an error, it will be caught by tests.
	_ = json.Unmarshal(licensesJSON, &r)

	licenseIdentifiers := make([]string, 0, len(r.Licenses))
	m := make(map[string]*License, len(r.Licenses))
	for _, l := range r.Licenses {
		m[strings.ToLower(l.LicenseID)] = &l
		licenseIdentifiers = append(licenseIdentifiers, l.LicenseID)
	}

	ll.lg = r
	ll.m = m
	ll.licenseIdentifiers = licenseIdentifiers
}

type licenseList struct {
	LicenseListVersion string    `json:"licenseListVersion"`
	Licenses           []License `json:"licenses"`
	ReleaseDate        string    `json:"releaseDate"`
}

type License struct {
	Reference             string   `json:"reference"`
	IsDeprecatedLicenseID bool     `json:"isDeprecatedLicenseId"`
	DetailsURL            string   `json:"detailsUrl"`
	ReferenceNumber       int64    `json:"referenceNumber"`
	Name                  string   `json:"name"`
	LicenseID             string   `json:"licenseId"`
	SeeAlso               []string `json:"seeAlso"`
	IsOSIApproved         bool     `json:"isOsiApproved"`
	IsFSFLibre            *bool    `json:"isFsfLibre,omitempty"`
}

type licenseListContainer struct {
	lg                 licenseList
	source             string
	m                  map[string]*License
	licenseIdentifiers []string
}

// Lookup returns the License object that corresponds to the sid provided. The sid must be a valid
// SPDX License Identifier.
//
// If the sid is not found, the function returns NoLicenseMatch.
//
// # Case sensitivity
//
// The matching of SPDX License Identifier is case-insensitive as per the SPDX specification (Annex B).
func Lookup(sid string) License {
	if ll.m[strings.ToLower(sid)] == nil {
		return NoLicenseMatch
	}
	return *ll.m[strings.ToLower(sid)]
}

// ReleaseDate returns the release date of the SPDX License List referenced.
func ReleaseDate() string {
	return ll.lg.ReleaseDate
}

// Version returns the version of the SPDX License List referenced.
func Version() string {
	return ll.lg.LicenseListVersion
}

// Source returns the URL of the SPDX License List.
func Source() string {
	return LicenseListSource
}

// Licenses returns a list of all SPDX License Identifiers in the SPDX License List.
// The returned list is a copy of the internal list and can be modified without affecting the internal list.
func Licenses() []string {
	out := make([]string, len(ll.licenseIdentifiers))
	copy(out, ll.licenseIdentifiers)
	return out
}

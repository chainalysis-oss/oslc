package sll

import "github.com/chainalysis-oss/oslc"

// AsLicenseRetriever returns a [oslc.LicenseRetriever] that uses the SPDX license list.
func AsLicenseRetriever() oslc.LicenseRetriever {
	return &lr{}
}

type lr struct{}

// Lookup returns the SPDX license with the given identifier. Lookups are case-insensitive as per the SPDX
// specification.
func (l *lr) Lookup(id string) oslc.License {
	lic := Lookup(id)
	if lic.Name == "" {
		return oslc.License{}
	}
	return oslc.License{
		Name: lic.Name,
		ID:   lic.LicenseID,
	}
}

// ReleaseDate returns the release date of the SPDX license list.
func (l *lr) ReleaseDate() string {
	return ReleaseDate()
}

// Version returns the version of the SPDX license list.
func (l *lr) Version() string {
	return Version()
}

// Source returns the URL of the SPDX license list.
func (l *lr) Source() string {
	return Source()
}

// Licenses returns a list of all SPDX license identifiers in the SPDX license list.
func (l *lr) Licenses() []string {
	return Licenses()
}

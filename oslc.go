package oslc

import (
	"context"
	"errors"
	"fmt"
)

type Entry struct {
	Name               string              `json:"name"`
	DistributionPoints []DistributionPoint `json:"distribution_points,omitempty"`
	License            string              `json:"license"`
	Version            string              `json:"version"`
}

type DistributionPoint struct {
	Name        string `json:"name"`
	URL         string `json:"url"`
	Distributor string `json:"distributor"`
}

const (
	DistributorPypi     = "pypi"
	DistributorNpm      = "npm"
	DistributorMaven    = "maven"
	DistributorCratesIo = "crates.io"
)

type DatastoreSaver interface {
	Save(ctx context.Context, entry Entry) error
}

type DatastoreRetriever interface {
	Retrieve(ctx context.Context, name, version, distributor string) (Entry, error)
}

type Datastore interface {
	DatastoreSaver
	DatastoreRetriever
}

type DistributorClient interface {
	GetPackage(name string) (Entry, error)
	GetPackageVersion(name, version string) (Entry, error)
}

var ErrDatastoreObjectNotFound = errors.New("not found")

var ErrVersionNotFound = fmt.Errorf("version not found")

type License struct {
	Name string
	ID   string
}

type LicenseRetriever interface {
	// Lookup returns the License object that corresponds to the provided id. Case-sensitivity is
	// implementation-specific.
	Lookup(id string) License
	ReleaseDate() string
	Version() string
	Source() string
	// Licenses returns a list of keys for all licenses in the license list. Implementations must return a copy
	// of the internal list and not the internal list itself to prevent modifications to the internal list.
	Licenses() []string
}

type LicenseIDNormalizer interface {
	// NormalizeID returns the normalized ID that corresponds to the provided id.
	//
	// Implementations must return an empty string if the provided id is not found in the license list.
	//
	// The following details are implementation-specific:
	// - Case-sensitivity of the id
	// - Matching of the id to the license list
	// - Normalization of the id to a canonical form
	// - List of available licenses
	//
	// Implementations should document these implementation-specific details.
	NormalizeID(ctx context.Context, id string) string
}

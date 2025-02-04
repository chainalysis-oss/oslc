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
	DistributorGo       = "go"
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

var ErrNoSuchPackage = errors.New("no such package")

// DistributorClient is an interface that represents a client that can communicate with a distributor.
//
// Errors from the distributor must be returned as a [DistributorError]. The distributor name must be set to the
// distributor's name. The format of the name is implementation-specific. The [DistributorError] will ensure the
// underlying error is not exposed to the caller. Exceptions to this rule are errors indicating that a specific
// package or version is not found. These errors must be returned as [ErrNoSuchPackage] and [ErrVersionNotFound],
// respectively.
//
// GetPackage returns the [Entry] object that corresponds to the provided name. If the package is not found,
// the implementation must return [ErrNoSuchPackage]. If a package version is not found, the implementation
// must return [ErrVersionNotFound]. If an error occurs while communicating with the distributor, the implementation
// must return a [DistributorError] with the distributor name set to the distributor's name. The format of the name
// is implementation-specific.
//
// GetPackageVersion returns the [Entry] object that corresponds to the provided name and version. If the package
// version is not found, the implementation must return [ErrVersionNotFound]. If the package is not found, the
// implementation must return [ErrNoSuchPackage]. If an error occurs while communicating
// with the distributor, the implementation must return a [DistributorError] with the distributor name set to the
// distributor's name. The format of the name and version is implementation-specific.
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

// DistributorError is an error type that represents an error that occurred in communicating with a distributor.
type DistributorError struct {
	Distributor string
	Err         error
}

func (e DistributorError) Error() string {
	if e.Err == nil {
		return fmt.Sprintf("unspecified error communicating with %s", e.Distributor)
	}

	return fmt.Sprintf("error communicating with %s: %s", e.Distributor, e.Err)
}

// Unwrap returns the underlying error of the DistributorError so that the caller can inspect the error. If the
// underlying error is not one of the following errors, Unwrap will return nil:
// - [ErrNoSuchPackage]
// - [ErrVersionNotFound]
//
// This allows us to hide implementation-specific errors from the caller, yet allow the caller to use [errors.Is] to
// determine if the DistributorError is caused by specific errors.
func (e DistributorError) Unwrap() error {
	if e.Err == nil {
		return nil
	}
	if errors.Is(e.Err, ErrNoSuchPackage) || errors.Is(e.Err, ErrVersionNotFound) {
		return e.Err
	}
	return nil
}

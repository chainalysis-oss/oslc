package cratesio

import (
	"encoding/json"
	"fmt"
	"github.com/chainalysis-oss/oslc"
	ownHTTP "github.com/chainalysis-oss/oslc/http"
	"io"
	"net/http"
	"strings"
)

var cratesIOBaseURL = "https://crates.io"

// crateVersionResponse is the response from the crates.io API for the `/api/v1/crates/{name}/{version}` endpoint.
type crateVersionResponse struct {
	Version crateVersion `json:"version"`
}

func (c crateVersionResponse) AsEntry() oslc.Entry {
	return c.Version.AsEntry()
}

type crateVersion struct {
	Crate   string `json:"crate"`
	Num     string `json:"num"`
	License string `json:"license"`
	Links   struct {
		VersionDownloads string `json:"version_downloads"`
	}
}

func (p crateVersion) AsEntry() oslc.Entry {
	entry := oslc.Entry{
		Name: p.Crate,
	}
	if p.Links.VersionDownloads != "" {
		entry.DistributionPoints = []oslc.DistributionPoint{{
			Name:        p.Crate,
			URL:         fmt.Sprintf("%s%s", cratesIOBaseURL, p.Links.VersionDownloads),
			Distributor: oslc.DistributorCratesIo,
		}}
	}
	if p.Num != "" {
		entry.Version = p.Num
	}
	if p.License != "" {
		entry.License = p.License
	} else {
		entry.License = "Unknown"
	}
	return entry
}

type crate struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	NewestVersion string `json:"newest_version"`
}

// crateResponse is the response from the crates.io API for the `/api/v1/crates/{name}` endpoint.
type crateResponse struct {
	Crate    crate          `json:"crate"`
	Versions []crateVersion `json:"versions"`
}

func (c crateResponse) newestVersion() (crateVersion, error) {
	if len(c.Versions) == 0 {
		return crateVersion{}, fmt.Errorf("%w: no versions in crate", oslc.ErrVersionNotFound)
	}
	if c.Crate.NewestVersion == "" {
		return crateVersion{}, fmt.Errorf("%w: crate has no newest version", oslc.ErrVersionNotFound)
	}

	for _, v := range c.Versions {
		if v.Num == c.Crate.NewestVersion {
			return v, nil
		}
	}
	return crateVersion{}, fmt.Errorf("%w: crate specifies a version not included in the crate", oslc.ErrVersionNotFound)
}

// Client is a client for the Crates.io API.
// It should never be created directly, use [NewClient] instead.
type Client struct {
	options *clientOptions
}

func NewClient(options ...ClientOption) (*Client, error) {
	opts := defaultClientOptions
	for _, opt := range globalClientOptions {
		opt.apply(&opts)
	}
	for _, opt := range options {
		opt.apply(&opts)
	}

	if opts.HttpClient == nil {
		c, _ := ownHTTP.NewClient(ownHTTP.WithLogger(opts.Logger), ownHTTP.WithHeaders(http.Header{
			"Accept": {"application/json"},
		}))
		opts.HttpClient = c
	}
	return &Client{
		options: &opts,
	}, nil
}

// GetPackageVersion returns the package with the given name and version.
// If version is empty, the latest version is returned.
func (c *Client) GetPackageVersion(name, version string) (oslc.Entry, error) {
	path := fmt.Sprintf("api/v1/crates/%s/%s", name, version)
	if version == "" {
		path = fmt.Sprintf("api/v1/crates/%s", name)
	}

	resp, err := c.options.HttpClient.Query(fmt.Sprintf("%s/%s", c.options.BaseURL, path))
	if err != nil {
		return oslc.Entry{}, oslc.DistributorError{Distributor: oslc.DistributorCratesIo, Err: err}
	}

	if resp.StatusCode == http.StatusNotFound {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 400))
		defer resp.Body.Close()
		if strings.Contains(string(body), "does not have a version") {
			return oslc.Entry{}, oslc.DistributorError{Distributor: oslc.DistributorCratesIo, Err: fmt.Errorf("%w: %s", oslc.ErrVersionNotFound, name)}
		}
		return oslc.Entry{}, oslc.DistributorError{Distributor: oslc.DistributorCratesIo, Err: fmt.Errorf("%w: %s", oslc.ErrNoSuchPackage, name)}
	}

	if resp.StatusCode != http.StatusOK {
		return oslc.Entry{}, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	defer resp.Body.Close()

	if version == "" {
		var crt crateResponse
		err = json.NewDecoder(resp.Body).Decode(&crt)
		if err != nil {
			return oslc.Entry{}, oslc.DistributorError{Distributor: oslc.DistributorCratesIo, Err: err}
		}
		pkg, err := crt.newestVersion()
		if err != nil {
			return oslc.Entry{}, oslc.DistributorError{Distributor: oslc.DistributorCratesIo, Err: err}
		}
		return pkg.AsEntry(), nil
	} else {
		var pkg crateVersionResponse
		err = json.NewDecoder(resp.Body).Decode(&pkg)
		if err != nil {
			return oslc.Entry{}, oslc.DistributorError{Distributor: oslc.DistributorCratesIo, Err: err}
		}
		return pkg.AsEntry(), nil
	}
}

// GetPackage returns the package with the given name. It is a convenience function for [GetPackageVersion]
// with an empty version.
func (c *Client) GetPackage(name string) (oslc.Entry, error) {
	return c.GetPackageVersion(name, "")
}

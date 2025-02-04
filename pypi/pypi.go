package pypi

import (
	"encoding/json"
	"fmt"
	"github.com/chainalysis-oss/oslc"
	ownHTTP "github.com/chainalysis-oss/oslc/http"
	"net/http"
)

// pypiPackageResponse is a Python package. It is the result of a query to the PyPI API as specified in the PyPI API documentation
// here: https://warehouse.pypa.io/api-reference/json.html#get--pypi--project_name--json
type pypiPackageResponse struct {
	Info struct {
		Name        string `json:"name"`
		License     string `json:"license"`
		PackageURL  string `json:"package_url"`
		ProjectURLs struct {
			Source string `json:"Source"`
		} `json:"project_urls"`
		Version string `json:"version"`
	} `json:"info"`
	Releases map[string][]struct{} `json:"releases"`
}

func (p pypiPackageResponse) AsEntry() oslc.Entry {
	entry := oslc.Entry{
		Name: p.Info.Name,
	}
	if p.Info.PackageURL != "" {
		entry.DistributionPoints = []oslc.DistributionPoint{{
			Name:        p.Info.Name,
			URL:         p.Info.PackageURL,
			Distributor: oslc.DistributorPypi,
		}}
	}
	if p.Info.Version != "" {
		entry.Version = p.Info.Version
	}
	if p.Info.License != "" {
		entry.License = p.Info.License
	} else {
		entry.License = "Unknown"
	}
	return entry
}

// Client is a client for the PyPI API.
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
	path := fmt.Sprintf("pypi/%s/%s/json", name, version)
	if version == "" {
		path = fmt.Sprintf("pypi/%s/json", name)
	}

	resp, err := c.options.HttpClient.Query(fmt.Sprintf("%s/%s", c.options.BaseURL, path))
	if err != nil {
		return oslc.Entry{}, oslc.DistributorError{Distributor: oslc.DistributorPypi, Err: err}
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		ok, err := c.packageExists(name)
		if err != nil {
			return oslc.Entry{}, oslc.DistributorError{Distributor: oslc.DistributorPypi, Err: err}
		}
		if !ok {
			return oslc.Entry{}, oslc.DistributorError{Distributor: oslc.DistributorPypi, Err: fmt.Errorf("%w: %s", oslc.ErrNoSuchPackage, name)}
		}
		return oslc.Entry{}, oslc.DistributorError{Distributor: oslc.DistributorPypi, Err: fmt.Errorf("%w: %s", oslc.ErrVersionNotFound, version)}
	}

	if resp.StatusCode != http.StatusOK {
		return oslc.Entry{}, oslc.DistributorError{Distributor: oslc.DistributorPypi, Err: fmt.Errorf("unexpected status code: %d", resp.StatusCode)}
	}

	var pkg pypiPackageResponse
	err = json.NewDecoder(resp.Body).Decode(&pkg)
	if err != nil {
		return pkg.AsEntry(), oslc.DistributorError{Distributor: oslc.DistributorPypi, Err: err}
	}
	return pkg.AsEntry(), nil
}

// GetPackage returns the package with the given name. It is a convenience function for [GetPackageVersion]
// with an empty version.
func (c *Client) GetPackage(name string) (oslc.Entry, error) {
	return c.GetPackageVersion(name, "")
}

func (c *Client) packageExists(name string) (bool, error) {
	resp, err := c.options.HttpClient.Query(fmt.Sprintf("%s/pypi/%s/json", c.options.BaseURL, name))
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusOK {
		return true, nil
	}
	if resp.StatusCode == http.StatusNotFound {
		return false, nil
	}
	return false, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
}

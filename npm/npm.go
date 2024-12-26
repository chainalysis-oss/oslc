package npm

import (
	"encoding/json"
	"fmt"
	"github.com/chainalysis-oss/oslc"
	ownHTTP "github.com/chainalysis-oss/oslc/http"
	"net/http"
)

// npmPackageResponse is an NPM package. It is the result of a query to the NPM API as specified in the NPM API documentation
// here: https://github.com/npm/registry/blob/main/docs/responses/package-metadata.md
type npmPackageResponse struct {
	Name     string `json:"name"`
	Version  string `json:"version"`
	License  string `json:"license"`
	DistTags struct {
		Latest string `json:"latest"`
	} `json:"dist-tags"`
	Repository struct {
		Type string `json:"type"`
		URL  string `json:"url"`
	} `json:"repository"`
	Versions map[string]struct {
		Version    string `json:"version"`
		License    string `json:"license"`
		Repository struct {
			Type string `json:"type"`
			URL  string `json:"url"`
		} `json:"repository"`
	} `json:"versions"`
}

func (p npmPackageResponse) AsEntry() oslc.Entry {
	entry := oslc.Entry{
		Name: p.Name,
	}
	// Set package URL
	entry.DistributionPoints = []oslc.DistributionPoint{{
		Name:        p.Name,
		URL:         fmt.Sprintf("https://www.npmjs.com/package/%s", p.Name),
		Distributor: oslc.DistributorNpm,
	}}

	if p.Version != "" {
		entry.Version = p.Version
	} else if p.Versions[p.DistTags.Latest].Version != "" {
		entry.Version = p.Versions[p.DistTags.Latest].Version
	} else {
		entry.Version = "Unknown"
	}

	if p.License != "" {
		entry.License = p.License
	} else if p.Versions[p.DistTags.Latest].License != "" {
		entry.License = p.Versions[p.DistTags.Latest].License
	} else {
		entry.License = "Unknown"
	}

	return entry
}

// Client is a client for the NPM API.
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
		c, _ := ownHTTP.NewClient(ownHTTP.WithLogger(opts.Logger))
		opts.HttpClient = c
	}
	return &Client{
		options: &opts,
	}, nil
}

// GetPackageVersion returns the package with the given name and version.
// If version is empty, the latest version is returned.
func (c *Client) GetPackageVersion(name, version string) (oslc.Entry, error) {
	path := fmt.Sprintf("%s/%s", name, version)
	if version == "" {
		path = fmt.Sprintf("%s", name)
	}

	resp, err := c.options.HttpClient.Query(fmt.Sprintf("%s/%s", c.options.BaseURL, path))
	if err != nil {
		return oslc.Entry{}, err
	}

	if resp.StatusCode != http.StatusOK {
		return oslc.Entry{}, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	defer resp.Body.Close()

	var pkg npmPackageResponse
	err = json.NewDecoder(resp.Body).Decode(&pkg)
	if err != nil {
		return pkg.AsEntry(), err
	}
	return pkg.AsEntry(), nil
}

// GetPackage returns the package with the given name. It is a convenience function for [GetPackageVersion]
// with an empty version.
func (c *Client) GetPackage(name string) (oslc.Entry, error) {
	return c.GetPackageVersion(name, "")
}

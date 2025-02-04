package maven

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"github.com/chainalysis-oss/oslc"
	ownHTTP "github.com/chainalysis-oss/oslc/http"
	"net/http"
	"strings"
)

type mavenPOM struct {
	Licenses []struct {
		License struct {
			Name string `xml:"name"`
		} `xml:"license"`
	} `xml:"licenses"`
	GroupId    string `xml:"groupId"`
	ArtifactId string `xml:"artifactId"`

	Version string `xml:"version"`
}

func (p mavenPOM) AsEntry() oslc.Entry {
	name := fmt.Sprintf("%s:%s", p.GroupId, p.ArtifactId)
	entry := oslc.Entry{
		Name: name,
	}
	// Set package URL
	entry.DistributionPoints = []oslc.DistributionPoint{{
		Name:        name,
		URL:         fmt.Sprintf("https://central.sonatype.com/artifact/%s/%s", p.GroupId, p.ArtifactId),
		Distributor: oslc.DistributorMaven,
	}}

	if p.Version != "" {
		entry.Version = p.Version
	} else {
		entry.Version = "Unknown"
	}

	if p.Licenses != nil {
		entry.License = p.Licenses[0].License.Name
	} else {
		entry.License = "Unknown"
	}

	return entry
}

// Client is a client for the Maven API.
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

func nameIsValid(name string) bool {
	s := strings.Split(name, ":")
	if len(s) != 2 {
		return false
	}
	if s[0] == "" || s[1] == "" {
		return false
	}
	return true
}

// GetPackageVersion returns the package with the given name and version.
// If version is empty, the latest version is returned.
func (c *Client) GetPackageVersion(name, version string) (oslc.Entry, error) {
	if !nameIsValid(name) {
		return oslc.Entry{}, fmt.Errorf("%w: %s", oslc.ErrNoSuchPackage, name)
	}
	if version == "latest" {
		version = ""
	}

	var groupId string
	var artifactId string
	var err error
	groupId = strings.Split(name, ":")[0]
	// When querying this endpoint, the dots separating the segments of the groupId need to be replaced with slashes.
	normGroupId := strings.ReplaceAll(groupId, ".", "/")
	artifactId = strings.Split(name, ":")[1]
	if version == "" {
		version, err = c.getLatestVersion(groupId, artifactId)
		if err != nil {
			return oslc.Entry{}, oslc.DistributorError{Distributor: oslc.DistributorMaven, Err: err}
		}
	}
	path := fmt.Sprintf("remotecontent?filepath=%s/%s/%s/%s-%s.pom", normGroupId, artifactId, version, artifactId, version)
	resp, err := c.options.HttpClient.Query(fmt.Sprintf("%s/%s", c.options.BaseURL, path))
	if err != nil {
		return oslc.Entry{}, oslc.DistributorError{Distributor: oslc.DistributorMaven, Err: err}
	}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		ok, err := c.doesPackageExist(groupId, artifactId)
		if err != nil {
			return oslc.Entry{}, oslc.DistributorError{Distributor: oslc.DistributorMaven, Err: err}
		}
		if !ok {
			return oslc.Entry{}, oslc.DistributorError{Distributor: oslc.DistributorMaven, Err: fmt.Errorf("%w: %s", oslc.ErrNoSuchPackage, name)}
		}
		return oslc.Entry{}, oslc.DistributorError{Distributor: oslc.DistributorMaven, Err: fmt.Errorf("%w: %s", oslc.ErrVersionNotFound, version)}
	}

	var pkg mavenPOM
	err = xml.NewDecoder(resp.Body).Decode(&pkg)
	if err != nil {
		return pkg.AsEntry(), oslc.DistributorError{Distributor: oslc.DistributorMaven, Err: err}
	}
	return pkg.AsEntry(), nil
}

// GetPackage returns the package with the given name. It is a convenience function for [GetPackageVersion]
// with an empty version.
func (c *Client) GetPackage(name string) (oslc.Entry, error) {
	return c.GetPackageVersion(name, "")
}

type solrResponse struct {
	ResponseHeader struct {
		Status int `json:"status"`
	} `json:"responseHeader"`
	Response struct {
		NumFound int `json:"numFound"`
		Docs     []struct {
			Id            string `json:"id"`
			LatestVersion string `json:"latestVersion"`
		} `json:"docs"`
	} `json:"response"`
}

func (c *Client) doesPackageExist(groupId, artifactId string) (bool, error) {
	path := fmt.Sprintf("solrsearch/select?q=g:%s+AND+a:%s&rows=1&wt=json", groupId, artifactId)
	resp, err := c.options.HttpClient.Query(fmt.Sprintf("%s/%s", c.options.BaseURL, path))
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return false, fmt.Errorf("error determining if package exists: unexpected status code: %d", resp.StatusCode)
	}

	var pkg solrResponse
	err = json.NewDecoder(resp.Body).Decode(&pkg)
	if err != nil {
		return false, err
	}
	if len(pkg.Response.Docs) == 0 {
		return false, nil
	}
	return true, nil
}

func (c *Client) getLatestVersion(groupId, artifactId string) (string, error) {
	path := fmt.Sprintf("solrsearch/select?q=g:%s+AND+a:%s&rows=1&wt=json", groupId, artifactId)
	resp, err := c.options.HttpClient.Query(fmt.Sprintf("%s/%s", c.options.BaseURL, path))
	if err != nil {
		return "", err
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("error getting latest version: unexpected status code: %d", resp.StatusCode)
	}

	defer resp.Body.Close()

	var pkg solrResponse
	err = json.NewDecoder(resp.Body).Decode(&pkg)
	if err != nil {
		return "", err
	}
	if len(pkg.Response.Docs) == 0 {
		return "", fmt.Errorf("%w: %s", oslc.ErrNoSuchPackage, fmt.Sprintf("%s:%s", groupId, artifactId))
	}
	return pkg.Response.Docs[0].LatestVersion, nil
}

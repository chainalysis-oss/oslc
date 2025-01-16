package goproxy

import (
	"archive/zip"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/chainalysis-oss/oslc"
	ownHTTP "github.com/chainalysis-oss/oslc/http"
	"github.com/go-enry/go-license-detector/v4/licensedb"
	"github.com/go-enry/go-license-detector/v4/licensedb/filer"
	"io"
	"math/rand/v2"
	"net/http"
	"os"
	"path"
	"strconv"
)

func init() {
	licensedb.Preload()
}

const goProxyBaseURL string = "https://proxy.golang.org"

type Client struct {
	options *clientOptions
}

func (c *Client) GetPackage(name string) (oslc.Entry, error) {
	return c.GetPackageVersion(name, "")
}

func (c *Client) GetPackageVersion(name, version string) (oslc.Entry, error) {
	vi, err := c.getInfo(name, version)
	if err != nil {
		return oslc.Entry{}, err
	}
	license, err := c.getLicense(name, vi.Version)
	if err != nil {
		return oslc.Entry{}, err
	}
	return oslc.Entry{
		Name:    name,
		Version: vi.Version,
		License: license,
		DistributionPoints: []oslc.DistributionPoint{
			{
				Name:        name,
				URL:         fmt.Sprintf("%s/%s/@v/%s.zip", c.options.BaseURL, name, vi.Version),
				Distributor: oslc.DistributorGo,
			},
		},
	}, nil
}

func (c *Client) getLicense(name, version string) (string, error) {
	if version == "" {
		return "", fmt.Errorf("version is empty")
	}
	resp, err := c.options.HttpClient.Query(c.options.BaseURL + "/" + name + "/@v/" + version + ".zip")
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	err = os.MkdirAll(c.options.TempDir, os.ModePerm)
	if err != nil {
		return "", err
	}

	fp := path.Join(c.options.TempDir, strconv.Itoa(rand.Int())+".zip")
	f, err := os.Create(fp)
	if err != nil {
		return "", err
	}
	defer f.Close()
	io.Copy(f, resp.Body)

	fpu := path.Join(c.options.TempDir, strconv.Itoa(rand.Int()))
	err = unzip(fp, fpu)
	if err != nil {
		return "", err
	}

	filer, err := filer.FromDirectory(path.Join(fpu, name+"@"+version))
	if err != nil {
		return "", err
	}
	matches, err := licensedb.Detect(filer)
	if err != nil {
		return "", err
	}
	if len(matches) == 0 {
		return "", fmt.Errorf("no license detected")
	}
	for lic, match := range matches {
		if match.Confidence == 1 {
			return lic, nil
		}
	}

	return "", fmt.Errorf("no license detected")

}

// NewClient creates a new client.
func NewClient(options ...ClientOption) (*Client, error) {
	opts := defaultClientOptions
	for _, opt := range GlobalClientOptions {
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

type versionOrigin struct {
	VCS  string
	URL  string
	Hash string
	Ref  string
}

type versionInfo struct {
	Origin  versionOrigin
	Version string
	Time    string
}

// getInfo returns information about the version of the package.
func (c *Client) getInfo(name, version string) (versionInfo, error) {
	var err error
	var resp *http.Response
	if version == "" {
		resp, err = c.options.HttpClient.Query(c.options.BaseURL + "/" + name + "/@latest")
	} else {
		resp, err = c.options.HttpClient.Query(c.options.BaseURL + "/" + name + "/@v/" + version + ".info")
	}
	if err != nil {
		return versionInfo{}, newDistributorError(fmt.Errorf("constructing HTTP query for upstream '%s': %s", c.options.BaseURL, err))
	}
	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode == http.StatusNotFound {
			return versionInfo{}, noSuchModuleOrVersionError{
				Upstream: c.options.BaseURL,
				Module:   name,
				Version:  version,
			}
		}
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 400)) // actual number pulled out of thin air
		return versionInfo{}, newDistributorError(fmt.Errorf("non-200 status code getting version '%s' of module '%s' from upstream '%s': %d - Body (base64): %s", version, name, c.options.BaseURL, resp.StatusCode, base64.StdEncoding.EncodeToString(body)))
	}
	defer resp.Body.Close()
	vi := versionInfo{}
	dec := json.NewDecoder(resp.Body)
	err = dec.Decode(&vi)
	if err != nil {
		return versionInfo{}, newDistributorError(fmt.Errorf("decoding JSON response from upstream '%s': %s", c.options.BaseURL, err))
	}

	return vi, nil
}

type noSuchModuleOrVersionError struct {
	Upstream string
	Module   string
	Version  string
}

func (e noSuchModuleOrVersionError) Error() string {
	return fmt.Sprintf("unable to find version '%s' of module '%s' in upstream '%s'", e.Version, e.Module, e.Upstream)
}

func newDistributorError(err error) oslc.DistributorError {
	return oslc.DistributorError{
		Distributor: oslc.DistributorGo,
		Err:         err,
	}
}

func unzip(source string, target string) error {
	r, err := zip.OpenReader(source)
	if err != nil {
		return fmt.Errorf("impossible to open zip file: %s", err)
	}
	defer r.Close()

	for k, f := range r.File {
		rc, err := f.Open()
		if err != nil {
			return fmt.Errorf("impossible to open file n°%d in archive: %w", k, err)
		}
		defer rc.Close()
		newFilePath := path.Join(target, f.Name)

		if f.FileInfo().IsDir() {
			err = os.MkdirAll(newFilePath, 0777)
			if err != nil {
				return fmt.Errorf("impossible to MkdirAll: %w", err)
			}
			continue
		}

		err = os.MkdirAll(path.Dir(newFilePath), 0777)
		if err != nil {
			return fmt.Errorf("impossible to MkdirAll: %w", err)
		}
		uncompressedFile, err := os.Create(newFilePath)
		if err != nil {
			return fmt.Errorf("impossible to create uncompressed: %w", err)
		}
		_, err = io.Copy(uncompressedFile, rc)
		if err != nil {
			return fmt.Errorf("impossible to copy file n°%d: %w", k, err)
		}
	}

	return nil
}

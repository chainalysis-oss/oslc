package oslc

import (
	"github.com/chainalysis-oss/oslc"
	"log/slog"
)

type serverOptions struct {
	Logger              *slog.Logger
	PypiClient          oslc.DistributorClient
	NpmClient           oslc.DistributorClient
	MavenClient         oslc.DistributorClient
	CratesIoClient      oslc.DistributorClient
	GoClient            oslc.DistributorClient
	Datastore           oslc.Datastore
	LicenseIDNormalizer oslc.LicenseIDNormalizer
}

var defaultServerOptions = serverOptions{
	Logger: slog.Default(),
}

var globalServerOptions []ServerOption

// ServerOption is an option for configuring a Client.
type ServerOption interface {
	apply(*serverOptions)
}

// funcServerOption is a ServerOption that calls a function.
// It is used to wrap a function, so it satisfies the ServerOption interface.
type funcServerOption struct {
	f func(*serverOptions)
}

func (fdo *funcServerOption) apply(opts *serverOptions) {
	fdo.f(opts)
}

func newFuncClientOption(f func(*serverOptions)) *funcServerOption {
	return &funcServerOption{
		f: f,
	}
}

// WithLogger returns a ServerOption that uses the provided logger.
func WithLogger(logger *slog.Logger) ServerOption {
	return newFuncClientOption(func(opts *serverOptions) {
		opts.Logger = logger
	})
}

// WithPypiClient returns a ServerOption that uses the provided Pypi client.
func WithPypiClient(c oslc.DistributorClient) ServerOption {
	return newFuncClientOption(func(opts *serverOptions) {
		opts.PypiClient = c
	})
}

// WithNpmClient returns a ServerOption that uses the provided Npm client.
func WithNpmClient(c oslc.DistributorClient) ServerOption {
	return newFuncClientOption(func(opts *serverOptions) {
		opts.NpmClient = c
	})
}

// WithMavenClient returns a ServerOption that uses the provided Maven client.
func WithMavenClient(c oslc.DistributorClient) ServerOption {
	return newFuncClientOption(func(opts *serverOptions) {
		opts.MavenClient = c
	})
}

// WithDatastore returns a ServerOption that uses the provided Datastore.
func WithDatastore(d oslc.Datastore) ServerOption {
	return newFuncClientOption(func(opts *serverOptions) {
		opts.Datastore = d
	})
}

// WithLicenseIDNormalizer returns a ServerOption that uses the provided LicenseIDNormalizer.
func WithLicenseIDNormalizer(l oslc.LicenseIDNormalizer) ServerOption {
	return newFuncClientOption(func(opts *serverOptions) {
		opts.LicenseIDNormalizer = l
	})
}

// WithCratesIoClient returns a ServerOption that uses the provided Crates.io client.
func WithCratesIoClient(c oslc.DistributorClient) ServerOption {
	return newFuncClientOption(func(opts *serverOptions) {
		opts.CratesIoClient = c
	})
}

// WithGoClient returns a ServerOption that uses the provided Go client.
func WithGoClient(c oslc.DistributorClient) ServerOption {
	return newFuncClientOption(func(opts *serverOptions) {
		opts.GoClient = c
	})
}

package spdxnormalizer

import (
	"github.com/chainalysis-oss/oslc"
	"log/slog"
)

type normalizerOptions struct {
	Logger           *slog.Logger
	LicenseRetriever oslc.LicenseRetriever
}

var defaultNormalizerOptions = normalizerOptions{
	Logger: slog.Default(),
}

var globalNormalizerOptions []NormalizerOption

// NormalizerOption is an option for configuring a Client.
type NormalizerOption interface {
	apply(*normalizerOptions)
}

// funcNormalizerOption is a NormalizerOption that calls a function.
// It is used to wrap a function, so it satisfies the NormalizerOption interface.
type funcNormalizerOption struct {
	f func(*normalizerOptions)
}

func (fdo *funcNormalizerOption) apply(opts *normalizerOptions) {
	fdo.f(opts)
}

func newFuncNormalizerOption(f func(*normalizerOptions)) *funcNormalizerOption {
	return &funcNormalizerOption{
		f: f,
	}
}

// WithLicenseRetriever returns a NormalizerOption that uses the provided license database. While any
// LicenseRetriever can be used, it is recommended that the license retriever adheres to the SPDX specification, as
// the normalizer is designed to work with SPDX license identifiers.
func WithLicenseRetriever(ldb oslc.LicenseRetriever) NormalizerOption {
	return newFuncNormalizerOption(func(opts *normalizerOptions) {
		opts.LicenseRetriever = ldb
	})
}

// WithLogger returns a NormalizerOption that uses the provided logger.
func WithLogger(logger *slog.Logger) NormalizerOption {
	return newFuncNormalizerOption(func(opts *normalizerOptions) {
		opts.Logger = logger
	})
}

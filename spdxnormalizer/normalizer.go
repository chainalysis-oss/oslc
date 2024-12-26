package spdxnormalizer

import (
	"context"
	"github.com/chainalysis-oss/oslc"
)

// Compile time check to ensure Normalizer implements [oslc.LicenseIDNormalizer].
var _ oslc.LicenseIDNormalizer = (*Normalizer)(nil)

// Normalizer implements the [oslc.LicenseIDNormalizer] interface.
//
// The normalizer normalizes SPDX license identifiers to the corresponding license object, and as such it is
// recommended that the [WithLicenseRetriever] option is used and is provided with a license retriever that implements
// the SPDX license list. The [sll] package provides a license retriever that can be used with this normalizer.
//
// The documentation for this object and its methods are written in the context of SPDX license identifiers, and as such
// expects a LicenseRetriever that adheres to the SPDX specification and contains a list of SPDX licenses.
type Normalizer struct {
	options *normalizerOptions
}

// NewNormalizer creates a new Normalizer instance with the provided options.
func NewNormalizer(options ...NormalizerOption) (*Normalizer, error) {
	opts := defaultNormalizerOptions
	for _, opt := range globalNormalizerOptions {
		opt.apply(&opts)
	}
	for _, opt := range options {
		opt.apply(&opts)
	}

	return &Normalizer{
		options: &opts,
	}, nil
}

// NormalizeID normalizes the provided id to the corresponding SPDX License Identifier.
//
// The method uses the [oslc.LicenseRetriever] provided in the options to look up the SPDX License Identifier and
// returns the normalized SPDX License Identifier. If the provided id is not found in the license list, the method
// returns an empty string.
//
// The logic for normalization is handed off to the [oslc.LicenseRetriever]'s Lookup method. This means the following
// details are dependent on the implementation of the [oslc.LicenseRetriever]:
// - Case-sensitivity of the id
// - The normalization process
// - The mapping of the id to the SPDX License Identifier
// - The list of available licenses.
//
// However, the method is written in the context of SPDX License Identifiers and expects the [oslc.LicenseRetriever] to
// adhere to the SPDX specification and contain a list of SPDX licenses.
func (n *Normalizer) NormalizeID(ctx context.Context, id string) string {
	n.options.Logger.DebugContext(ctx, "normalizing license id", "id", id)
	norm := n.options.LicenseRetriever.Lookup(id).ID
	n.options.Logger.DebugContext(ctx, "normalized license id", "id", id, "normalized", norm, "success", norm != "")
	return norm
}

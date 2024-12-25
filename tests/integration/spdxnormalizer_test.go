package integration

import (
	"github.com/chainalysis-oss/oslc/sll"
	"github.com/chainalysis-oss/oslc/spdxnormalizer"
	"github.com/stretchr/testify/require"
	"io"
	"log/slog"
	"testing"
)

func TestSpdxNormalizer(t *testing.T) {
	licenseRetriever := sll.AsLicenseRetriever()
	normalizer, err := spdxnormalizer.NewNormalizer(
		spdxnormalizer.WithLogger(slog.New(slog.NewTextHandler(io.Discard, nil))),
		spdxnormalizer.WithLicenseRetriever(licenseRetriever),
	)
	require.NoError(t, err)

	license := normalizer.NormalizeID(nil, "mit")
	require.NotEqual(t, "", license)
	require.Equal(t, "MIT", license)
}

package spdxnormalizer

import (
	"github.com/chainalysis-oss/oslc"
	oslcmocks "github.com/chainalysis-oss/oslc/mocks/oslc"
	"github.com/stretchr/testify/require"
	"io"
	"log/slog"
	"testing"
)

func TestNormalizer_NormalizeID(t *testing.T) {
	mockLR := oslcmocks.NewMockLicenseRetriever(t)
	nm := &Normalizer{
		options: &normalizerOptions{
			Logger:           slog.New(slog.NewTextHandler(io.Discard, nil)),
			LicenseRetriever: mockLR,
		},
	}
	mockLR.EXPECT().Lookup("mit").Return(oslc.License{ID: "MIT"})

	out := nm.NormalizeID(nil, "mit")

	require.Equal(t, "MIT", out)
}

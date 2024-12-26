package spdxnormalizer

import (
	oslcMocks "github.com/chainalysis-oss/oslc/mocks/oslc"
	"github.com/stretchr/testify/require"
	"io"
	"log/slog"
	"testing"
)

func TestNewNormalizer(t *testing.T) {
	normalizer, err := NewNormalizer(WithLogger(slog.New(slog.NewTextHandler(io.Discard, nil))))
	require.NoError(t, err)
	require.NotNil(t, normalizer)
}

func TestNewNormalizer_globalOptionsAreApplied(t *testing.T) {
	optCopy := make([]NormalizerOption, len(globalNormalizerOptions))
	copy(optCopy, globalNormalizerOptions)
	defer func() {
		globalNormalizerOptions = optCopy
	}()

	globalNormalizerOptions = append(globalNormalizerOptions, WithLogger(slog.New(slog.NewTextHandler(io.Discard, nil))))
	normalizer, err := NewNormalizer()
	require.NoError(t, err)
	require.Equal(t, slog.New(slog.NewTextHandler(io.Discard, nil)), normalizer.options.Logger)
}

func TestFuncClientOption_apply(t *testing.T) {
	opts := normalizerOptions{}
	fdo := newFuncNormalizerOption(func(o *normalizerOptions) {
		o.Logger = slog.New(slog.NewTextHandler(io.Discard, nil))
	})
	fdo.apply(&opts)
	require.Equal(t, slog.New(slog.NewTextHandler(io.Discard, nil)), opts.Logger)
}

func TestNewFuncClientOption(t *testing.T) {
	fdo := newFuncNormalizerOption(func(o *normalizerOptions) {
		o.Logger = slog.New(slog.NewTextHandler(io.Discard, nil))
	})
	require.NotNil(t, fdo)
}

func TestWithLogger(t *testing.T) {
	opts := normalizerOptions{}
	WithLogger(slog.New(slog.NewTextHandler(io.Discard, nil))).apply(&opts)
	require.Equal(t, slog.New(slog.NewTextHandler(io.Discard, nil)), opts.Logger)
}

func TestWithLicenseRetriever(t *testing.T) {
	mock := oslcMocks.NewMockLicenseRetriever(t)
	opts := normalizerOptions{}
	WithLicenseRetriever(mock).apply(&opts)
	require.Equal(t, mock, opts.LicenseRetriever)
}

package main

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestNewListeners(t *testing.T) {
	cfg := defaultCfg
	listener, err := NewListeners(&cfg)
	require.NoError(t, err)
	require.NotNil(t, listener)
	require.NotNil(t, listener.Grpc)
	require.NotNil(t, listener.Metrics)
}

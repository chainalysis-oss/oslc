package main

import (
	"bytes"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestAsMarkdownAction(t *testing.T) {
	out := bytes.Buffer{}
	cCtx := createContextWithStringFlag(t, "config", "testdata/config.yaml")
	cCtx.App.Writer = &out
	err := asMarkdownAction(cCtx)
	require.NoError(t, err)
	require.NotEmpty(t, out.String())
	require.Equal(t, "\n", out.String()[len(out.String())-1:])
	require.NotEqual(t, "\n", out.String()[len(out.String())-2:])
}

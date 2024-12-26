package main

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func Test_translateValidationError(t *testing.T) {
	invalidConfig := defaultCfg
	invalidConfig.Datastore.Username = ""
	err := validateConfig(&invalidConfig)
	require.Error(t, err)
	require.Error(t, translateValidationError(err))
}

func Test_translateValidationError_pass_through(t *testing.T) {
	_, err := createConfiguration("./testdata/unparseable_config_invalid_json.json")
	require.Error(t, err)
	require.Error(t, translateValidationError(err))
	require.NotContains(t, err.Error(), "invalid configuration")
}

package config

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLoadConfiguration(t *testing.T) {
	t.Run("should error with invalid ports", func(t *testing.T) {
		t.Setenv("GRPC_PORT", "a")
		t.Setenv("HTTP_PORT", "b")
		t.Setenv("TOKEN_KEY", "t")
		_, err := LoadConfiguration()
		require.Error(t, err)
	})

	t.Run("should pass with valid env variables", func(t *testing.T) {
		t.Setenv("GRPC_PORT", "1001")
		t.Setenv("HTTP_PORT", "1002")
		t.Setenv("TOKEN_KEY", "t")
		_, err := LoadConfiguration()
		require.NoError(t, err)
	})
}

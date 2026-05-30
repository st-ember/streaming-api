package hash_test

import (
	"testing"

	"github.com/st-ember/streaming-api/internal/adapter/driven/hash"
	"github.com/stretchr/testify/require"
)

func TestArgon2Hasher(t *testing.T) {
	hasher := hash.NewArgon2Hasher()
	password := "my-secure-password"

	t.Run("should hash and verify correctly", func(t *testing.T) {
		encoded, err := hasher.Hash(password)
		require.NoError(t, err)
		require.Contains(t, encoded, "$argon2id$")

		success := hasher.Verify(password, encoded)
		require.True(t, success)
	})

	t.Run("should fail verification with wrong password", func(t *testing.T) {
		encoded, _ := hasher.Hash(password)
		success := hasher.Verify("wrong-password", encoded)
		require.False(t, success)
	})

	t.Run("should fail with invalid hash format", func(t *testing.T) {
		success := hasher.Verify(password, "invalid-format")
		require.False(t, success)
	})
}

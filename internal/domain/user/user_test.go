package user_test

import (
	"testing"

	"github.com/st-ember/streaming-api/internal/domain/user"
	"github.com/stretchr/testify/require"
)

func TestUser(t *testing.T) {
	t.Run("NewUser should succeed with valid inputs", func(t *testing.T) {
		u, err := user.NewUser("user-1", "test@test.com", "tester", "hashed-pwd")
		require.NoError(t, err)
		require.Equal(t, "user-1", u.ID)
		require.Equal(t, "test@test.com", u.Email)
		require.Equal(t, "tester", u.Username)
		require.NotZero(t, u.CreatedAt)
	})

	t.Run("NewUser should fail if any field is empty", func(t *testing.T) {
		_, err := user.NewUser("", "email", "user", "pass")
		require.ErrorIs(t, err, user.ErrIDEmpty)

		_, err = user.NewUser("id", "", "user", "pass")
		require.ErrorIs(t, err, user.ErrEmailEmpty)

		_, err = user.NewUser("id", "email", "", "pass")
		require.ErrorIs(t, err, user.ErrUsernameEmpty)

		_, err = user.NewUser("id", "email", "user", "")
		require.ErrorIs(t, err, user.ErrPwdHashEmpty)
	})

	t.Run("UpdateUsername should update username and timestamp", func(t *testing.T) {
		u, _ := user.NewUser("user-1", "test@test.com", "tester", "hashed-pwd")
		oldUpdateAt := u.UpdatedAt

		u.UpdateUsername("new-tester")
		require.Equal(t, "new-tester", u.Username)
		require.True(t, u.UpdatedAt.After(oldUpdateAt) || u.UpdatedAt.Equal(oldUpdateAt))
	})

	t.Run("Should correctly identify permissions", func(t *testing.T) {
		u := &user.User{
			Permissions: []string{"video:upload", "video:delete"},
		}

		require.True(t, u.Can("video:upload"))
		require.True(t, u.Can("video:delete"))
		require.False(t, u.Can("video:archive"))
		require.False(t, u.Can(""))
	})
}

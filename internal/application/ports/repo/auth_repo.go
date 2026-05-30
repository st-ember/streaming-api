package repo

import (
	"context"

	"github.com/st-ember/streaming-api/internal/domain/user"
)

type AuthRepo interface {
	// SyncPermissions syncs permissions defined in domain
	// This function runs on startup to seed permissions from domain
	SyncPermissions(ctx context.Context, permissions []string) error

	// FindPermissionsByRole finds permissions related to role
	FindPermissionsByRole(ctx context.Context, roleName string) ([]string, error)

	// FindUserByKey finds user by unique username or email
	FindUserByKey(ctx context.Context, login string) (*user.User, error)

	// FindPermissionsByUserID finds permissions related to user
	FindPermissionsByUserID(ctx context.Context, userID string) ([]string, error)

	// SaveUser upserts a user
	SaveUser(ctx context.Context, u *user.User) error

	// SaveUserRole upserts a userRole
	SaveUserRole(ctx context.Context, userID, roleName string) error
}

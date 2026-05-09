package repo

import "context"

type AuthRepo interface {
	// Run on startup to seed permissions from domain
	SyncPermissions(ctx context.Context, slugs []string) error
}

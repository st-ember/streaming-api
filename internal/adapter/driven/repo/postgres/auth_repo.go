package postgres

import (
	"context"
	"fmt"

	"github.com/st-ember/streaming-api/internal/application/ports/repo"
	"github.com/st-ember/streaming-api/internal/domain/user"

	"database/sql"
)

type PostgresAuthRepo struct {
	q QueryExecutor
}

// QueryExecutor is used to execute queries with db conn as well as transaction
type QueryExecutor interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

func NewPostgresAuthRepo(db *sql.DB) repo.AuthRepo {
	return &PostgresAuthRepo{db}
}

func NewPostgresAuthRepoWithTransaction(tx *sql.Tx) repo.AuthRepo {
	return &PostgresAuthRepo{tx}
}

func (ar *PostgresAuthRepo) SyncPermissions(ctx context.Context, permissions []string) error {
	query := `
		INSERT INTO permissions (id, slug)
		SELECT gen_random_uuid(), s
		FROM UNNEST($1::text[]) AS s
		ON CONFLICT (slug) DO NOTHING
		`

	_, err := ar.q.ExecContext(ctx, query, permissions)
	if err != nil {
		return fmt.Errorf("batch sync permissions: %w", err)
	}

	return nil
}

func (ar *PostgresAuthRepo) FindPermissionsByRole(ctx context.Context, roleName string) ([]string, error) {
	query := `
		SELECT DISTINCT p.slug 
		FROM permissions p 
		JOIN role_permissions rp ON p.id = rp.permission_id
		JOIN roles r ON rp.role_id = r.id
		WHERE r.name = $1;
	`

	rows, err := ar.q.QueryContext(ctx, query, roleName)
	if err != nil {
		return nil, fmt.Errorf("find permissions: %w", err)
	}
	defer rows.Close()

	var permissions []string
	for rows.Next() {
		var p string
		if err := rows.Scan(&p); err != nil {
			return nil, fmt.Errorf("scan permission: %w", err)
		}
		permissions = append(permissions, p)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("row iteration error: %w", err)
	}

	return permissions, nil
}

func (ar *PostgresAuthRepo) FindUserByKey(ctx context.Context, key string) (*user.User, error) {
	query := `
		SELECT id, email, username, password_hash, 
		created_at, updated_at
		FROM users
		WHERE LOWER(email) = LOWER($1) 
		OR LOWER(username) = LOWER($1)
	`

	u := &user.User{}
	err := ar.q.QueryRowContext(ctx, query, key).Scan(
		&u.ID,
		&u.Email,
		&u.Username,
		&u.PasswordHash,
		&u.CreatedAt,
		&u.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found with key %s", key)
		}
		return nil, fmt.Errorf("find user by key %s: %w", key, err)
	}

	return u, nil
}

func (ar *PostgresAuthRepo) FindPermissionsByUserID(ctx context.Context, userID string) ([]string, error) {
	query := `
		SELECT DISTINCT p.slug 
		FROM permissions p
		JOIN role_permissions rp ON p.id = rp.permission_id
		JOIN user_roles ur ON rp.role_id = ur.role_id
		WHERE ur.user_id = $1
	`

	rows, err := ar.q.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("find permissions: %w", err)
	}
	defer rows.Close()

	var permissions []string
	for rows.Next() {
		var p string
		if err := rows.Scan(&p); err != nil {
			return nil, fmt.Errorf("scan permission: %w", err)
		}

		permissions = append(permissions, p)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("row iteration error: %w", err)
	}

	return permissions, nil
}

func (ar *PostgresAuthRepo) SaveUser(ctx context.Context, u *user.User) error {
	query := `
		INSERT INTO users (id, email, username, password_hash, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (id) DO UPDATE SET
			email = EXCLUDED.email,
			username = EXCLUDED.username,
			password_hash = EXCLUDED.password_hash,
			updated_at = EXCLUDED.updated_at
	`

	_, err := ar.q.ExecContext(
		ctx, query, u.ID, u.Email, u.Username,
		u.PasswordHash, u.CreatedAt, u.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("save user %s: %w", u.Email, err)
	}

	return nil
}

func (ar *PostgresAuthRepo) SaveUserRole(ctx context.Context, userID, roleName string) error {
	query := `
		INSERT INTO user_roles (user_id, role_id)
		SELECT $1, id FROM roles WHERE name = $2
		ON CONFLICT (user_id, role_id) DO NOTHING
	`

	res, err := ar.q.ExecContext(ctx, query, userID, roleName)
	if err != nil {
		return fmt.Errorf("assign role %s to user %s: %w", roleName, userID, err)
	}
	cnt, _ := res.RowsAffected() // only returns error when lacking driver support, can safely ignore
	if cnt == 0 {
		return fmt.Errorf("cannot find role id for rolename %s", roleName)
	}

	return nil
}

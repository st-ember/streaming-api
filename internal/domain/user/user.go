package user

import (
	"slices"
	"time"
)

type User struct {
	ID           string
	Email        string
	Username     string
	PasswordHash string
	Permissions  []string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func NewUser(id, email, username, passwordHash string) (*User, error) {
	if id == "" {
		return nil, ErrIDEmpty
	}

	if email == "" {
		return nil, ErrEmailEmpty

	}

	if username == "" {
		return nil, ErrUsernameEmpty

	}

	if passwordHash == "" {
		return nil, ErrPwdHashEmpty
	}

	return &User{
		ID:           id,
		Email:        email,
		Username:     username,
		PasswordHash: passwordHash,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}, nil
}

func (u *User) UpdateUsername(username string) *User {
	u.Username = username
	u.UpdatedAt = time.Now()

	return u
}

func (u *User) Can(permission string) bool {
	return slices.Contains(u.Permissions, permission)
}

package postgres

import (
	"database/sql"
	"fmt"

	_ "github.com/jackc/pgx/v5/stdlib"
)

type DB struct {
	Conn *sql.DB
}

func NewDB(connString string) (*DB, error) {
	conn, err := sql.Open("pgx", connString)
	if err != nil {
		return nil, fmt.Errorf("error opening postgres connection: %w", err)
	}
	if err := conn.Ping(); err != nil {
		return nil, fmt.Errorf("error pinging postgres database: %w", err)
	}

	return &DB{conn}, nil
}

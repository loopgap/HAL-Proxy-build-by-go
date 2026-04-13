package store

import (
	"context"
	"database/sql"
	"time"
)

type TokenBlacklist struct {
	db *sql.DB
}

func NewTokenBlacklist(db *sql.DB) *TokenBlacklist {
	return &TokenBlacklist{db: db}
}

func (tb *TokenBlacklist) Init(ctx context.Context) error {
	_, err := tb.db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS token_blacklist (
			jti TEXT PRIMARY KEY,
			expires_at INTEGER NOT NULL,
			revoked_at INTEGER NOT NULL
		)
	`)
	return err
}

// Add adds a token (by JTI) to the blacklist
func (tb *TokenBlacklist) Add(ctx context.Context, jti string, expiresAt time.Time) error {
	_, err := tb.db.ExecContext(ctx, `
		INSERT OR REPLACE INTO token_blacklist (jti, expires_at, revoked_at)
		VALUES (?, ?, ?)`,
		jti, expiresAt.Unix(), time.Now().Unix())
	return err
}

// IsRevoked checks if a token is in the blacklist
func (tb *TokenBlacklist) IsRevoked(ctx context.Context, jti string) (bool, error) {
	var exists int
	err := tb.db.QueryRowContext(ctx,
		"SELECT 1 FROM token_blacklist WHERE jti = ?", jti).Scan(&exists)
	if err == sql.ErrNoRows {
		return false, nil
	}
	return err == nil, err
}

// Cleanup removes expired entries from the blacklist
func (tb *TokenBlacklist) Cleanup(ctx context.Context) error {
	_, err := tb.db.ExecContext(ctx, `
		DELETE FROM token_blacklist WHERE expires_at < ?`,
		time.Now().Unix())
	return err
}

package store

import (
	"context"
	"testing"
	"time"
)

func TestTokenBlacklistLifecycleAndCleanup(t *testing.T) {
	t.Parallel()

	repo, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	if err := repo.Blacklist.Add(ctx, "active-token", time.Now().Add(time.Hour)); err != nil {
		t.Fatalf("add active token: %v", err)
	}
	if revoked, err := repo.Blacklist.IsRevoked(ctx, "active-token"); err != nil || !revoked {
		t.Fatalf("active token revoked=%v err=%v", revoked, err)
	}

	if err := repo.Blacklist.Add(ctx, "expired-token", time.Now().Add(-time.Hour)); err != nil {
		t.Fatalf("add expired token: %v", err)
	}
	if err := repo.CleanupExpiredTokens(ctx); err != nil {
		t.Fatalf("cleanup expired tokens: %v", err)
	}
	if revoked, err := repo.Blacklist.IsRevoked(ctx, "expired-token"); err != nil || revoked {
		t.Fatalf("expired token revoked=%v err=%v", revoked, err)
	}
}

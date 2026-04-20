package middleware

import (
	"context"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

func TestGenerateToken(t *testing.T) {
	config := JWTConfig{
		Secret:          "test-secret-key",
		ExpirationHours: 24,
		Issuer:          "test-issuer",
	}

	t.Run("token contains jti claim", func(t *testing.T) {
		tokenString, err := GenerateToken(config, "user123", "testuser", []string{"admin"})
		if err != nil {
			t.Fatalf("GenerateToken() error = %v", err)
		}

		// Parse the token to verify jti claim
		token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
			return []byte(config.Secret), nil
		})
		if err != nil {
			t.Fatalf("Failed to parse token: %v", err)
		}

		claims, ok := token.Claims.(*Claims)
		if !ok {
			t.Fatal("Failed to get claims")
		}

		if claims.ID == "" {
			t.Error("GenerateToken() jti (ID) claim is empty, expected non-empty UUID")
		}
	})

	t.Run("each token has unique jti", func(t *testing.T) {
		token1, err := GenerateToken(config, "user123", "testuser", []string{"admin"})
		if err != nil {
			t.Fatalf("GenerateToken() error = %v", err)
		}

		token2, err := GenerateToken(config, "user123", "testuser", []string{"admin"})
		if err != nil {
			t.Fatalf("GenerateToken() error = %v", err)
		}

		if token1 == token2 {
			t.Error("GenerateToken() produced identical tokens, expected unique tokens")
		}

		// Parse both tokens and compare their IDs
		claims1, err := parseToken(token1, config.Secret)
		if err != nil {
			t.Fatalf("Failed to parse token1: %v", err)
		}
		claims2, err := parseToken(token2, config.Secret)
		if err != nil {
			t.Fatalf("Failed to parse token2: %v", err)
		}

		if claims1.ID == claims2.ID {
			t.Errorf("GenerateToken() tokens have same jti %q, expected unique", claims1.ID)
		}
	})

	t.Run("token contains expected claims", func(t *testing.T) {
		tokenString, err := GenerateToken(config, "user123", "testuser", []string{"admin", "editor"})
		if err != nil {
			t.Fatalf("GenerateToken() error = %v", err)
		}

		claims, err := parseToken(tokenString, config.Secret)
		if err != nil {
			t.Fatalf("Failed to parse token: %v", err)
		}

		if claims.UserID != "user123" {
			t.Errorf("UserID = %q, want %q", claims.UserID, "user123")
		}
		if claims.Username != "testuser" {
			t.Errorf("Username = %q, want %q", claims.Username, "testuser")
		}
		if len(claims.Roles) != 2 {
			t.Errorf("Roles len = %d, want 2", len(claims.Roles))
		}
		if claims.Issuer != "test-issuer" {
			t.Errorf("Issuer = %q, want %q", claims.Issuer, "test-issuer")
		}
		if claims.Subject != "user123" {
			t.Errorf("Subject = %q, want %q", claims.Subject, "user123")
		}
		if claims.ExpiresAt == nil {
			t.Error("ExpiresAt is nil")
		}
		if claims.IssuedAt == nil {
			t.Error("IssuedAt is nil")
		}
	})

	t.Run("token expiration is set correctly", func(t *testing.T) {
		tokenString, err := GenerateToken(config, "user123", "testuser", []string{})
		if err != nil {
			t.Fatalf("GenerateToken() error = %v", err)
		}

		claims, err := parseToken(tokenString, config.Secret)
		if err != nil {
			t.Fatalf("Failed to parse token: %v", err)
		}

		// Check expiration is approximately 24 hours from now
		expectedExpiry := time.Now().Add(24 * time.Hour)
		actualExpiry := claims.ExpiresAt.Time

		diff := actualExpiry.Sub(expectedExpiry)
		if diff < -time.Minute || diff > time.Minute {
			t.Errorf("ExpiresAt = %v, expected approximately %v", actualExpiry, expectedExpiry)
		}
	})
}

// parseToken parses a JWT token string with the given secret
func parseToken(tokenString string, secret string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})
	if err != nil {
		return nil, err
	}
	claims, ok := token.Claims.(*Claims)
	if !ok {
		return nil, err
	}
	return claims, nil
}

func TestValidateToken(t *testing.T) {
	config := JWTConfig{
		Secret:          "test-secret-key",
		ExpirationHours: 24,
		Issuer:          "test-issuer",
	}

	t.Run("valid token passes validation", func(t *testing.T) {
		tokenString, err := GenerateToken(config, "user123", "testuser", []string{"admin"})
		if err != nil {
			t.Fatalf("GenerateToken() error = %v", err)
		}

		auth := NewJWTAuthenticator(config)
		claims, err := auth.ValidateToken(context.Background(), tokenString)
		if err != nil {
			t.Errorf("ValidateToken() error = %v", err)
		}
		if claims.UserID != "user123" {
			t.Errorf("UserID = %q, want %q", claims.UserID, "user123")
		}
	})

	t.Run("token with empty jti can still be validated", func(t *testing.T) {
		// Create a token manually without ID (simulating old token)
		claims := &Claims{
			UserID:   "user123",
			Username: "testuser",
			Roles:    []string{"admin"},
			RegisteredClaims: jwt.RegisteredClaims{
				Issuer:    config.Issuer,
				Subject:   "user123",
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(config.ExpirationHours) * time.Hour)),
				IssuedAt:  jwt.NewNumericDate(time.Now()),
				// No ID set - simulating old token
			},
		}
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		tokenString, err := token.SignedString([]byte(config.Secret))
		if err != nil {
			t.Fatalf("Failed to create token: %v", err)
		}

		// ValidateToken should still work (just can't check blacklist)
		auth := NewJWTAuthenticator(config)
		_, err = auth.ValidateToken(context.Background(), tokenString)
		if err != nil {
			t.Errorf("ValidateToken() error = %v (old tokens without jti should still validate)", err)
		}
	})

	t.Run("invalid token fails validation", func(t *testing.T) {
		auth := NewJWTAuthenticator(config)
		_, err := auth.ValidateToken(context.Background(), "invalid-token")
		if err == nil {
			t.Error("ValidateToken() expected error for invalid token")
		}
	})

	t.Run("token with wrong secret fails validation", func(t *testing.T) {
		tokenString, err := GenerateToken(config, "user123", "testuser", []string{"admin"})
		if err != nil {
			t.Fatalf("GenerateToken() error = %v", err)
		}

		auth := NewJWTAuthenticator(JWTConfig{
			Secret:          "wrong-secret",
			ExpirationHours: 24,
			Issuer:          "test-issuer",
		})
		_, err = auth.ValidateToken(context.Background(), tokenString)
		if err == nil {
			t.Error("ValidateToken() expected error for token with wrong secret")
		}
	})

	t.Run("expired token fails validation", func(t *testing.T) {
		// Create a token that expired 1 hour ago
		claims := &Claims{
			UserID:   "user123",
			Username: "testuser",
			Roles:    []string{"admin"},
			RegisteredClaims: jwt.RegisteredClaims{
				ID:        uuid.New().String(),
				Issuer:    config.Issuer,
				Subject:   "user123",
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(-1 * time.Hour)), // Expired 1 hour ago
				IssuedAt:  jwt.NewNumericDate(time.Now().Add(-2 * time.Hour)),
			},
		}
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		tokenString, err := token.SignedString([]byte(config.Secret))
		if err != nil {
			t.Fatalf("Failed to create expired token: %v", err)
		}

		auth := NewJWTAuthenticator(config)
		_, err = auth.ValidateToken(context.Background(), tokenString)
		if err == nil {
			t.Error("ValidateToken() expected error for expired token")
		}
	})

	t.Run("token without expiration fails validation", func(t *testing.T) {
		// Create a token without expiration claim
		claims := &Claims{
			UserID:   "user123",
			Username: "testuser",
			Roles:    []string{"admin"},
			RegisteredClaims: jwt.RegisteredClaims{
				ID:       uuid.New().String(),
				Issuer:   config.Issuer,
				Subject:  "user123",
				IssuedAt: jwt.NewNumericDate(time.Now()),
				// No ExpiresAt - this should fail with WithExpirationRequired
			},
		}
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		tokenString, err := token.SignedString([]byte(config.Secret))
		if err != nil {
			t.Fatalf("Failed to create token without expiration: %v", err)
		}

		auth := NewJWTAuthenticator(config)
		_, err = auth.ValidateToken(context.Background(), tokenString)
		if err == nil {
			t.Error("ValidateToken() expected error for token without expiration")
		}
	})
}

func TestHasRole(t *testing.T) {
	claims := &Claims{
		UserID:   "user123",
		Username: "testuser",
		Roles:    []string{"admin", "editor"},
	}

	tests := []struct {
		name     string
		role     string
		expected bool
	}{
		{"admin role exists", "admin", true},
		{"editor role exists", "editor", true},
		{"viewer role does not exist", "viewer", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := HasRole(claims, tt.role); got != tt.expected {
				t.Errorf("HasRole(%q) = %v, want %v", tt.role, got, tt.expected)
			}
		})
	}

	t.Run("nil claims returns false", func(t *testing.T) {
		if got := HasRole(nil, "admin"); got != false {
			t.Errorf("HasRole(nil, \"admin\") = %v, want false", got)
		}
	})
}

func TestIsAdmin(t *testing.T) {
	tests := []struct {
		name     string
		roles    []string
		expected bool
	}{
		{"user with admin role", []string{"admin"}, true},
		{"user with admin role among others", []string{"editor", "admin", "viewer"}, true},
		{"user without admin role", []string{"editor", "viewer"}, false},
		{"user with no roles", []string{}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			claims := &Claims{Roles: tt.roles}
			if got := IsAdmin(claims); got != tt.expected {
				t.Errorf("IsAdmin() = %v, want %v", got, tt.expected)
			}
		})
	}

	t.Run("nil claims returns false", func(t *testing.T) {
		if got := IsAdmin(nil); got != false {
			t.Errorf("IsAdmin(nil) = %v, want false", got)
		}
	})
}

func TestGetClaimsFromContext(t *testing.T) {
	claims := &Claims{
		UserID:   "user123",
		Username: "testuser",
		Roles:    []string{"admin"},
	}

	t.Run("claims found in context", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), claimsContextKey, claims)
		got, ok := GetClaimsFromContext(ctx)
		if !ok {
			t.Error("GetClaimsFromContext() ok = false, want true")
		}
		if got.UserID != "user123" {
			t.Errorf("UserID = %q, want %q", got.UserID, "user123")
		}
	})

	t.Run("claims not found in context", func(t *testing.T) {
		_, ok := GetClaimsFromContext(context.Background())
		if ok {
			t.Error("GetClaimsFromContext() ok = true, want false")
		}
	})

	t.Run("nil context", func(t *testing.T) {
		_, ok := GetClaimsFromContext(nil)
		if ok {
			t.Error("GetClaimsFromContext(nil) ok = true, want false")
		}
	})
}

func TestGetUserIDFromContext(t *testing.T) {
	claims := &Claims{
		UserID:   "user123",
		Username: "testuser",
		Roles:    []string{"admin"},
	}

	t.Run("user ID found", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), claimsContextKey, claims)
		got := GetUserIDFromContext(ctx)
		if got != "user123" {
			t.Errorf("GetUserIDFromContext() = %q, want %q", got, "user123")
		}
	})

	t.Run("user ID not found", func(t *testing.T) {
		got := GetUserIDFromContext(context.Background())
		if got != "" {
			t.Errorf("GetUserIDFromContext() = %q, want %q", got, "")
		}
	})
}

package middleware

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type JWTConfig struct {
	Secret          string
	ExpirationHours int
	Issuer          string
}

// TokenBlacklist interface for checking revoked tokens
type TokenBlacklist interface {
	IsRevoked(ctx context.Context, jti string) (bool, error)
}

type Claims struct {
	UserID   string   `json:"user_id"`
	Username string   `json:"username"`
	Roles    []string `json:"roles"`
	jwt.RegisteredClaims
}

// JWTAuthenticator handles JWT authentication with optional token blacklist
type JWTAuthenticator struct {
	Config    JWTConfig
	Blacklist TokenBlacklist
}

// NewJWTAuthenticator creates a new JWTAuthenticator instance
func NewJWTAuthenticator(config JWTConfig) *JWTAuthenticator {
	return &JWTAuthenticator{Config: config}
}

// Middleware returns an HTTP middleware for JWT authentication
func (m *JWTAuthenticator) Middleware() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				http.Error(w, "missing_authorization_header", http.StatusUnauthorized)
				return
			}
			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
				http.Error(w, "invalid_authorization_format", http.StatusUnauthorized)
				return
			}
			tokenString := parts[1]
			claims, err := m.ValidateToken(ctx, tokenString)
			if err != nil {
				http.Error(w, "invalid_token", http.StatusUnauthorized)
				return
			}
			ctx = context.WithValue(ctx, claimsContextKey, claims)
			r = r.WithContext(ctx)
			next.ServeHTTP(w, r)
		})
	}
}

// ValidateToken validates a JWT token and checks the blacklist
func (m *JWTAuthenticator) ValidateToken(ctx context.Context, tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("invalid signing method")
		}
		return []byte(m.Config.Secret), nil
	}, jwt.WithExpirationRequired())
	if err != nil {
		return nil, err
	}
	if !token.Valid {
		return nil, errors.New("invalid token")
	}
	claims, ok := token.Claims.(*Claims)
	if !ok {
		return nil, errors.New("invalid claims")
	}
	if m.Config.Issuer != "" {
		issuer, err := claims.GetIssuer()
		if err != nil {
			return nil, errors.New("invalid_claims")
		}
		if issuer != m.Config.Issuer {
			return nil, errors.New("invalid_issuer")
		}
	}

	// Check blacklist for revoked tokens
	if m.Blacklist != nil {
		revoked, err := m.Blacklist.IsRevoked(ctx, claims.ID)
		if err != nil {
			return nil, err
		}
		if revoked {
			return nil, errors.New("token has been revoked")
		}
	}

	return claims, nil
}

// JWTAuth is kept for backwards compatibility - wraps JWTAuthenticator
func JWTAuth(config JWTConfig) Middleware {
	return NewJWTAuthenticator(config).Middleware()
}

func GenerateToken(config JWTConfig, userID, username string, roles []string) (string, error) {
	claims := &Claims{
		UserID:   userID,
		Username: username,
		Roles:    roles,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        uuid.New().String(),
			Issuer:    config.Issuer,
			Subject:   userID,
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(config.ExpirationHours) * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(config.Secret))
}

type contextKey string

const claimsContextKey contextKey = "claims"

func GetClaimsFromContext(ctx context.Context) (*Claims, bool) {
	if ctx == nil {
		return nil, false
	}
	if val := ctx.Value(claimsContextKey); val != nil {
		if claims, ok := val.(*Claims); ok {
			return claims, true
		}
	}
	return nil, false
}

func ContextWithClaims(ctx context.Context, claims *Claims) context.Context {
	return context.WithValue(ctx, claimsContextKey, claims)
}

func GetUserIDFromContext(ctx context.Context) string {
	claims, ok := GetClaimsFromContext(ctx)
	if !ok || claims == nil {
		return ""
	}
	return claims.UserID
}

// HasRole checks if the claims contain the specified role
func HasRole(claims *Claims, role string) bool {
	if claims == nil {
		return false
	}
	for _, r := range claims.Roles {
		if r == role {
			return true
		}
	}
	return false
}

// IsAdmin returns true if the claims have the admin role
func IsAdmin(claims *Claims) bool {
	return HasRole(claims, "admin")
}

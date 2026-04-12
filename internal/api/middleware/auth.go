package middleware

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type JWTConfig struct {
	Secret          string
	ExpirationHours int
	Issuer          string
}

type Claims struct {
	UserID   string   `json:"user_id"`
	Username string   `json:"username"`
	Roles    []string `json:"roles"`
	jwt.RegisteredClaims
}

func JWTAuth(config JWTConfig) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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
			token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, errors.New("invalid signing method")
				}
				return []byte(config.Secret), nil
			})
			if err != nil || !token.Valid {
				http.Error(w, "invalid_token", http.StatusUnauthorized)
				return
			}
			claims, ok := token.Claims.(*Claims)
			if !ok {
				http.Error(w, "invalid_claims", http.StatusUnauthorized)
				return
			}
			if config.Issuer != "" {
				issuer, _ := claims.GetIssuer()
				if issuer != config.Issuer {
					http.Error(w, "invalid_issuer", http.StatusUnauthorized)
					return
				}
			}
			ctx := context.WithValue(r.Context(), claimsContextKey, claims)
			r = r.WithContext(ctx)
			next.ServeHTTP(w, r)
		})
	}
}

func GenerateToken(config JWTConfig, userID, username string, roles []string) (string, error) {
	claims := &Claims{
		UserID:   userID,
		Username: username,
		Roles:    roles,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    config.Issuer,
			Subject:   userID,
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(config.ExpirationHours) * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(config.Secret))
}

func ValidateToken(config JWTConfig, tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("invalid signing method")
		}
		return []byte(config.Secret), nil
	})
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
	return claims, nil
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

func GetUserIDFromContext(ctx context.Context) string {
	claims, ok := GetClaimsFromContext(ctx)
	if !ok || claims == nil {
		return ""
	}
	return claims.UserID
}

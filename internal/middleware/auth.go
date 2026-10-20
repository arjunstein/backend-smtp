package middleware

import (
	"context"
	"crypto/subtle"
	"net/http"
	"strings"

	"github.com/arjunstein/backend-smtp/internal/service"
	"github.com/arjunstein/backend-smtp/pkg/response"
)

type contextKey string

const (
	UserClaimsKey contextKey = "jwt_user_claims"
	UserIDKey     contextKey = "jwt_user_id"
)

// APIKeyAuth validates incoming requests using X-API-Key or Authorization Bearer header.
// It uses constant-time string comparison to prevent timing side-channel attacks.
func APIKeyAuth(expectedKey string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if expectedKey == "" {
				next.ServeHTTP(w, r)
				return
			}

			clientKey := extractToken(r)
			if clientKey == "" || subtle.ConstantTimeCompare([]byte(clientKey), []byte(expectedKey)) != 1 {
				response.Unauthorized(w, "Invalid or missing API authentication credentials")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// JWTAuth validates incoming requests using JWT Bearer token in Authorization header.
func JWTAuth(authService service.AuthService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			tokenStr := extractBearerToken(r)
			if tokenStr == "" {
				response.Unauthorized(w, "Missing or invalid authorization token")
				return
			}

			claims, err := authService.ValidateToken(tokenStr)
			if err != nil {
				response.Unauthorized(w, "Invalid or expired authorization token")
				return
			}

			ctx := context.WithValue(r.Context(), UserClaimsKey, claims)
			ctx = context.WithValue(ctx, UserIDKey, claims.UserID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// CombinedAuth allows access if the request has EITHER a valid API Key OR a valid JWT token.
func CombinedAuth(expectedAPIKey string, authService service.AuthService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// 1. Check if valid API Key
			clientKey := extractToken(r)
			if expectedAPIKey != "" && clientKey != "" && subtle.ConstantTimeCompare([]byte(clientKey), []byte(expectedAPIKey)) == 1 {
				next.ServeHTTP(w, r)
				return
			}

			// 2. Check if valid JWT Token
			tokenStr := extractBearerToken(r)
			if tokenStr != "" && authService != nil {
				claims, err := authService.ValidateToken(tokenStr)
				if err == nil && claims != nil {
					ctx := context.WithValue(r.Context(), UserClaimsKey, claims)
					ctx = context.WithValue(ctx, UserIDKey, claims.UserID)
					next.ServeHTTP(w, r.WithContext(ctx))
					return
				}
			}

			// If both expectedAPIKey is empty and no authService, bypass (e.g. dev mode)
			if expectedAPIKey == "" && authService == nil {
				next.ServeHTTP(w, r)
				return
			}

			response.Unauthorized(w, "Invalid or missing authentication credentials (API key or JWT Bearer token required)")
		})
	}
}

// GetUserID extracts authenticated user ID from context.
func GetUserID(ctx context.Context) (uint, bool) {
	id, ok := ctx.Value(UserIDKey).(uint)
	return id, ok
}

// GetUserClaims extracts authenticated user claims from context.
func GetUserClaims(ctx context.Context) (*service.JWTClaims, bool) {
	claims, ok := ctx.Value(UserClaimsKey).(*service.JWTClaims)
	return claims, ok
}

func extractToken(r *http.Request) string {
	if key := strings.TrimSpace(r.Header.Get("X-API-Key")); key != "" {
		return key
	}
	return extractBearerToken(r)
}

func extractBearerToken(r *http.Request) string {
	if authHeader := strings.TrimSpace(r.Header.Get("Authorization")); authHeader != "" {
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) == 2 && strings.EqualFold(parts[0], "Bearer") {
			return strings.TrimSpace(parts[1])
		}
		return authHeader
	}
	return ""
}

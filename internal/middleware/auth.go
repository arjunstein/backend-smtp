package middleware

import (
	"crypto/subtle"
	"net/http"
	"strings"

	"github.com/arjunstein/backend-smtp/pkg/response"
)

// APIKeyAuth validates incoming requests using X-API-Key or Authorization Bearer header.
// It uses constant-time string comparison to prevent timing side-channel attacks.
func APIKeyAuth(expectedKey string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// If no API key is configured, bypass authentication (convenient for local open dev)
			if expectedKey == "" {
				next.ServeHTTP(w, r)
				return
			}

			clientKey := extractAPIKey(r)
			if clientKey == "" || subtle.ConstantTimeCompare([]byte(clientKey), []byte(expectedKey)) != 1 {
				response.Unauthorized(w, "Invalid or missing API authentication credentials")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func extractAPIKey(r *http.Request) string {
	// 1. Check X-API-Key header
	if key := strings.TrimSpace(r.Header.Get("X-API-Key")); key != "" {
		return key
	}

	// 2. Check Authorization header (Bearer <token> or raw token)
	if authHeader := strings.TrimSpace(r.Header.Get("Authorization")); authHeader != "" {
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) == 2 && strings.EqualFold(parts[0], "Bearer") {
			return strings.TrimSpace(parts[1])
		}
		return authHeader
	}

	return ""
}

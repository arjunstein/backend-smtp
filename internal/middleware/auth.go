package middleware

import (
	"net/http"

	"github.com/arjunstein/backend-smtp/pkg/response"
)

// APIKeyAuth validates incoming requests using X-API-Key HTTP header.
// If expectedKey is empty, authentication is bypassed (convenient for local dev).
func APIKeyAuth(expectedKey string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if expectedKey == "" {
				next.ServeHTTP(w, r)
				return
			}

			clientKey := r.Header.Get("X-API-Key")
			if clientKey == "" || clientKey != expectedKey {
				response.Unauthorized(w, "Invalid or missing API authentication credentials")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

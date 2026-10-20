package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/arjunstein/backend-smtp/internal/model"
	"github.com/arjunstein/backend-smtp/internal/service"
	"github.com/stretchr/testify/assert"
)

func TestAPIKeyAuth_ValidXAPIKey(t *testing.T) {
	handler := APIKeyAuth("secret-key-123")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("X-API-Key", "secret-key-123")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "OK", w.Body.String())
}

func TestAPIKeyAuth_ValidBearerToken(t *testing.T) {
	handler := APIKeyAuth("secret-key-123")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer secret-key-123")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "OK", w.Body.String())
}

func TestAPIKeyAuth_ValidRawAuthorizationHeader(t *testing.T) {
	handler := APIKeyAuth("secret-key-123")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "secret-key-123")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "OK", w.Body.String())
}

func TestAPIKeyAuth_InvalidKey(t *testing.T) {
	handler := APIKeyAuth("secret-key-123")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("X-API-Key", "wrong-key")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAPIKeyAuth_MissingKey(t *testing.T) {
	handler := APIKeyAuth("secret-key-123")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAPIKeyAuth_EmptyKeyBypass(t *testing.T) {
	handler := APIKeyAuth("")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("BYPASS"))
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "BYPASS", w.Body.String())
}

type dummyAuthService struct {
	validToken string
	claims     *service.JWTClaims
}

func (d *dummyAuthService) Login(ctx context.Context, req *model.LoginRequest) (*model.AuthResponse, error) {
	return nil, nil
}
func (d *dummyAuthService) RefreshToken(ctx context.Context, req *model.RefreshTokenRequest) (*model.RefreshTokenResponse, error) {
	return nil, nil
}
func (d *dummyAuthService) GetProfile(ctx context.Context, userID uint) (*model.UserResponse, error) {
	return nil, nil
}
func (d *dummyAuthService) SeedInitialUser(ctx context.Context, name, username, email, password string) (*model.User, error) {
	return nil, nil
}
func (d *dummyAuthService) ValidateToken(tokenString string) (*service.JWTClaims, error) {
	if tokenString == d.validToken {
		return d.claims, nil
	}
	return nil, service.ErrInvalidToken
}

func TestJWTAuth_ValidToken(t *testing.T) {
	mockAuth := &dummyAuthService{
		validToken: "valid-jwt-token",
		claims: &service.JWTClaims{
			UserID:   42,
			Username: "arjun",
			Email:    "arjun@example.com",
		},
	}

	handler := JWTAuth(mockAuth)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID, ok := GetUserID(r.Context())
		assert.True(t, ok)
		assert.Equal(t, uint(42), userID)
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer valid-jwt-token")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestJWTAuth_InvalidToken(t *testing.T) {
	mockAuth := &dummyAuthService{validToken: "valid-jwt-token"}

	handler := JWTAuth(mockAuth)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer expired-token")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestCombinedAuth_AllowsAPIKey(t *testing.T) {
	mockAuth := &dummyAuthService{validToken: "valid-jwt"}

	handler := CombinedAuth("secret-api-key", mockAuth)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("API_KEY_SUCCESS"))
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("X-API-Key", "secret-api-key")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "API_KEY_SUCCESS", w.Body.String())
}

func TestCombinedAuth_AllowsJWT(t *testing.T) {
	mockAuth := &dummyAuthService{
		validToken: "valid-jwt",
		claims: &service.JWTClaims{
			UserID:   10,
			Username: "dashboard_user",
		},
	}

	handler := CombinedAuth("secret-api-key", mockAuth)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID, ok := GetUserID(r.Context())
		assert.True(t, ok)
		assert.Equal(t, uint(10), userID)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("JWT_SUCCESS"))
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer valid-jwt")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "JWT_SUCCESS", w.Body.String())
}

func TestCombinedAuth_RejectsInvalid(t *testing.T) {
	mockAuth := &dummyAuthService{validToken: "valid-jwt"}

	handler := CombinedAuth("secret-api-key", mockAuth)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer invalid-credentials")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}


package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/arjunstein/backend-smtp/internal/middleware"
	"github.com/arjunstein/backend-smtp/internal/model"
	"github.com/arjunstein/backend-smtp/internal/service"
	"github.com/arjunstein/backend-smtp/pkg/response"
)

// AuthHandler handles authentication and user profile HTTP requests.
type AuthHandler struct {
	authService service.AuthService
}

// NewAuthHandler creates a new AuthHandler instance.
func NewAuthHandler(authService service.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}


// Login godoc
// @Summary Authenticate user
// @Description Authenticates credentials and returns a signed JWT token
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body model.LoginRequest true "User Login Credentials"
// @Success 200 {object} response.APIResponse{data=model.AuthResponse} "Successfully authenticated"
// @Failure 400 {object} response.ErrorResponse "Invalid input format"
// @Failure 401 {object} response.UnauthorizedResponse "Invalid username/email or password"
// @Failure 500 {object} response.InternalErrorResponse "Internal server error"
// @Router /api/v1/auth/login [post]
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req model.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "Invalid JSON payload", err)
		return
	}

	authResp, err := h.authService.Login(r.Context(), &req)
	if err != nil {
		if errors.Is(err, service.ErrInvalidCredentials) {
			response.Unauthorized(w, "Invalid username/email or password")
			return
		}
		if err.Error() == "username or email is required" || err.Error() == "password is required" {
			response.BadRequest(w, err.Error(), nil)
			return
		}
		response.InternalServerError(w, "Failed to authenticate", err)
		return
	}

	response.Success(w, http.StatusOK, "Login successful", authResp)
}

// Me godoc
// @Summary Get current user profile
// @Description Returns the profile of the currently authenticated user using JWT
// @Tags Auth
// @Security BearerAuth
// @Produce json
// @Success 200 {object} response.APIResponse{data=model.UserResponse} "Current user profile"
// @Failure 401 {object} response.UnauthorizedResponse "Missing or invalid token"
// @Failure 500 {object} response.InternalErrorResponse "Internal server error"
// @Router /api/v1/auth/me [get]
func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		response.Unauthorized(w, "Authentication required")
		return
	}

	userResp, err := h.authService.GetProfile(r.Context(), userID)
	if err != nil {
		if errors.Is(err, service.ErrInvalidCredentials) {
			response.Unauthorized(w, "User not found")
			return
		}
		response.InternalServerError(w, "Failed to get profile", err)
		return
	}

	response.Success(w, http.StatusOK, "User profile retrieved successfully", userResp)
}

// RefreshToken godoc
// @Summary Refresh access token
// @Description Uses a valid refresh token to issue a new dynamic access token and rotated refresh token
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body model.RefreshTokenRequest true "Refresh Token Payload"
// @Success 200 {object} response.APIResponse{data=model.RefreshTokenResponse} "Successfully refreshed tokens"
// @Failure 400 {object} response.ErrorResponse "Invalid input format"
// @Failure 401 {object} response.UnauthorizedResponse "Invalid, expired, or wrong token type"
// @Failure 500 {object} response.InternalErrorResponse "Internal server error"
// @Router /api/v1/auth/refresh [post]
func (h *AuthHandler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	var req model.RefreshTokenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "Invalid JSON payload", err)
		return
	}

	refreshResp, err := h.authService.RefreshToken(r.Context(), &req)
	if err != nil {
		if errors.Is(err, service.ErrInvalidToken) || errors.Is(err, service.ErrInvalidTokenType) {
			response.Unauthorized(w, "Invalid or expired refresh token")
			return
		}
		if err.Error() == "refresh_token is required" {
			response.BadRequest(w, err.Error(), nil)
			return
		}
		response.InternalServerError(w, "Failed to refresh token", err)
		return
	}

	response.Success(w, http.StatusOK, "Token refreshed successfully", refreshResp)
}


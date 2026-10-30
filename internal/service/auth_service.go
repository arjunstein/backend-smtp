package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/arjunstein/backend-smtp/internal/model"
	"github.com/arjunstein/backend-smtp/internal/repository"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// Token type constants for dynamic JWT access & refresh separation
const (
	TokenTypeAccess  = "access"
	TokenTypeRefresh = "refresh"
)

var (
	ErrInvalidCredentials = errors.New("invalid username/email or password")
	ErrInvalidToken       = errors.New("invalid or expired authentication token")
	ErrInvalidTokenType   = errors.New("token type is invalid for this operation")
)

// JWTClaims represents claims stored inside dynamic JWT tokens.
type JWTClaims struct {
	UserID    uint   `json:"user_id"`
	Username  string `json:"username"`
	Email     string `json:"email"`
	TokenType string `json:"token_type"`
	jwt.RegisteredClaims
}

// AuthService defines user authentication and authorization logic.
type AuthService interface {
	Login(ctx context.Context, req *model.LoginRequest) (*model.AuthResponse, error)
	RefreshToken(ctx context.Context, req *model.RefreshTokenRequest) (*model.RefreshTokenResponse, error)
	GetProfile(ctx context.Context, userID uint) (*model.UserResponse, error)
	ValidateToken(tokenString string) (*JWTClaims, error)
	SeedInitialUser(ctx context.Context, name, username, email, password string) (*model.User, error)
}

type authService struct {
	userRepo   repository.UserRepository
	jwtSecret  []byte
	accessExp  time.Duration
	refreshExp time.Duration
}

// NewAuthService creates a new AuthService instance with dynamic token expirations.
func NewAuthService(userRepo repository.UserRepository, jwtSecret string, accessMins int, refreshDays int) AuthService {
	if accessMins <= 0 {
		accessMins = 60
	}
	if refreshDays <= 0 {
		refreshDays = 7
	}
	return &authService{
		userRepo:   userRepo,
		jwtSecret:  []byte(jwtSecret),
		accessExp:  time.Duration(accessMins) * time.Minute,
		refreshExp: time.Duration(refreshDays) * 24 * time.Hour,
	}
}

// SeedInitialUser creates the initial single user if no user exists in the database.
func (s *authService) SeedInitialUser(ctx context.Context, name, username, email, password string) (*model.User, error) {
	count, err := s.userRepo.Count(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed checking user count: %w", err)
	}
	if count > 0 {
		return nil, nil // Initial user already exists
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed hashing password: %w", err)
	}

	user := &model.User{
		Name:     strings.TrimSpace(name),
		Username: strings.ToLower(strings.TrimSpace(username)),
		Email:    strings.ToLower(strings.TrimSpace(email)),
		Password: string(hashedPassword),
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("failed saving initial user: %w", err)
	}

	return user, nil
}

func (s *authService) Login(ctx context.Context, req *model.LoginRequest) (*model.AuthResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	user, err := s.userRepo.FindByUsernameOrEmail(ctx, req.UsernameOrEmail)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return nil, ErrInvalidCredentials
		}
		return nil, fmt.Errorf("failed finding user: %w", err)
	}

	// Constant-time bcrypt comparison
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return nil, ErrInvalidCredentials
	}

	accessToken, refreshToken, expiresIn, err := s.generateTokenPair(user)
	if err != nil {
		return nil, err
	}

	userResp := model.ToUserResponse(user)
	return &model.AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    expiresIn,
		User:         userResp,
	}, nil
}

func (s *authService) RefreshToken(ctx context.Context, req *model.RefreshTokenRequest) (*model.RefreshTokenResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	token, err := jwt.ParseWithClaims(req.RefreshToken, &JWTClaims{}, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return s.jwtSecret, nil
	})

	if err != nil || !token.Valid {
		return nil, ErrInvalidToken
	}

	claims, ok := token.Claims.(*JWTClaims)
	if !ok || claims.TokenType != TokenTypeRefresh {
		return nil, ErrInvalidTokenType
	}

	user, err := s.userRepo.FindByID(ctx, claims.UserID)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return nil, ErrInvalidToken
		}
		return nil, err
	}

	accessToken, newRefreshToken, expiresIn, err := s.generateTokenPair(user)
	if err != nil {
		return nil, err
	}

	return &model.RefreshTokenResponse{
		AccessToken:  accessToken,
		RefreshToken: newRefreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    expiresIn,
	}, nil
}

func (s *authService) GetProfile(ctx context.Context, userID uint) (*model.UserResponse, error) {
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return nil, ErrInvalidCredentials
		}
		return nil, err
	}
	resp := model.ToUserResponse(user)
	return &resp, nil
}

func (s *authService) ValidateToken(tokenString string) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return s.jwtSecret, nil
	})

	if err != nil || !token.Valid {
		return nil, ErrInvalidToken
	}

	claims, ok := token.Claims.(*JWTClaims)
	if !ok {
		return nil, ErrInvalidToken
	}

	// Dynamic tokens must have type "access" to access resource endpoints
	if claims.TokenType != TokenTypeAccess {
		return nil, ErrInvalidTokenType
	}

	return claims, nil
}

func (s *authService) generateTokenPair(user *model.User) (accessToken string, refreshToken string, expiresIn int64, err error) {
	now := time.Now()
	accessExpiresAt := now.Add(s.accessExp)
	refreshExpiresAt := now.Add(s.refreshExp)

	// 1. Short-lived Access Token
	accessClaims := JWTClaims{
		UserID:    user.ID,
		Username:  user.Username,
		Email:     user.Email,
		TokenType: TokenTypeAccess,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   fmt.Sprintf("%d", user.ID),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(accessExpiresAt),
			Issuer:    "backend-smtp",
		},
	}
	accToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	accessToken, err = accToken.SignedString(s.jwtSecret)
	if err != nil {
		return "", "", 0, fmt.Errorf("failed signing access token: %w", err)
	}

	// 2. Long-lived Refresh Token
	refreshClaims := JWTClaims{
		UserID:    user.ID,
		Username:  user.Username,
		Email:     user.Email,
		TokenType: TokenTypeRefresh,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   fmt.Sprintf("%d", user.ID),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(refreshExpiresAt),
			Issuer:    "backend-smtp",
		},
	}
	refToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	refreshToken, err = refToken.SignedString(s.jwtSecret)
	if err != nil {
		return "", "", 0, fmt.Errorf("failed signing refresh token: %w", err)
	}

	return accessToken, refreshToken, int64(s.accessExp.Seconds()), nil
}

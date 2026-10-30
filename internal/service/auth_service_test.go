package service

import (
	"context"
	"testing"
	"time"

	"github.com/arjunstein/backend-smtp/internal/model"
	"github.com/arjunstein/backend-smtp/internal/repository"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"
)

type mockUserRepo struct {
	users             map[uint]*model.User
	usersByIdentifier map[string]*model.User
	nextID            uint
}

func newMockUserRepo() *mockUserRepo {
	return &mockUserRepo{
		users:             make(map[uint]*model.User),
		usersByIdentifier: make(map[string]*model.User),
		nextID:            1,
	}
}

func (m *mockUserRepo) Create(ctx context.Context, user *model.User) error {
	user.ID = m.nextID
	m.nextID++
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()

	m.users[user.ID] = user
	m.usersByIdentifier[user.Username] = user
	m.usersByIdentifier[user.Email] = user
	return nil
}

func (m *mockUserRepo) FindByID(ctx context.Context, id uint) (*model.User, error) {
	user, exists := m.users[id]
	if !exists {
		return nil, repository.ErrUserNotFound
	}
	return user, nil
}

func (m *mockUserRepo) FindByUsernameOrEmail(ctx context.Context, identifier string) (*model.User, error) {
	user, exists := m.usersByIdentifier[identifier]
	if !exists {
		return nil, repository.ErrUserNotFound
	}
	return user, nil
}

func (m *mockUserRepo) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	_, exists := m.usersByIdentifier[email]
	return exists, nil
}

func (m *mockUserRepo) ExistsByUsername(ctx context.Context, username string) (bool, error) {
	_, exists := m.usersByIdentifier[username]
	return exists, nil
}

func (m *mockUserRepo) Count(ctx context.Context) (int64, error) {
	return int64(len(m.users)), nil
}

func TestAuthService_SeedInitialUser_AndLogin(t *testing.T) {
	repo := newMockUserRepo()
	svc := NewAuthService(repo, "test-jwt-secret-key-1234567890", 15, 7)
	ctx := context.Background()

	// 1. Seed initial single user
	seeded, err := svc.SeedInitialUser(ctx, "Administrator", "admin", "admin@example.com", "password123")
	assert.NoError(t, err)
	assert.NotNil(t, seeded)
	assert.Equal(t, "admin", seeded.Username)
	assert.Equal(t, "admin@example.com", seeded.Email)

	// Second seed call should safely do nothing (count > 0)
	seededAgain, err := svc.SeedInitialUser(ctx, "Other", "other", "other@example.com", "password456")
	assert.NoError(t, err)
	assert.Nil(t, seededAgain)

	// 2. Login with username
	loginReq := &model.LoginRequest{
		UsernameOrEmail: "admin",
		Password:        "password123",
	}
	loginResp, err := svc.Login(ctx, loginReq)
	assert.NoError(t, err)
	assert.NotNil(t, loginResp)
	assert.NotEmpty(t, loginResp.AccessToken)
	assert.NotEmpty(t, loginResp.RefreshToken)
	assert.Equal(t, "Bearer", loginResp.TokenType)
	assert.Equal(t, int64(15*60), loginResp.ExpiresIn)
	assert.Equal(t, "admin", loginResp.User.Username)

	// Verify access token
	claims, err := svc.ValidateToken(loginResp.AccessToken)
	assert.NoError(t, err)
	assert.Equal(t, seeded.ID, claims.UserID)
	assert.Equal(t, "admin", claims.Username)
	assert.Equal(t, TokenTypeAccess, claims.TokenType)

	// Reject refresh token on resource ValidateToken
	_, err = svc.ValidateToken(loginResp.RefreshToken)
	assert.ErrorIs(t, err, ErrInvalidTokenType)

	// 3. Refresh token
	refreshReq := &model.RefreshTokenRequest{
		RefreshToken: loginResp.RefreshToken,
	}
	refreshed, err := svc.RefreshToken(ctx, refreshReq)
	assert.NoError(t, err)
	assert.NotNil(t, refreshed)
	assert.NotEmpty(t, refreshed.AccessToken)
	assert.NotEmpty(t, refreshed.RefreshToken)
	assert.Equal(t, "Bearer", refreshed.TokenType)
	assert.Equal(t, int64(15*60), refreshed.ExpiresIn)

	// 4. GetProfile
	profile, err := svc.GetProfile(ctx, seeded.ID)
	assert.NoError(t, err)
	assert.Equal(t, "admin", profile.Username)
}

func TestAuthService_Login_InvalidCredentials(t *testing.T) {
	repo := newMockUserRepo()
	svc := NewAuthService(repo, "test-jwt-secret-key-1234567890", 15, 7)
	ctx := context.Background()

	hashed, _ := bcrypt.GenerateFromPassword([]byte("correctpassword"), bcrypt.DefaultCost)
	_ = repo.Create(ctx, &model.User{
		Name:     "User",
		Username: "existinguser",
		Email:    "user@example.com",
		Password: string(hashed),
	})

	// Wrong password
	_, err := svc.Login(ctx, &model.LoginRequest{
		UsernameOrEmail: "existinguser",
		Password:        "wrongpassword",
	})
	assert.ErrorIs(t, err, ErrInvalidCredentials)

	// Non-existent user
	_, err = svc.Login(ctx, &model.LoginRequest{
		UsernameOrEmail: "unknownuser",
		Password:        "any",
	})
	assert.ErrorIs(t, err, ErrInvalidCredentials)
}

func TestAuthService_RefreshToken_Invalid(t *testing.T) {
	repo := newMockUserRepo()
	svc := NewAuthService(repo, "test-jwt-secret-key-1234567890", 15, 7)
	ctx := context.Background()

	// Invalid token string
	_, err := svc.RefreshToken(ctx, &model.RefreshTokenRequest{
		RefreshToken: "not-a-valid-jwt",
	})
	assert.ErrorIs(t, err, ErrInvalidToken)
}

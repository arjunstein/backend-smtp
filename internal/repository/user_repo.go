package repository

import (
	"context"
	"errors"
	"strings"

	"github.com/arjunstein/backend-smtp/internal/model"
	"gorm.io/gorm"
)

var (
	ErrUserNotFound = errors.New("user not found")
)

// UserRepository defines database operations for user accounts.
type UserRepository interface {
	Create(ctx context.Context, user *model.User) error
	FindByID(ctx context.Context, id uint) (*model.User, error)
	FindByUsernameOrEmail(ctx context.Context, identifier string) (*model.User, error)
	ExistsByEmail(ctx context.Context, email string) (bool, error)
	ExistsByUsername(ctx context.Context, username string) (bool, error)
	Count(ctx context.Context) (int64, error)
}

type userRepository struct {
	db *gorm.DB
}

// NewUserRepository creates a new instance of UserRepository.
func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) Create(ctx context.Context, user *model.User) error {
	return r.db.WithContext(ctx).Create(user).Error
}

func (r *userRepository) FindByID(ctx context.Context, id uint) (*model.User, error) {
	var user model.User
	if err := r.db.WithContext(ctx).First(&user, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) FindByUsernameOrEmail(ctx context.Context, identifier string) (*model.User, error) {
	var user model.User
	cleanIdentifier := strings.TrimSpace(identifier)

	err := r.db.WithContext(ctx).
		Where("username = ? OR email = ?", cleanIdentifier, cleanIdentifier).
		First(&user).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	var count int64
	cleanEmail := strings.TrimSpace(email)
	if err := r.db.WithContext(ctx).Model(&model.User{}).Where("email = ?", cleanEmail).Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *userRepository) ExistsByUsername(ctx context.Context, username string) (bool, error) {
	var count int64
	cleanUsername := strings.TrimSpace(username)
	if err := r.db.WithContext(ctx).Model(&model.User{}).Where("username = ?", cleanUsername).Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *userRepository) Count(ctx context.Context) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&model.User{}).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

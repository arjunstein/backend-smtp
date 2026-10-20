package repository

import (
	"context"
	"errors"

	"github.com/arjunstein/backend-smtp/internal/model"
	"gorm.io/gorm"
)

var (
	ErrSMTPConfigNotFound = errors.New("smtp configuration not found")
)

// SMTPConfigRepository defines database operations for SMTP configurations.
type SMTPConfigRepository interface {
	WithTx(tx *gorm.DB) SMTPConfigRepository
	Create(ctx context.Context, cfg *model.SMTPConfig) error
	FindAll(ctx context.Context) ([]model.SMTPConfig, error)
	FindByID(ctx context.Context, id uint) (*model.SMTPConfig, error)
	FindByName(ctx context.Context, name string) (*model.SMTPConfig, error)
	FindDefault(ctx context.Context) (*model.SMTPConfig, error)
	Update(ctx context.Context, cfg *model.SMTPConfig) error
	Delete(ctx context.Context, id uint) error
	SetDefault(ctx context.Context, id uint) error
}

type smtpConfigRepository struct {
	db *gorm.DB
}

// NewSMTPConfigRepository creates a new instance of SMTPConfigRepository.
func NewSMTPConfigRepository(db *gorm.DB) SMTPConfigRepository {
	return &smtpConfigRepository{db: db}
}

func (r *smtpConfigRepository) WithTx(tx *gorm.DB) SMTPConfigRepository {
	return &smtpConfigRepository{db: tx}
}

func (r *smtpConfigRepository) Create(ctx context.Context, cfg *model.SMTPConfig) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if cfg.IsDefault {
			// Unset any previous default configs
			if err := tx.Model(&model.SMTPConfig{}).Where("is_default = ?", true).Update("is_default", false).Error; err != nil {
				return err
			}
		}
		return tx.Create(cfg).Error
	})
}

func (r *smtpConfigRepository) FindAll(ctx context.Context) ([]model.SMTPConfig, error) {
	var list []model.SMTPConfig
	if err := r.db.WithContext(ctx).Order("is_default DESC, id ASC").Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

func (r *smtpConfigRepository) FindByID(ctx context.Context, id uint) (*model.SMTPConfig, error) {
	var cfg model.SMTPConfig
	err := r.db.WithContext(ctx).First(&cfg, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrSMTPConfigNotFound
		}
		return nil, err
	}
	return &cfg, nil
}

func (r *smtpConfigRepository) FindByName(ctx context.Context, name string) (*model.SMTPConfig, error) {
	var cfg model.SMTPConfig
	err := r.db.WithContext(ctx).Where("name = ?", name).First(&cfg).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrSMTPConfigNotFound
		}
		return nil, err
	}
	return &cfg, nil
}

func (r *smtpConfigRepository) FindDefault(ctx context.Context) (*model.SMTPConfig, error) {
	var cfg model.SMTPConfig
	err := r.db.WithContext(ctx).Where("is_default = ? AND is_active = ?", true, true).First(&cfg).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// Fallback: pick first active config if no explicit default is marked
			errFallback := r.db.WithContext(ctx).Where("is_active = ?", true).Order("id ASC").First(&cfg).Error
			if errFallback != nil {
				if errors.Is(errFallback, gorm.ErrRecordNotFound) {
					return nil, ErrSMTPConfigNotFound
				}
				return nil, errFallback
			}
			return &cfg, nil
		}
		return nil, err
	}
	return &cfg, nil
}

func (r *smtpConfigRepository) Update(ctx context.Context, cfg *model.SMTPConfig) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if cfg.IsDefault {
			// Unset any previous default configs
			if err := tx.Model(&model.SMTPConfig{}).Where("id != ? AND is_default = ?", cfg.ID, true).Update("is_default", false).Error; err != nil {
				return err
			}
		}
		return tx.Save(cfg).Error
	})
}

func (r *smtpConfigRepository) Delete(ctx context.Context, id uint) error {
	res := r.db.WithContext(ctx).Delete(&model.SMTPConfig{}, id)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrSMTPConfigNotFound
	}
	return nil
}

func (r *smtpConfigRepository) SetDefault(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var cfg model.SMTPConfig
		if err := tx.First(&cfg, id).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrSMTPConfigNotFound
			}
			return err
		}

		if err := tx.Model(&model.SMTPConfig{}).Where("id != ?", id).Update("is_default", false).Error; err != nil {
			return err
		}

		return tx.Model(&cfg).Updates(map[string]any{
			"is_default": true,
			"is_active":  true,
		}).Error
	})
}

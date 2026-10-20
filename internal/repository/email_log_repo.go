package repository

import (
	"context"

	"github.com/arjunstein/backend-smtp/internal/model"
	"gorm.io/gorm"
)

// EmailLogRepository defines database operations for email audit logs.
type EmailLogRepository interface {
	WithTx(tx *gorm.DB) EmailLogRepository
	Create(ctx context.Context, log *model.EmailLog) error
	FindAll(ctx context.Context, limit int, offset int) ([]model.EmailLog, int64, error)
	FindByID(ctx context.Context, id uint) (*model.EmailLog, error)
}

type emailLogRepository struct {
	db *gorm.DB
}

// NewEmailLogRepository creates a new instance of EmailLogRepository.
func NewEmailLogRepository(db *gorm.DB) EmailLogRepository {
	return &emailLogRepository{db: db}
}

func (r *emailLogRepository) WithTx(tx *gorm.DB) EmailLogRepository {
	return &emailLogRepository{db: tx}
}

func (r *emailLogRepository) Create(ctx context.Context, log *model.EmailLog) error {
	return r.db.WithContext(ctx).Create(log).Error
}

func (r *emailLogRepository) FindAll(ctx context.Context, limit int, offset int) ([]model.EmailLog, int64, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}

	var total int64
	if err := r.db.WithContext(ctx).Model(&model.EmailLog{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var logs []model.EmailLog
	err := r.db.WithContext(ctx).
		Preload("SMTPConfig").
		Order("id DESC").
		Limit(limit).
		Offset(offset).
		Find(&logs).Error
	if err != nil {
		return nil, 0, err
	}

	return logs, total, nil
}

func (r *emailLogRepository) FindByID(ctx context.Context, id uint) (*model.EmailLog, error) {
	var log model.EmailLog
	err := r.db.WithContext(ctx).Preload("SMTPConfig").First(&log, id).Error
	if err != nil {
		return nil, err
	}
	return &log, nil
}

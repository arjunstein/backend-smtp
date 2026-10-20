package service

import (
	"context"
	"errors"
	"testing"

	"github.com/arjunstein/backend-smtp/internal/model"
	"github.com/arjunstein/backend-smtp/internal/repository"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

// MockEmailLogRepository implements repository.EmailLogRepository
type MockEmailLogRepository struct {
	logs   map[uint]*model.EmailLog
	nextID uint
}

func NewMockEmailLogRepository() *MockEmailLogRepository {
	return &MockEmailLogRepository{
		logs:   make(map[uint]*model.EmailLog),
		nextID: 1,
	}
}

func (m *MockEmailLogRepository) WithTx(tx *gorm.DB) repository.EmailLogRepository {
	return m
}

func (m *MockEmailLogRepository) Create(ctx context.Context, log *model.EmailLog) error {
	log.ID = m.nextID
	m.nextID++
	m.logs[log.ID] = log
	return nil
}

func (m *MockEmailLogRepository) FindAll(ctx context.Context, limit int, offset int) ([]model.EmailLog, int64, error) {
	var list []model.EmailLog
	for _, l := range m.logs {
		list = append(list, *l)
	}
	return list, int64(len(list)), nil
}

func (m *MockEmailLogRepository) FindByID(ctx context.Context, id uint) (*model.EmailLog, error) {
	l, ok := m.logs[id]
	if !ok {
		return nil, gorm.ErrRecordNotFound
	}
	return l, nil
}

func TestEmailService_SendEmail_FallbackToDefault(t *testing.T) {
	smtpRepo := NewMockSMTPConfigRepository()
	logRepo := NewMockEmailLogRepository()
	mockMail := &MockMailer{}
	svc := NewEmailService(smtpRepo, logRepo, mockMail)

	// Add default active config
	cfg := &model.SMTPConfig{
		Name:       "Default SMTP",
		Host:       "smtp.default.com",
		Port:       587,
		Username:   "default",
		Password:   "pass",
		FromEmail:  "default@example.com",
		Encryption: "STARTTLS",
		IsActive:   true,
		IsDefault:  true,
	}
	_ = smtpRepo.Create(context.Background(), cfg)

	req := &model.SendEmailRequest{
		To:      []string{"recipient@example.com"},
		Subject: "Welcome!",
		Body:    "Welcome to our service.",
		IsHTML:  false,
	}

	res, err := svc.SendEmail(context.Background(), req)
	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.Equal(t, model.EmailStatusSuccess, res.Status)
	assert.Equal(t, "Default SMTP", res.SMTPName)

	// Verify mailer was called with default config
	assert.Equal(t, cfg.Host, mockMail.SentCfg.Host)
	assert.Equal(t, "Welcome!", mockMail.SentReq.Subject)

	// Verify audit log was recorded
	logs, total, err := logRepo.FindAll(context.Background(), 10, 0)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Equal(t, model.EmailStatusSuccess, logs[0].Status)
	assert.Equal(t, "recipient@example.com", logs[0].To)
}

func TestEmailService_SendEmail_FailureLogsError(t *testing.T) {
	smtpRepo := NewMockSMTPConfigRepository()
	logRepo := NewMockEmailLogRepository()
	mockMail := &MockMailer{
		SendErr: errors.New("auth authentication rejected"),
	}
	svc := NewEmailService(smtpRepo, logRepo, mockMail)

	cfg := &model.SMTPConfig{
		Name:       "Default SMTP",
		Host:       "smtp.default.com",
		Port:       587,
		FromEmail:  "default@example.com",
		Encryption: "STARTTLS",
		IsActive:   true,
		IsDefault:  true,
	}
	_ = smtpRepo.Create(context.Background(), cfg)

	req := &model.SendEmailRequest{
		To:      []string{"client@example.com"},
		Subject: "Invoice",
		Body:    "Please pay invoice.",
	}

	res, err := svc.SendEmail(context.Background(), req)
	assert.Error(t, err)
	assert.Nil(t, res)

	// Verify audit log recorded failure
	logs, total, _ := logRepo.FindAll(context.Background(), 10, 0)
	assert.Equal(t, int64(1), total)
	assert.Equal(t, model.EmailStatusFailed, logs[0].Status)
	assert.Contains(t, logs[0].ErrorMessage, "auth authentication rejected")
}

func TestEmailService_SendEmail_NoActiveConfig(t *testing.T) {
	smtpRepo := NewMockSMTPConfigRepository()
	logRepo := NewMockEmailLogRepository()
	mockMail := &MockMailer{}
	svc := NewEmailService(smtpRepo, logRepo, mockMail)

	req := &model.SendEmailRequest{
		To:      []string{"client@example.com"},
		Subject: "Notice",
		Body:    "System update.",
	}

	res, err := svc.SendEmail(context.Background(), req)
	assert.ErrorIs(t, err, ErrNoActiveSMTPConfig)
	assert.Nil(t, res)
}

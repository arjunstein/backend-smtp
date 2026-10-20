package service

import (
	"context"
	"errors"
	"testing"

	"github.com/arjunstein/backend-smtp/internal/model"
	"github.com/stretchr/testify/assert"
)

func TestEmailService_SendEmail_FallbackToDefault(t *testing.T) {
	smtpRepo := NewMockSMTPConfigRepository()
	mockMail := &MockMailer{}
	svc := NewEmailService(smtpRepo, mockMail)

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
	assert.Equal(t, "SUCCESS", res.Status)
	assert.Equal(t, "Default SMTP", res.SMTPName)

	// Verify mailer was called with default config
	assert.Equal(t, cfg.Host, mockMail.SentCfg.Host)
	assert.Equal(t, "Welcome!", mockMail.SentReq.Subject)
}

func TestEmailService_SendEmail_FailureReturnsError(t *testing.T) {
	smtpRepo := NewMockSMTPConfigRepository()
	mockMail := &MockMailer{
		SendErr: errors.New("auth authentication rejected"),
	}
	svc := NewEmailService(smtpRepo, mockMail)

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
	assert.Contains(t, err.Error(), "auth authentication rejected")
}

func TestEmailService_SendEmail_NoActiveConfig(t *testing.T) {
	smtpRepo := NewMockSMTPConfigRepository()
	mockMail := &MockMailer{}
	svc := NewEmailService(smtpRepo, mockMail)

	req := &model.SendEmailRequest{
		To:      []string{"client@example.com"},
		Subject: "Notice",
		Body:    "System update.",
	}

	res, err := svc.SendEmail(context.Background(), req)
	assert.ErrorIs(t, err, ErrNoActiveSMTPConfig)
	assert.Nil(t, res)
}

package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/arjunstein/backend-smtp/internal/model"
	"github.com/arjunstein/backend-smtp/internal/repository"
	"github.com/arjunstein/backend-smtp/internal/service/mailer"
)

var (
	ErrNoActiveSMTPConfig = errors.New("no active default SMTP configuration found")
	ErrSMTPConfigInactive = errors.New("specified SMTP configuration is inactive")
)

// EmailService defines operations for central email sending and log retrieval.
type EmailService interface {
	SendEmail(ctx context.Context, req *model.SendEmailRequest) (*model.SendEmailResponse, error)
	GetEmailLogs(ctx context.Context, limit int, offset int) ([]model.EmailLog, int64, error)
	GetEmailLogByID(ctx context.Context, id uint) (*model.EmailLog, error)
}

type emailService struct {
	smtpRepo repository.SMTPConfigRepository
	logRepo  repository.EmailLogRepository
	mailer   mailer.Mailer
}

// NewEmailService instantiates EmailService.
func NewEmailService(
	smtpRepo repository.SMTPConfigRepository,
	logRepo repository.EmailLogRepository,
	mailer mailer.Mailer,
) EmailService {
	return &emailService{
		smtpRepo: smtpRepo,
		logRepo:  logRepo,
		mailer:   mailer,
	}
}

func (s *emailService) SendEmail(ctx context.Context, req *model.SendEmailRequest) (*model.SendEmailResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	// 1. Resolve SMTP Configuration
	var (
		cfg *model.SMTPConfig
		err error
	)

	if req.SMTPConfigID != nil && *req.SMTPConfigID > 0 {
		cfg, err = s.smtpRepo.FindByID(ctx, *req.SMTPConfigID)
		if err != nil {
			return nil, fmt.Errorf("failed to locate smtp config id %d: %w", *req.SMTPConfigID, err)
		}
		if !cfg.IsActive {
			return nil, ErrSMTPConfigInactive
		}
	} else {
		cfg, err = s.smtpRepo.FindDefault(ctx)
		if err != nil {
			return nil, ErrNoActiveSMTPConfig
		}
	}

	// 2. Prepare Audit Log
	emailLog := &model.EmailLog{
		SMTPConfigID: &cfg.ID,
		To:           strings.Join(req.To, ", "),
		CC:           strings.Join(req.CC, ", "),
		BCC:          strings.Join(req.BCC, ", "),
		Subject:      req.Subject,
		SentAt:       time.Now(),
	}

	// 3. Dispatch Email via Mailer
	sendErr := s.mailer.Send(cfg, req)
	if sendErr != nil {
		emailLog.Status = model.EmailStatusFailed
		emailLog.ErrorMessage = sendErr.Error()
		_ = s.logRepo.Create(ctx, emailLog)
		return nil, fmt.Errorf("failed to transmit email: %w", sendErr)
	}

	emailLog.Status = model.EmailStatusSuccess
	if err := s.logRepo.Create(ctx, emailLog); err != nil {
		// Log failed to persist audit record, but email was successfully sent
		// Don't fail the client request, but note in logs
	}

	return &model.SendEmailResponse{
		Status:       model.EmailStatusSuccess,
		SMTPConfigID: cfg.ID,
		SMTPName:     cfg.Name,
		Recipients:   req.To,
		Subject:      req.Subject,
		SentAt:       emailLog.SentAt,
	}, nil
}

func (s *emailService) GetEmailLogs(ctx context.Context, limit int, offset int) ([]model.EmailLog, int64, error) {
	return s.logRepo.FindAll(ctx, limit, offset)
}

func (s *emailService) GetEmailLogByID(ctx context.Context, id uint) (*model.EmailLog, error) {
	return s.logRepo.FindByID(ctx, id)
}

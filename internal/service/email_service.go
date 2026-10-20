package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
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

// EmailService defines operations for central email sending.
type EmailService interface {
	SendEmail(ctx context.Context, req *model.SendEmailRequest) (*model.SendEmailResponse, error)
}

type emailService struct {
	smtpRepo repository.SMTPConfigRepository
	mailer   mailer.Mailer
}

// NewEmailService instantiates EmailService.
func NewEmailService(
	smtpRepo repository.SMTPConfigRepository,
	mailer mailer.Mailer,
) EmailService {
	return &emailService{
		smtpRepo: smtpRepo,
		mailer:   mailer,
	}
}

func (s *emailService) SendEmail(ctx context.Context, req *model.SendEmailRequest) (*model.SendEmailResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	startTime := time.Now()

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

	// 2. Dispatch Email via Mailer
	sendErr := s.mailer.Send(cfg, req)
	duration := time.Since(startTime)

	if sendErr != nil {
		slog.ErrorContext(ctx, "email delivery failed",
			slog.String("event", "email_dispatch"),
			slog.String("status", "FAILED"),
			slog.Uint64("smtp_config_id", uint64(cfg.ID)),
			slog.String("smtp_name", cfg.Name),
			slog.String("recipients", strings.Join(req.To, ", ")),
			slog.String("subject", req.Subject),
			slog.Duration("duration", duration),
			slog.String("error", sendErr.Error()),
		)
		return nil, fmt.Errorf("failed to transmit email: %w", sendErr)
	}

	slog.InfoContext(ctx, "email dispatched successfully",
		slog.String("event", "email_dispatch"),
		slog.String("status", "SUCCESS"),
		slog.Uint64("smtp_config_id", uint64(cfg.ID)),
		slog.String("smtp_name", cfg.Name),
		slog.String("recipients", strings.Join(req.To, ", ")),
		slog.String("subject", req.Subject),
		slog.Duration("duration", duration),
	)

	return &model.SendEmailResponse{
		Status:       "SUCCESS",
		SMTPConfigID: cfg.ID,
		SMTPName:     cfg.Name,
		Recipients:   req.To,
		Subject:      req.Subject,
		SentAt:       time.Now(),
	}, nil
}

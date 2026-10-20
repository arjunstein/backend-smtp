package service

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/arjunstein/backend-smtp/internal/model"
	"github.com/arjunstein/backend-smtp/internal/repository"
	"github.com/arjunstein/backend-smtp/internal/service/mailer"
)

var (
	ErrDuplicateConfigName = errors.New("smtp configuration name already exists")
)

// SMTPConfigService defines business operations for managing SMTP configs.
type SMTPConfigService interface {
	Create(ctx context.Context, req *model.CreateSMTPConfigRequest) (*model.SMTPConfigResponse, error)
	FindAll(ctx context.Context) ([]model.SMTPConfigResponse, error)
	FindByID(ctx context.Context, id uint) (*model.SMTPConfigResponse, error)
	Update(ctx context.Context, id uint, req *model.UpdateSMTPConfigRequest) (*model.SMTPConfigResponse, error)
	Delete(ctx context.Context, id uint) error
	TestConnection(ctx context.Context, id uint) error
	SetDefault(ctx context.Context, id uint) error
}

type smtpConfigService struct {
	repo   repository.SMTPConfigRepository
	mailer mailer.Mailer
}

// NewSMTPConfigService instantiates SMTPConfigService.
func NewSMTPConfigService(repo repository.SMTPConfigRepository, mailer mailer.Mailer) SMTPConfigService {
	return &smtpConfigService{
		repo:   repo,
		mailer: mailer,
	}
}

func (s *smtpConfigService) Create(ctx context.Context, req *model.CreateSMTPConfigRequest) (*model.SMTPConfigResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	// Check name uniqueness
	existing, _ := s.repo.FindByName(ctx, strings.TrimSpace(req.Name))
	if existing != nil {
		return nil, ErrDuplicateConfigName
	}

	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}

	isDefault := false
	if req.IsDefault != nil {
		isDefault = *req.IsDefault
	}

	cfg := &model.SMTPConfig{
		Name:       strings.TrimSpace(req.Name),
		Host:       strings.TrimSpace(req.Host),
		Port:       req.Port,
		Username:   strings.TrimSpace(req.Username),
		Password:   req.Password,
		FromEmail:  strings.TrimSpace(req.FromEmail),
		FromName:   strings.TrimSpace(req.FromName),
		Encryption: strings.ToUpper(strings.TrimSpace(req.Encryption)),
		IsActive:   isActive,
		IsDefault:  isDefault,
	}

	if err := s.repo.Create(ctx, cfg); err != nil {
		return nil, fmt.Errorf("failed to save smtp config: %w", err)
	}

	res := cfg.ToResponse()
	return &res, nil
}

func (s *smtpConfigService) FindAll(ctx context.Context) ([]model.SMTPConfigResponse, error) {
	configs, err := s.repo.FindAll(ctx)
	if err != nil {
		return nil, err
	}

	responses := make([]model.SMTPConfigResponse, len(configs))
	for i, c := range configs {
		responses[i] = c.ToResponse()
	}
	return responses, nil
}

func (s *smtpConfigService) FindByID(ctx context.Context, id uint) (*model.SMTPConfigResponse, error) {
	cfg, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	res := cfg.ToResponse()
	return &res, nil
}

func (s *smtpConfigService) Update(ctx context.Context, id uint, req *model.UpdateSMTPConfigRequest) (*model.SMTPConfigResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	cfg, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if req.Name != nil {
		name := strings.TrimSpace(*req.Name)
		if name != cfg.Name {
			existing, _ := s.repo.FindByName(ctx, name)
			if existing != nil && existing.ID != cfg.ID {
				return nil, ErrDuplicateConfigName
			}
			cfg.Name = name
		}
	}

	if req.Host != nil {
		cfg.Host = strings.TrimSpace(*req.Host)
	}
	if req.Port != nil {
		cfg.Port = *req.Port
	}
	if req.Username != nil {
		cfg.Username = strings.TrimSpace(*req.Username)
	}
	if req.Password != nil && strings.TrimSpace(*req.Password) != "" {
		cfg.Password = *req.Password
	}
	if req.FromEmail != nil {
		cfg.FromEmail = strings.TrimSpace(*req.FromEmail)
	}
	if req.FromName != nil {
		cfg.FromName = strings.TrimSpace(*req.FromName)
	}
	if req.Encryption != nil {
		cfg.Encryption = strings.ToUpper(strings.TrimSpace(*req.Encryption))
	}
	if req.IsActive != nil {
		cfg.IsActive = *req.IsActive
	}
	if req.IsDefault != nil {
		cfg.IsDefault = *req.IsDefault
	}

	if err := s.repo.Update(ctx, cfg); err != nil {
		return nil, fmt.Errorf("failed to update smtp config: %w", err)
	}

	res := cfg.ToResponse()
	return &res, nil
}

func (s *smtpConfigService) Delete(ctx context.Context, id uint) error {
	return s.repo.Delete(ctx, id)
}

func (s *smtpConfigService) SetDefault(ctx context.Context, id uint) error {
	return s.repo.SetDefault(ctx, id)
}

func (s *smtpConfigService) TestConnection(ctx context.Context, id uint) error {
	cfg, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return err
	}

	return s.mailer.Ping(cfg)
}

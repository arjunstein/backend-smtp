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

// MockSMTPConfigRepository implements repository.SMTPConfigRepository for testing
type MockSMTPConfigRepository struct {
	configs        map[uint]*model.SMTPConfig
	nextID         uint
	createErr      error
	findByIDErr    error
	findByNameErr  error
	findDefaultErr error
}

func NewMockSMTPConfigRepository() *MockSMTPConfigRepository {
	return &MockSMTPConfigRepository{
		configs: make(map[uint]*model.SMTPConfig),
		nextID:  1,
	}
}

func (m *MockSMTPConfigRepository) WithTx(tx *gorm.DB) repository.SMTPConfigRepository {
	return m
}

func (m *MockSMTPConfigRepository) Create(ctx context.Context, cfg *model.SMTPConfig) error {
	if m.createErr != nil {
		return m.createErr
	}
	cfg.ID = m.nextID
	m.nextID++
	m.configs[cfg.ID] = cfg
	return nil
}

func (m *MockSMTPConfigRepository) FindAll(ctx context.Context) ([]model.SMTPConfig, error) {
	var list []model.SMTPConfig
	for _, c := range m.configs {
		list = append(list, *c)
	}
	return list, nil
}

func (m *MockSMTPConfigRepository) FindByID(ctx context.Context, id uint) (*model.SMTPConfig, error) {
	if m.findByIDErr != nil {
		return nil, m.findByIDErr
	}
	cfg, ok := m.configs[id]
	if !ok {
		return nil, repository.ErrSMTPConfigNotFound
	}
	return cfg, nil
}

func (m *MockSMTPConfigRepository) FindByName(ctx context.Context, name string) (*model.SMTPConfig, error) {
	if m.findByNameErr != nil {
		return nil, m.findByNameErr
	}
	for _, c := range m.configs {
		if c.Name == name {
			return c, nil
		}
	}
	return nil, repository.ErrSMTPConfigNotFound
}

func (m *MockSMTPConfigRepository) FindDefault(ctx context.Context) (*model.SMTPConfig, error) {
	if m.findDefaultErr != nil {
		return nil, m.findDefaultErr
	}
	for _, c := range m.configs {
		if c.IsDefault && c.IsActive {
			return c, nil
		}
	}
	for _, c := range m.configs {
		if c.IsActive {
			return c, nil
		}
	}
	return nil, repository.ErrSMTPConfigNotFound
}

func (m *MockSMTPConfigRepository) Update(ctx context.Context, cfg *model.SMTPConfig) error {
	m.configs[cfg.ID] = cfg
	return nil
}

func (m *MockSMTPConfigRepository) Delete(ctx context.Context, id uint) error {
	if _, ok := m.configs[id]; !ok {
		return repository.ErrSMTPConfigNotFound
	}
	delete(m.configs, id)
	return nil
}

func (m *MockSMTPConfigRepository) SetDefault(ctx context.Context, id uint) error {
	cfg, ok := m.configs[id]
	if !ok {
		return repository.ErrSMTPConfigNotFound
	}
	for _, c := range m.configs {
		c.IsDefault = false
	}
	cfg.IsDefault = true
	cfg.IsActive = true
	return nil
}

// MockMailer implements mailer.Mailer for testing
type MockMailer struct {
	SendErr error
	PingErr error
	SentReq *model.SendEmailRequest
	SentCfg *model.SMTPConfig
}

func (m *MockMailer) Send(cfg *model.SMTPConfig, req *model.SendEmailRequest) error {
	m.SentCfg = cfg
	m.SentReq = req
	return m.SendErr
}

func (m *MockMailer) Ping(cfg *model.SMTPConfig) error {
	return m.PingErr
}

func TestSMTPConfigService_Create_Success(t *testing.T) {
	repo := NewMockSMTPConfigRepository()
	mockMail := &MockMailer{}
	svc := NewSMTPConfigService(repo, mockMail)

	req := &model.CreateSMTPConfigRequest{
		Name:       "Primary Mailer",
		Host:       "smtp.example.com",
		Port:       587,
		Username:   "user@example.com",
		Password:   "secret",
		FromEmail:  "user@example.com",
		FromName:   "Primary",
		Encryption: "STARTTLS",
	}

	res, err := svc.Create(context.Background(), req)
	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.Equal(t, uint(1), res.ID)
	assert.Equal(t, "Primary Mailer", res.Name)
	assert.True(t, res.IsActive)
}

func TestSMTPConfigService_Create_Duplicate(t *testing.T) {
	repo := NewMockSMTPConfigRepository()
	mockMail := &MockMailer{}
	svc := NewSMTPConfigService(repo, mockMail)

	req := &model.CreateSMTPConfigRequest{
		Name:       "Primary Mailer",
		Host:       "smtp.example.com",
		Port:       587,
		Username:   "user@example.com",
		Password:   "secret",
		FromEmail:  "user@example.com",
		FromName:   "Primary",
		Encryption: "STARTTLS",
	}

	_, err := svc.Create(context.Background(), req)
	assert.NoError(t, err)

	_, errDuplicate := svc.Create(context.Background(), req)
	assert.ErrorIs(t, errDuplicate, ErrDuplicateConfigName)
}

func TestSMTPConfigService_TestConnection(t *testing.T) {
	repo := NewMockSMTPConfigRepository()
	mockMail := &MockMailer{}
	svc := NewSMTPConfigService(repo, mockMail)

	req := &model.CreateSMTPConfigRequest{
		Name:       "Test Server",
		Host:       "smtp.test.com",
		Port:       587,
		Username:   "test",
		Password:   "pass",
		FromEmail:  "test@test.com",
		Encryption: "STARTTLS",
	}

	res, _ := svc.Create(context.Background(), req)

	// Successful ping
	err := svc.TestConnection(context.Background(), res.ID)
	assert.NoError(t, err)

	// Failed ping
	mockMail.PingErr = errors.New("connection timeout")
	errFail := svc.TestConnection(context.Background(), res.ID)
	assert.Error(t, errFail)
	assert.Contains(t, errFail.Error(), "connection timeout")
}

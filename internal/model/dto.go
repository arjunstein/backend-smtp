package model

import (
	"errors"
	"net/mail"
	"strings"
	"time"
)

// CreateSMTPConfigRequest payload to register new SMTP configuration.
type CreateSMTPConfigRequest struct {
	Name       string `json:"name" example:"Gmail Service"`
	Host       string `json:"host" example:"smtp.gmail.com"`
	Port       int    `json:"port" example:"587"`
	Username   string `json:"username" example:"no-reply@example.com"`
	Password   string `json:"password" example:"secret123"`
	FromEmail  string `json:"from_email" example:"no-reply@example.com"`
	FromName   string `json:"from_name" example:"Notification Agent"`
	Encryption string `json:"encryption" example:"STARTTLS"`
	IsActive   *bool  `json:"is_active,omitempty" example:"true"`
	IsDefault  *bool  `json:"is_default,omitempty" example:"false"`
}

// Validate checks request validity.
func (r *CreateSMTPConfigRequest) Validate() error {
	if strings.TrimSpace(r.Name) == "" {
		return errors.New("name is required")
	}
	if strings.TrimSpace(r.Host) == "" {
		return errors.New("host is required")
	}
	if r.Port <= 0 || r.Port > 65535 {
		return errors.New("port must be between 1 and 65535")
	}
	if strings.TrimSpace(r.Username) == "" {
		return errors.New("username is required")
	}
	if strings.TrimSpace(r.Password) == "" {
		return errors.New("password is required")
	}
	if strings.TrimSpace(r.FromEmail) == "" {
		return errors.New("from_email is required")
	}
	if _, err := mail.ParseAddress(r.FromEmail); err != nil {
		return errors.New("from_email is not a valid email address")
	}

	enc := strings.ToUpper(strings.TrimSpace(r.Encryption))
	if enc == "" {
		enc = EncryptionSTARTTLS
		r.Encryption = enc
	}
	if enc != EncryptionNone && enc != EncryptionSTARTTLS && enc != EncryptionSSL && enc != EncryptionTLS {
		return errors.New("encryption must be STARTTLS, SSL, TLS, or NONE")
	}

	return nil
}

// UpdateSMTPConfigRequest payload to update an existing SMTP configuration.
type UpdateSMTPConfigRequest struct {
	Name       *string `json:"name,omitempty" example:"Updated Service Name"`
	Host       *string `json:"host,omitempty" example:"smtp.mailtrap.io"`
	Port       *int    `json:"port,omitempty" example:"587"`
	Username   *string `json:"username,omitempty" example:"user@mailtrap.io"`
	Password   *string `json:"password,omitempty" example:"newsecret123"`
	FromEmail  *string `json:"from_email,omitempty" example:"notifications@mailtrap.io"`
	FromName   *string `json:"from_name,omitempty" example:"Central Mailer"`
	Encryption *string `json:"encryption,omitempty" example:"STARTTLS"`
	IsActive   *bool   `json:"is_active,omitempty" example:"true"`
	IsDefault  *bool   `json:"is_default,omitempty" example:"true"`
}

// Validate validates update fields if provided.
func (r *UpdateSMTPConfigRequest) Validate() error {
	if r.Name != nil && strings.TrimSpace(*r.Name) == "" {
		return errors.New("name cannot be empty")
	}
	if r.Host != nil && strings.TrimSpace(*r.Host) == "" {
		return errors.New("host cannot be empty")
	}
	if r.Port != nil && (*r.Port <= 0 || *r.Port > 65535) {
		return errors.New("port must be between 1 and 65535")
	}
	if r.FromEmail != nil {
		if strings.TrimSpace(*r.FromEmail) == "" {
			return errors.New("from_email cannot be empty")
		}
		if _, err := mail.ParseAddress(*r.FromEmail); err != nil {
			return errors.New("from_email is not a valid email address")
		}
	}
	if r.Encryption != nil {
		enc := strings.ToUpper(strings.TrimSpace(*r.Encryption))
		if enc != EncryptionNone && enc != EncryptionSTARTTLS && enc != EncryptionSSL && enc != EncryptionTLS {
			return errors.New("encryption must be STARTTLS, SSL, TLS, or NONE")
		}
	}
	return nil
}

// SMTPConfigResponse sanitized response for SMTP configurations without exposing plain password.
type SMTPConfigResponse struct {
	ID         uint      `json:"id" example:"1"`
	Name       string    `json:"name" example:"Gmail Service"`
	Host       string    `json:"host" example:"smtp.gmail.com"`
	Port       int       `json:"port" example:"587"`
	Username   string    `json:"username" example:"no-reply@example.com"`
	FromEmail  string    `json:"from_email" example:"no-reply@example.com"`
	FromName   string    `json:"from_name" example:"Notification Agent"`
	Encryption string    `json:"encryption" example:"STARTTLS"`
	IsActive   bool      `json:"is_active" example:"true"`
	IsDefault  bool      `json:"is_default" example:"false"`
	CreatedAt  time.Time `json:"created_at" example:"2026-10-20T10:00:00Z"`
	UpdatedAt  time.Time `json:"updated_at" example:"2026-10-20T10:00:00Z"`
}

// ToResponse converts SMTPConfig entity to safe client DTO.
func (c *SMTPConfig) ToResponse() SMTPConfigResponse {
	return SMTPConfigResponse{
		ID:         c.ID,
		Name:       c.Name,
		Host:       c.Host,
		Port:       c.Port,
		Username:   c.Username,
		FromEmail:  c.FromEmail,
		FromName:   c.FromName,
		Encryption: c.Encryption,
		IsActive:   c.IsActive,
		IsDefault:  c.IsDefault,
		CreatedAt:  c.CreatedAt,
		UpdatedAt:  c.UpdatedAt,
	}
}

// AttachmentDTO represents an attached file encoded in base64.
type AttachmentDTO struct {
	Filename      string `json:"filename" example:"invoice.pdf"`
	ContentType   string `json:"content_type" example:"application/pdf"`
	ContentBase64 string `json:"content_base64" example:"JVBERi0xLjQK..."`
}

// SendEmailRequest payload for dispatching email through the centralized service.
type SendEmailRequest struct {
	SMTPConfigID *uint           `json:"smtp_config_id,omitempty" example:"1"`
	To           []string        `json:"to" example:"recipient@example.com"`
	CC           []string        `json:"cc,omitempty" example:"cc@example.com"`
	BCC          []string        `json:"bcc,omitempty" example:"bcc@example.com"`
	Subject      string          `json:"subject" example:"Your verification code"`
	Body         string          `json:"body" example:"<h1>Hello!</h1><p>Your code is 123456</p>"`
	IsHTML       bool            `json:"is_html" example:"true"`
	Attachments  []AttachmentDTO `json:"attachments,omitempty"`
}

// Validate checks send email request.
func (r *SendEmailRequest) Validate() error {
	if len(r.To) == 0 {
		return errors.New("at least one 'to' recipient email is required")
	}
	for _, to := range r.To {
		if strings.TrimSpace(to) == "" {
			return errors.New("recipient email cannot be empty")
		}
		if _, err := mail.ParseAddress(strings.TrimSpace(to)); err != nil {
			return errors.New("invalid recipient email: " + to)
		}
	}
	for _, cc := range r.CC {
		if strings.TrimSpace(cc) != "" {
			if _, err := mail.ParseAddress(strings.TrimSpace(cc)); err != nil {
				return errors.New("invalid cc email: " + cc)
			}
		}
	}
	for _, bcc := range r.BCC {
		if strings.TrimSpace(bcc) != "" {
			if _, err := mail.ParseAddress(strings.TrimSpace(bcc)); err != nil {
				return errors.New("invalid bcc email: " + bcc)
			}
		}
	}
	if strings.TrimSpace(r.Subject) == "" {
		return errors.New("subject is required")
	}
	if strings.TrimSpace(r.Body) == "" {
		return errors.New("body is required")
	}
	return nil
}

// SendEmailResponse result of email delivery.
type SendEmailResponse struct {
	Status       string    `json:"status" example:"SUCCESS"`
	SMTPConfigID uint      `json:"smtp_config_id" example:"1"`
	SMTPName     string    `json:"smtp_name" example:"Gmail Service"`
	Recipients   []string  `json:"recipients" example:"recipient@example.com"`
	Subject      string    `json:"subject" example:"Your verification code"`
	SentAt       time.Time `json:"sent_at" example:"2026-10-20T10:00:00Z"`
}

// TestConnectionResponse payload for SMTP ping verification.
type TestConnectionResponse struct {
	Status string `json:"status" example:"CONNECTED"`
}


// LoginRequest payload for user authentication.
type LoginRequest struct {
	UsernameOrEmail string `json:"username_or_email" example:"arjun"`
	Password        string `json:"password" example:"secret123"`
}

// Validate validates LoginRequest fields.
func (r *LoginRequest) Validate() error {
	if strings.TrimSpace(r.UsernameOrEmail) == "" {
		return errors.New("username or email is required")
	}
	if strings.TrimSpace(r.Password) == "" {
		return errors.New("password is required")
	}
	return nil
}

// UserResponse represents safe user account response.
type UserResponse struct {
	ID        uint      `json:"id" example:"1"`
	Name      string    `json:"name" example:"Arjun Gunawan"`
	Username  string    `json:"username" example:"arjun"`
	Email     string    `json:"email" example:"arjun@example.com"`
	CreatedAt time.Time `json:"created_at" example:"2026-10-20T10:00:00Z"`
}

// ToUserResponse converts User entity to UserResponse DTO.
func ToUserResponse(user *User) UserResponse {
	return UserResponse{
		ID:        user.ID,
		Name:      user.Name,
		Username:  user.Username,
		Email:     user.Email,
		CreatedAt: user.CreatedAt,
	}
}

// AuthResponse payload returned upon successful login or registration.
type AuthResponse struct {
	AccessToken  string       `json:"access_token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
	RefreshToken string       `json:"refresh_token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
	TokenType    string       `json:"token_type" example:"Bearer"`
	ExpiresIn    int64        `json:"expires_in" example:"3600"`
	User         UserResponse `json:"user"`
}

// RefreshTokenRequest payload for renewing an expired access token.
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
}

// Validate validates RefreshTokenRequest fields.
func (r *RefreshTokenRequest) Validate() error {
	if strings.TrimSpace(r.RefreshToken) == "" {
		return errors.New("refresh_token is required")
	}
	return nil
}

// RefreshTokenResponse payload containing renewed access & refresh tokens.
type RefreshTokenResponse struct {
	AccessToken  string `json:"access_token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
	RefreshToken string `json:"refresh_token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
	TokenType    string `json:"token_type" example:"Bearer"`
	ExpiresIn    int64  `json:"expires_in" example:"3600"`
}



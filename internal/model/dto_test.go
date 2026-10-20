package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateSMTPConfigRequest_Validate(t *testing.T) {
	tests := []struct {
		name    string
		req     CreateSMTPConfigRequest
		wantErr bool
		errMsg  string
	}{
		{
			name: "Valid request",
			req: CreateSMTPConfigRequest{
				Name:       "Primary Mailer",
				Host:       "smtp.gmail.com",
				Port:       587,
				Username:   "user@example.com",
				Password:   "password123",
				FromEmail:  "user@example.com",
				Encryption: "STARTTLS",
			},
			wantErr: false,
		},
		{
			name: "Missing name",
			req: CreateSMTPConfigRequest{
				Host:      "smtp.gmail.com",
				Port:      587,
				Username:  "user@example.com",
				Password:  "password123",
				FromEmail: "user@example.com",
			},
			wantErr: true,
			errMsg:  "name is required",
		},
		{
			name: "Invalid port",
			req: CreateSMTPConfigRequest{
				Name:      "Mailer",
				Host:      "smtp.gmail.com",
				Port:      0,
				Username:  "user@example.com",
				Password:  "password123",
				FromEmail: "user@example.com",
			},
			wantErr: true,
			errMsg:  "port must be between 1 and 65535",
		},
		{
			name: "Invalid from email",
			req: CreateSMTPConfigRequest{
				Name:      "Mailer",
				Host:      "smtp.gmail.com",
				Port:      587,
				Username:  "user@example.com",
				Password:  "password123",
				FromEmail: "invalid-email",
			},
			wantErr: true,
			errMsg:  "from_email is not a valid email address",
		},
		{
			name: "Invalid encryption mode",
			req: CreateSMTPConfigRequest{
				Name:       "Mailer",
				Host:       "smtp.gmail.com",
				Port:       587,
				Username:   "user@example.com",
				Password:   "password123",
				FromEmail:  "user@example.com",
				Encryption: "INVALID_MODE",
			},
			wantErr: true,
			errMsg:  "encryption must be STARTTLS, SSL, TLS, or NONE",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.req.Validate()
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestSendEmailRequest_Validate(t *testing.T) {
	tests := []struct {
		name    string
		req     SendEmailRequest
		wantErr bool
		errMsg  string
	}{
		{
			name: "Valid send request",
			req: SendEmailRequest{
				To:      []string{"recipient@example.com"},
				Subject: "Test Subject",
				Body:    "Test Body",
			},
			wantErr: false,
		},
		{
			name: "Missing to recipient",
			req: SendEmailRequest{
				Subject: "Test Subject",
				Body:    "Test Body",
			},
			wantErr: true,
			errMsg:  "at least one 'to' recipient email is required",
		},
		{
			name: "Invalid to recipient email",
			req: SendEmailRequest{
				To:      []string{"not-an-email"},
				Subject: "Test Subject",
				Body:    "Test Body",
			},
			wantErr: true,
			errMsg:  "invalid recipient email",
		},
		{
			name: "Missing subject",
			req: SendEmailRequest{
				To:   []string{"recipient@example.com"},
				Body: "Test Body",
			},
			wantErr: true,
			errMsg:  "subject is required",
		},
		{
			name: "Missing body",
			req: SendEmailRequest{
				To:      []string{"recipient@example.com"},
				Subject: "Test Subject",
			},
			wantErr: true,
			errMsg:  "body is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.req.Validate()
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

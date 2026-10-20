package mailer

import (
	"strings"
	"testing"

	"github.com/arjunstein/backend-smtp/internal/model"
	"github.com/stretchr/testify/assert"
)

func TestBuildMIMEMessage_PlainText(t *testing.T) {
	cfg := &model.SMTPConfig{
		FromEmail: "sender@example.com",
		FromName:  "Mailer Bot",
		Host:      "smtp.example.com",
	}

	req := &model.SendEmailRequest{
		To:      []string{"recipient@example.com"},
		Subject: "Test Email",
		Body:    "Hello World!",
		IsHTML:  false,
	}

	raw, err := BuildMIMEMessage(cfg, req)
	assert.NoError(t, err)

	content := string(raw)
	assert.Contains(t, content, "From: Mailer Bot <sender@example.com>")
	assert.Contains(t, content, "To: recipient@example.com")
	assert.Contains(t, content, "Subject: Test Email")
	assert.Contains(t, content, "Content-Type: text/plain; charset=\"UTF-8\"")
	assert.Contains(t, content, "Hello World!")
}

func TestBuildMIMEMessage_HTMLWithAttachments(t *testing.T) {
	cfg := &model.SMTPConfig{
		FromEmail: "sender@example.com",
		FromName:  "Mailer Bot",
		Host:      "smtp.example.com",
	}

	attachmentData := EncodeToBase64([]byte("dummy pdf file content"))

	req := &model.SendEmailRequest{
		To:      []string{"user@example.com"},
		CC:      []string{"manager@example.com"},
		Subject: "Monthly Report",
		Body:    "<h1>Monthly Report</h1><p>See attached</p>",
		IsHTML:  true,
		Attachments: []model.AttachmentDTO{
			{
				Filename:      "report.pdf",
				ContentType:   "application/pdf",
				ContentBase64: attachmentData,
			},
		},
	}

	raw, err := BuildMIMEMessage(cfg, req)
	assert.NoError(t, err)

	content := string(raw)
	assert.Contains(t, content, "multipart/mixed")
	assert.Contains(t, content, "multipart/alternative")
	assert.Contains(t, content, "Cc: manager@example.com")
	assert.Contains(t, content, "<h1>Monthly Report</h1>")
	assert.Contains(t, content, "filename=\"report.pdf\"")
	assert.Contains(t, content, strings.TrimSpace(attachmentData))
}

package mailer

import (
	"bytes"
	"crypto/rand"
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"mime"
	"net"
	"net/smtp"
	"path/filepath"
	"strings"
	"time"

	"github.com/arjunstein/backend-smtp/internal/model"
)

// Mailer interface defines methods to interact with SMTP servers.
type Mailer interface {
	Send(cfg *model.SMTPConfig, req *model.SendEmailRequest) error
	Ping(cfg *model.SMTPConfig) error
}

type smtpMailer struct{}

// NewMailer creates a new production SMTP mailer.
func NewMailer() Mailer {
	return &smtpMailer{}
}

// Ping tests connection and authentication against the SMTP server.
func (m *smtpMailer) Ping(cfg *model.SMTPConfig) error {
	client, err := createSMTPClient(cfg)
	if err != nil {
		return fmt.Errorf("failed to connect to SMTP server: %w", err)
	}
	defer client.Close()

	if cfg.Username != "" && cfg.Password != "" {
		auth := smtp.PlainAuth("", cfg.Username, cfg.Password, cfg.Host)
		if err := client.Auth(auth); err != nil {
			return fmt.Errorf("SMTP authentication failed: %w", err)
		}
	}

	return client.Noop()
}

// Send composes and transmits an email via configured SMTP transport.
func (m *smtpMailer) Send(cfg *model.SMTPConfig, req *model.SendEmailRequest) error {
	client, err := createSMTPClient(cfg)
	if err != nil {
		return fmt.Errorf("failed to establish SMTP connection: %w", err)
	}
	defer client.Close()

	if cfg.Username != "" && cfg.Password != "" {
		auth := smtp.PlainAuth("", cfg.Username, cfg.Password, cfg.Host)
		if err := client.Auth(auth); err != nil {
			return fmt.Errorf("SMTP authentication failed: %w", err)
		}
	}

	senderEmail := cfg.FromEmail
	if err := client.Mail(senderEmail); err != nil {
		return fmt.Errorf("failed to set sender: %w", err)
	}

	// Collect all recipients (To + CC + BCC)
	allRecipients := make([]string, 0, len(req.To)+len(req.CC)+len(req.BCC))
	for _, to := range req.To {
		trimmed := strings.TrimSpace(to)
		if trimmed != "" {
			allRecipients = append(allRecipients, trimmed)
			if err := client.Rcpt(trimmed); err != nil {
				return fmt.Errorf("failed to add recipient %s: %w", trimmed, err)
			}
		}
	}
	for _, cc := range req.CC {
		trimmed := strings.TrimSpace(cc)
		if trimmed != "" {
			allRecipients = append(allRecipients, trimmed)
			if err := client.Rcpt(trimmed); err != nil {
				return fmt.Errorf("failed to add cc recipient %s: %w", trimmed, err)
			}
		}
	}
	for _, bcc := range req.BCC {
		trimmed := strings.TrimSpace(bcc)
		if trimmed != "" {
			allRecipients = append(allRecipients, trimmed)
			if err := client.Rcpt(trimmed); err != nil {
				return fmt.Errorf("failed to add bcc recipient %s: %w", trimmed, err)
			}
		}
	}

	wc, err := client.Data()
	if err != nil {
		return fmt.Errorf("failed to open data writer: %w", err)
	}
	defer wc.Close()

	msgBytes, err := BuildMIMEMessage(cfg, req)
	if err != nil {
		return fmt.Errorf("failed to build MIME message: %w", err)
	}

	if _, err := wc.Write(msgBytes); err != nil {
		return fmt.Errorf("failed to write email body: %w", err)
	}

	return client.Quit()
}

func createSMTPClient(cfg *model.SMTPConfig) (*smtp.Client, error) {
	addr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
	tlsConfig := &tls.Config{
		ServerName: cfg.Host,
		MinVersion: tls.VersionTLS12,
	}

	enc := strings.ToUpper(cfg.Encryption)
	isSSL := enc == model.EncryptionSSL || enc == model.EncryptionTLS || cfg.Port == 465

	if isSSL {
		// Implicit TLS connection
		dialer := &net.Dialer{Timeout: 10 * time.Second}
		tlsConn, err := tls.DialWithDialer(dialer, "tcp", addr, tlsConfig)
		if err != nil {
			return nil, err
		}
		return smtp.NewClient(tlsConn, cfg.Host)
	}

	// Plain or STARTTLS connection
	conn, err := net.DialTimeout("tcp", addr, 10*time.Second)
	if err != nil {
		return nil, err
	}

	client, err := smtp.NewClient(conn, cfg.Host)
	if err != nil {
		_ = conn.Close()
		return nil, err
	}

	if enc == model.EncryptionSTARTTLS || cfg.Port == 587 {
		if ok, _ := client.Extension("STARTTLS"); ok {
			if err := client.StartTLS(tlsConfig); err != nil {
				_ = client.Close()
				return nil, fmt.Errorf("failed to upgrade connection with STARTTLS: %w", err)
			}
		}
	}

	return client, nil
}

// BuildMIMEMessage constructs RFC 5322 MIME message byte stream.
func BuildMIMEMessage(cfg *model.SMTPConfig, req *model.SendEmailRequest) ([]byte, error) {
	var buf bytes.Buffer

	// Headers
	fromHeader := cfg.FromEmail
	if cfg.FromName != "" {
		fromHeader = fmt.Sprintf("%s <%s>", mime.QEncoding.Encode("utf-8", cfg.FromName), cfg.FromEmail)
	}
	buf.WriteString(fmt.Sprintf("From: %s\r\n", fromHeader))
	buf.WriteString(fmt.Sprintf("To: %s\r\n", strings.Join(req.To, ", ")))
	if len(req.CC) > 0 {
		buf.WriteString(fmt.Sprintf("Cc: %s\r\n", strings.Join(req.CC, ", ")))
	}
	buf.WriteString(fmt.Sprintf("Subject: %s\r\n", mime.QEncoding.Encode("utf-8", req.Subject)))
	buf.WriteString(fmt.Sprintf("Date: %s\r\n", time.Now().Format(time.RFC1123Z)))
	buf.WriteString(fmt.Sprintf("Message-ID: <%s@%s>\r\n", generateRandomID(), cfg.Host))
	buf.WriteString("MIME-Version: 1.0\r\n")

	hasAttachments := len(req.Attachments) > 0
	mixedBoundary := "mixed_" + generateRandomID()
	altBoundary := "alt_" + generateRandomID()

	if hasAttachments {
		buf.WriteString(fmt.Sprintf("Content-Type: multipart/mixed; boundary=\"%s\"\r\n\r\n", mixedBoundary))
		buf.WriteString(fmt.Sprintf("--%s\r\n", mixedBoundary))
	}

	if req.IsHTML {
		buf.WriteString(fmt.Sprintf("Content-Type: multipart/alternative; boundary=\"%s\"\r\n\r\n", altBoundary))

		// Plain text fallback part
		buf.WriteString(fmt.Sprintf("--%s\r\n", altBoundary))
		buf.WriteString("Content-Type: text/plain; charset=\"UTF-8\"\r\n")
		buf.WriteString("Content-Transfer-Encoding: 7bit\r\n\r\n")
		buf.WriteString("This email requires an HTML compatible email viewer.\r\n\r\n")

		// HTML part
		buf.WriteString(fmt.Sprintf("--%s\r\n", altBoundary))
		buf.WriteString("Content-Type: text/html; charset=\"UTF-8\"\r\n")
		buf.WriteString("Content-Transfer-Encoding: 8bit\r\n\r\n")
		buf.WriteString(req.Body)
		buf.WriteString(fmt.Sprintf("\r\n--%s--\r\n", altBoundary))
	} else {
		buf.WriteString("Content-Type: text/plain; charset=\"UTF-8\"\r\n")
		buf.WriteString("Content-Transfer-Encoding: 8bit\r\n\r\n")
		buf.WriteString(req.Body)
		buf.WriteString("\r\n")
	}

	// Append attachments if any
	if hasAttachments {
		for _, att := range req.Attachments {
			buf.WriteString(fmt.Sprintf("\r\n--%s\r\n", mixedBoundary))
			contentType := att.ContentType
			if contentType == "" {
				contentType = "application/octet-stream"
			}
			filename := filepath.Base(att.Filename)
			buf.WriteString(fmt.Sprintf("Content-Type: %s; name=\"%s\"\r\n", contentType, filename))
			buf.WriteString("Content-Transfer-Encoding: base64\r\n")
			buf.WriteString(fmt.Sprintf("Content-Disposition: attachment; filename=\"%s\"\r\n\r\n", filename))

			encoded := formatBase64(att.ContentBase64)
			buf.WriteString(encoded)
			buf.WriteString("\r\n")
		}
		buf.WriteString(fmt.Sprintf("--%s--\r\n", mixedBoundary))
	}

	return buf.Bytes(), nil
}

func generateRandomID() string {
	b := make([]byte, 12)
	_, _ = rand.Read(b)
	return fmt.Sprintf("%x%d", b, time.Now().UnixNano())
}

func formatBase64(raw string) string {
	clean := strings.ReplaceAll(raw, "\n", "")
	clean = strings.ReplaceAll(clean, "\r", "")
	clean = strings.ReplaceAll(clean, " ", "")

	// Break into standard 76 char lines
	const lineLen = 76
	var buf strings.Builder
	for i := 0; i < len(clean); i += lineLen {
		end := i + lineLen
		if end > len(clean) {
			end = len(clean)
		}
		buf.WriteString(clean[i:end])
		buf.WriteString("\r\n")
	}
	return buf.String()
}

// EncodeToBase64 helper for tests or external usage
func EncodeToBase64(data []byte) string {
	return base64.StdEncoding.EncodeToString(data)
}

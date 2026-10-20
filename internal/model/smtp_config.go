package model

import (
	"time"
)

// Encryption types
const (
	EncryptionNone     = "NONE"
	EncryptionSTARTTLS = "STARTTLS"
	EncryptionSSL      = "SSL"
	EncryptionTLS      = "TLS"
)

// SMTPConfig represents SMTP server connection configuration.
type SMTPConfig struct {
	ID         uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	Name       string    `gorm:"size:100;not null;uniqueIndex" json:"name"`
	Host       string    `gorm:"size:255;not null" json:"host"`
	Port       int       `gorm:"not null" json:"port"`
	Username   string    `gorm:"size:255;not null" json:"username"`
	Password   string    `gorm:"size:255;not null" json:"-"`
	FromEmail  string    `gorm:"size:255;not null" json:"from_email"`
	FromName   string    `gorm:"size:255" json:"from_name"`
	Encryption string    `gorm:"size:20;default:'STARTTLS'" json:"encryption"`
	IsActive   bool      `gorm:"default:true;index" json:"is_active"`
	IsDefault  bool      `gorm:"default:false;index" json:"is_default"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// TableName overrides the table name for GORM.
func (SMTPConfig) TableName() string {
	return "smtp_configs"
}

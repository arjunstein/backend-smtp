package model

import (
	"time"
)

// Email delivery status constants
const (
	EmailStatusSuccess = "SUCCESS"
	EmailStatusFailed  = "FAILED"
)

// EmailLog records every email sending attempt for auditing and observability.
type EmailLog struct {
	ID           uint        `gorm:"primaryKey;autoIncrement" json:"id"`
	SMTPConfigID *uint       `gorm:"index" json:"smtp_config_id,omitempty"`
	SMTPConfig   *SMTPConfig `gorm:"foreignKey:SMTPConfigID;constraint:OnDelete:SET NULL" json:"smtp_config,omitempty"`
	To           string      `gorm:"type:text;not null" json:"to"`
	CC           string      `gorm:"type:text" json:"cc,omitempty"`
	BCC          string      `gorm:"type:text" json:"bcc,omitempty"`
	Subject      string      `gorm:"size:500;not null" json:"subject"`
	Status       string      `gorm:"size:50;not null;index" json:"status"`
	ErrorMessage string      `gorm:"type:text" json:"error_message,omitempty"`
	SentAt       time.Time   `gorm:"index" json:"sent_at"`
	CreatedAt    time.Time   `json:"created_at"`
}

// TableName overrides the table name for GORM.
func (EmailLog) TableName() string {
	return "email_logs"
}

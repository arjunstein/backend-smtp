package model

import (
	"time"

	"gorm.io/gorm"
)

// User represents an authorized user in the system.
type User struct {
	ID        uint           `gorm:"primaryKey;autoIncrement" json:"id"`
	Name      string         `gorm:"size:100;not null" json:"name"`
	Username  string         `gorm:"size:50;not null;uniqueIndex" json:"username"`
	Email     string         `gorm:"size:150;not null;uniqueIndex" json:"email"`
	Password  string         `gorm:"size:255;not null" json:"-"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName specifies table name for GORM.
func (User) TableName() string {
	return "users"
}

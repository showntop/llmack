package models

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	Username    string    `gorm:"uniqueIndex;not null" json:"username"`
	Password    string    `gorm:"not null" json:"-"`
	Email       string    `gorm:"uniqueIndex;not null" json:"email"`
	LastLoginAt time.Time `json:"last_login_at"`
	IsActive    bool      `gorm:"default:true" json:"is_active"`
}

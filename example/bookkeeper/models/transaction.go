package models

import (
	"time"

	"gorm.io/gorm"
)

type Transaction struct {
	gorm.Model
	Amount      float64   `json:"amount"`
	Category    string    `json:"category"`
	Description string    `json:"description"`
	Date        time.Time `json:"date"`
	ImageURL    string    `json:"image_url"`
	UserID      uint      `json:"user_id"`
}

type Category struct {
	gorm.Model
	Name        string `json:"name"`
	Description string `json:"description"`
	Type        string `json:"type"` // income or expense
}

type Budget struct {
	gorm.Model
	Category  string    `json:"category"`
	Amount    float64   `json:"amount"`
	StartDate time.Time `json:"start_date"`
	EndDate   time.Time `json:"end_date"`
	UserID    uint      `json:"user_id"`
}

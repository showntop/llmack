package storage

import (
	"time"

	"github.com/showntop/llmack/memory"
)

// for domain usage
type Session struct {
	ID         string         `json:"id" gorm:"column:id;type:bigint(20);AUTO_INCREMENT;primary_key;"` // session id
	EngineID   string         `json:"engine_id" gorm:"column:engine_id;type:varchar(255);"`            // engine id
	EngineType string         `json:"engine_type" gorm:"column:engine_type;type:varchar(255);"`        // engine type
	EngineData map[string]any `json:"engine_data" gorm:"column:engine_data;type:jsonb"`                // engine data

	Memory memory.Memory `json:"memory"` // 记忆

	CreatedAt time.Time `json:"created_at" gorm:"column:created_at;autoCreateTime"`
	UpdatedAt time.Time `json:"updated_at" gorm:"column:updated_at;autoUpdateTime"`
}

func NewSession(id string) *Session {
	return &Session{
		ID: id,
	}
}

// For DB row
type SessionRow struct {
	ID         string         `gorm:"column:id;type:bigint(20);AUTO_INCREMENT;primary_key;"` // session id
	EngineID   string         `gorm:"column:engine_id;type:varchar(255);"`                   // engine id
	EngineType string         `gorm:"column:engine_type;type:varchar(255);"`                 // engine type
	EngineData map[string]any `gorm:"column:engine_data;type:jsonb"`                         // engine data

	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt time.Time `gorm:"column:updated_at;autoUpdateTime"`
}

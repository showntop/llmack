package storage

import (
	"time"

	"github.com/showntop/llmack/llm"
	"github.com/showntop/llmack/memory"
)

// for domain usage
type Session struct {
	UID        string         `json:"uid" gorm:"column:uid;type:varchar(255);"`                 // session id
	EngineID   string         `json:"engine_id" gorm:"column:engine_id;type:varchar(255);"`     // engine id
	EngineType string         `json:"engine_type" gorm:"column:engine_type;type:varchar(255);"` // engine type
	EngineData map[string]any `json:"engine_data" gorm:"column:engine_data;type:jsonb"`         // engine data

	Memory memory.Memory `json:"memory"` // 记忆

	Messages []llm.Message `json:"messages"`

	Data any `json:"data"`

	CreatedAt time.Time `json:"created_at" gorm:"column:created_at;autoCreateTime"`
	UpdatedAt time.Time `json:"updated_at" gorm:"column:updated_at;autoUpdateTime"`
}

func NewSession(id string) *Session {
	return &Session{
		UID: id,
	}
}

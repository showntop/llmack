package memory

import "time"

type MemoryItem struct {
	ID        int64
	SessionID string
	Content   string
	Topics    []string
	Extra     *Extra
	CreatedAt time.Time
	UpdatedAt time.Time
}

func NewMemoryItem(sessionID string, content string, extra *Extra) *MemoryItem {
	return &MemoryItem{
		SessionID: sessionID,
		Content:   content,
		Extra:     extra,
	}
}

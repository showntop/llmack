package memory

import (
	"context"
)

// Extra ...
type Extra any

type Memory interface {
	Add(context.Context, string, *MemoryItem) error
	FetchHistories(context.Context, string) ([]*MemoryItem, error)
}

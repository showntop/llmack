package vdb

import (
	"context"
	"fmt"
	"time"

	"github.com/pgvector/pgvector-go"
)

var constructors = map[string]func(config any) (VDB, error){}

// NewVDB ...
func NewVDB(name string, config any) (VDB, error) {
	constructor, ok := constructors[name]
	if !ok {
		return nil, fmt.Errorf("vdb %s not found", name)
	}
	vdb, err := constructor(config)
	if err != nil {
		return nil, fmt.Errorf("new vdb %s failed: %w", name, err)
	}
	return vdb, nil
}

// Register ...
func Register(name string, constructor func(config any) (VDB, error)) {
	constructors[name] = constructor
}

// VDB ...
type VDB interface {
	Create(context.Context) error
	Store(context.Context, ...*Document) error
	// BatchStore(context.Context, []string, [][]float64) error

	SearchWithOptions(context.Context, []float32, *SearchOptions) ([]*Document, error)
	Search(context.Context, []float32, ...SearchOption) ([]*Document, error)
	SearchQuery(context.Context, string, ...SearchOption) ([]*Document, error)
	SearchQueryWithOptions(context.Context, string, *SearchOptions) ([]*Document, error)

	Delete(context.Context, string) error
	Close() error
}

// Document 文档
type Document struct {
	ID          string          `json:"id"`
	Title       string          `json:"title"`
	Content     string          `json:"content"`
	ContentHash string          `json:"content_hash"`
	Embedding   pgvector.Vector `json:"embedding"`
	Similarity  float32         `json:"similarity"`
	Metadata    map[string]any  `json:"metadata"`
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
}

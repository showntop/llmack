package vdb

import "context"

// VDB ...
type VDB interface {
	Store(context.Context, string, []float64) error
	// BatchStore(context.Context, []string, [][]float64) error

	Search(context.Context, []float64, ...SearchOption) ([]Document, error)
	SearchQuery(context.Context, string, ...SearchOption) ([]Document, error)

	Delete(context.Context, string) error
	Close() error
}

// Document 文档
type Document struct {
	ID     string    `json:"id"`
	Query  string    `json:"query"`
	Answer string    `json:"answer"`
	Score  []float64 `json:"score"`
	Scores []float32 `json:"scores"`
	Vector []float64 `json:"vector"`
}

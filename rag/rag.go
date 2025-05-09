package rag

import (
	"context"

	"github.com/showntop/llmack/vdb"
)

// ScalarDB ...
type ScalarDB interface {
	Fetch(context.Context, string, *SearchOptions) (*vdb.Document, error)
}

// Options ...
type SearchOptions struct {
	LibraryID      int64   `json:"id"`
	Kind           string  `json:"kind"`
	IndexID        int32   `json:"index_id"`
	TopK           int     `json:"top_k"`
	ScoreThreshold float64 `json:"score_threshold"`
}

type SearchOption func(*SearchOptions)

func WithTopK(topK int) SearchOption {
	return func(o *SearchOptions) {
		o.TopK = topK
	}
}

func WithScoreThreshold(scoreThreshold float64) SearchOption {
	return func(o *SearchOptions) {
		o.ScoreThreshold = scoreThreshold
	}
}

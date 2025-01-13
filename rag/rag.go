package rag

import "context"

// KnowledgeEntity TODO
type KnowledgeEntity struct {
	ID        int64   `json:"id"`
	LibraryID int64   `json:"lib_id,omitempty"`
	Question  string  `json:"question"`
	Answer    string  `json:"answer"`
	Score     float64 `json:"score"`
}

// ScalarDB ...
type ScalarDB interface {
	Fetch(context.Context, string, *Options) (*KnowledgeEntity, error)
}

// Options ...
type Options struct {
	LibraryID      int64   `json:"id"`
	Kind           string  `json:"kind"`
	IndexID        int32   `json:"index_id"`
	TopK           int32   `json:"top_k"`
	ScoreThreshold float64 `json:"score_threshold"`
}

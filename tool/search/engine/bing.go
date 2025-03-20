package engine

import (
	"context"
)

// NewBing bing search
func NewBing(apiKey string) Searcher {
	return &Serper{apiKey: apiKey}
}

// Bing search by big
type Bing struct {
	apiKey string
}

// Search 使用serper搜索
func (s *Bing) Search(ctx context.Context, query string) ([]*Result, error) {
	return nil, nil
}

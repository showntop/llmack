package engine

import "context"

// Searcher 搜索器
type Searcher interface {
	Search(ctx context.Context, query string) ([]*Result, error)
}

// Result 搜索结果
type Result struct {
	Time    string `json:"time"`
	Link    string `json:"link"`
	Image   string `json:"image"`
	Video   string `json:"id"` // video id
	Title   string `json:"title"`
	Snippet string `json:"snippet"`
	Content string `json:"content"`
}

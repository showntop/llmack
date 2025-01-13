package llm

import (
	"context"
	"errors"
	"time"

	"github.com/showntop/llmack/embedding"
	"github.com/showntop/llmack/vdb"
)

// CachedDocument ...
type CachedDocument struct {
	ID     string
	Query  string
	Answer string
	Score  float64
	Vector []float64
}

// CacheFactory ...
type CacheFactory func() Cache

// Cache ...
type Cache interface {
	Fetch(context.Context, []Message) (*CachedDocument, bool, error)
	Store(context.Context, *CachedDocument, string) error
}

// BaseCache ...
type BaseCache struct {
	Config      *CacheConfig
	Embedder    embedding.Embedder
	VectorStore vdb.VDB
	Ranker      Scorer
}

// SetQueryProcessor ...
func (c *BaseCache) SetQueryProcessor(p QueryProcessor) {
	c.Config.QueryProcessor = p
}

// CacheConfig 定义缓存的配置选项
type CacheConfig struct {
	TTL             time.Duration
	MaxEntries      int
	CleanupInterval time.Duration
	EnableMetrics   bool
	QueryProcessor  QueryProcessor
}

// DefaultConfig 返回默认配置
func DefaultConfig() *CacheConfig {
	return &CacheConfig{
		TTL:             5 * time.Minute,
		MaxEntries:      1000,
		CleanupInterval: 1 * time.Minute,
		EnableMetrics:   true,
		QueryProcessor:  LastQueryMessage,
	}
}

// 添加自定义错误类型
var (
	ErrCacheMiss = errors.New("cache miss")
	ErrCacheFull = errors.New("cache is full")
)

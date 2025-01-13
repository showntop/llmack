package llm

import (
	"context"

	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"github.com/showntop/llmack/embedding"
	"github.com/showntop/llmack/vdb"
	"github.com/showntop/llmack/vdb/memo"
)

// RedisCache ...
type RedisCache struct {
	BaseCache
	cli *redis.Client
}

// NewRedisCache ...
func NewRedisCache(cli *redis.Client) Cache {
	return &RedisCache{
		BaseCache: BaseCache{
			Config:      DefaultConfig(),
			Embedder:    embedding.NewStringEmbedder(),
			VectorStore: memo.New(),
		},
		cli: cli,
	}
}

// Fetch ...
func (m *RedisCache) Fetch(ctx context.Context, messages []Message) (*CachedDocument, bool, error) {
	// 处理query
	query, err := m.Config.QueryProcessor(messages)
	if err != nil {
		return nil, false, err
	}
	// embedding vector
	vector, err := m.Embedder.Embed(ctx, query)
	if err != nil {
		return nil, false, err
	}
	// search vector ids
	relations, err := m.VectorStore.Search(ctx, vector, vdb.WithTopk(5))
	if err != nil {
		return nil, false, err
	}
	if relations == nil || len(relations) == 0 {
		return &CachedDocument{Query: query, Vector: vector}, false, nil
	}
	ids := make([]string, 0, len(relations))
	for i := 0; i < len(relations); i++ {
		ids[i] = relations[i].ID
	}
	// batch get
	res, err := m.cli.MGet(ctx, ids...).Result()
	if err != nil {
		return nil, false, err
	}
	_ = res
	// object data
	documents := make([]*CachedDocument, 0, len(relations))

	for i := 0; i < len(relations); i++ {

	}
	if len(documents) == 0 {
		return nil, false, nil
	}

	// 排序
	document := documents[0]
	return document, true, nil
}

// Store ...
func (m *RedisCache) Store(ctx context.Context, document *CachedDocument, value string) error {
	document.ID = uuid.NewString()
	if err := m.VectorStore.Store(ctx, document.ID, document.Vector); err != nil {
		return err
	}
	return nil
}

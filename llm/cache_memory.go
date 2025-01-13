package llm

import (
	"context"
	"sort"

	"github.com/google/uuid"
	"github.com/showntop/llmack/embedding"
	"github.com/showntop/llmack/vdb"
	"github.com/showntop/llmack/vdb/memo"
)

// MemoCache ...
type MemoCache struct {
	BaseCache

	documentTable map[string]*CachedDocument
}

// NewMemoCache ...
func NewMemoCache() Cache {
	return &MemoCache{
		BaseCache: BaseCache{
			Config:      DefaultConfig(),
			Embedder:    embedding.NewStringEmbedder(),
			VectorStore: memo.New(),
		},
		documentTable: make(map[string]*CachedDocument),
	}
}

// Fetch ...
func (m *MemoCache) Fetch(ctx context.Context, messages []Message) (*CachedDocument, bool, error) {
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
	// object data
	documents := make([]*CachedDocument, 0, len(relations))
	for _, dc := range relations {
		document := m.documentTable[dc.ID]
		document.Query = query
		document.Vector = vector
		documents = append(documents, document)
	}
	if len(documents) == 0 {
		return nil, false, nil
	}

	// 1. score
	sort.Slice(documents, func(i, j int) bool {
		return documents[i].Score > documents[j].Score
	})
	// 2. rank Temperature Softmax / sort

	document := documents[0]

	// threshold
	const similarityThreshold = 0.8
	if document.Score < similarityThreshold {
		return &CachedDocument{Query: query, Vector: vector}, false, nil
	}

	return document, true, nil
}

// Store ...
func (m *MemoCache) Store(ctx context.Context, document *CachedDocument, value string) error {
	// 检查容量限制
	if len(m.documentTable) >= m.Config.MaxEntries {
		// 可以实现 LRU 或其他淘汰策略
	}

	document.ID = uuid.NewString()
	document.Answer = value
	// document.CreatedAt = time.Now()
	// document.UpdatedAt = time.Now()
	if err := m.VectorStore.Store(ctx, document.ID, document.Vector); err != nil {
		return err
	}
	m.documentTable[document.ID] = document
	return nil
}

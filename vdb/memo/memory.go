package memo

import (
	"context"
	"fmt"
	"math"
	"sort"
	"sync"
	"time"

	"github.com/showntop/llmack/embedding"
	"github.com/showntop/llmack/vdb"
)

// VDB 实现基于内存的向量存储
type VDB struct {
	embedder embedding.Embedder
	vectors  map[string]document
	mutex    sync.RWMutex
}

type document struct {
	ID        string
	Vector    []float32
	Title     string
	Content   string
	Metadata  map[string]any
	CreatedAt time.Time
	UpdatedAt time.Time
}

// New 创建新的内存向量存储实例
func New() vdb.VDB {
	return &VDB{
		vectors:  make(map[string]document),
		embedder: embedding.NewStringEmbedder(),
	}
}

func (m *VDB) Create(_ context.Context) error {
	return nil
}

// Store 存储向量
func (m *VDB) Store(ctx context.Context, docs ...*vdb.Document) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// 创建副本以避免外部修改
	for _, doc := range docs {
		embedding, err := m.embedder.Embed(ctx, doc.Content)
		if err != nil {
			return err
		}
		m.vectors[doc.ID] = document{
			ID:        doc.ID,
			Vector:    embedding,
			Title:     doc.Title,
			Content:   doc.Content,
			Metadata:  doc.Metadata,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
	}
	return nil
}

// SearchQuery 搜索最相似的向量
func (m *VDB) SearchQuery(ctx context.Context, query string, options ...vdb.SearchOption) ([]*vdb.Document, error) {
	searchOptions := &vdb.SearchOptions{
		TopK:      10,
		Threshold: 0.5,
	}
	for _, option := range options {
		option(searchOptions)
	}
	return m.SearchQueryWithOptions(ctx, query, searchOptions)
}

// SearchQueryWithOptions 搜索最相似的向量
func (m *VDB) SearchQueryWithOptions(ctx context.Context, query string, options *vdb.SearchOptions) ([]*vdb.Document, error) {
	embedding, err := m.embedder.Embed(ctx, query)
	if err != nil {
		return nil, err
	}
	return m.SearchWithOptions(ctx, embedding, options)
}

func (m *VDB) SearchWithOptions(ctx context.Context, vector []float32, options *vdb.SearchOptions) ([]*vdb.Document, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	// 计算所有向量的相似度
	documents := make([]*vdb.Document, 0, len(m.vectors))
	for _, v := range m.vectors {
		if len(v.Vector) != len(vector) {
			return nil, fmt.Errorf("vector dimension mismatch")
		}

		similarity := cosineSimilarity(vector, v.Vector)
		documents = append(documents, &vdb.Document{
			ID:         v.ID,
			Title:      v.Title,
			Content:    v.Content,
			Metadata:   v.Metadata,
			CreatedAt:  v.CreatedAt,
			UpdatedAt:  v.UpdatedAt,
			Similarity: similarity,
		})
	}

	// 按相似度排序
	sort.Slice(documents, func(i, j int) bool {
		return documents[i].Similarity > documents[j].Similarity
	})

	k := min(options.TopK, len(documents))
	return documents[:k], nil
}

// Search 搜索最相似的向量
func (m *VDB) Search(ctx context.Context, vector []float32, options ...vdb.SearchOption) ([]*vdb.Document, error) {
	// 获取搜索选项
	searchOptions := &vdb.SearchOptions{
		TopK:      10,
		Threshold: 0.5,
	}
	for _, option := range options {
		option(searchOptions)
	}
	return m.SearchWithOptions(ctx, vector, searchOptions)
}

// Delete 删除向量
func (m *VDB) Delete(_ context.Context, id string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	delete(m.vectors, id)
	return nil
}

// Close 关闭并清理资源
func (m *VDB) Close() error {
	return nil
}

// 计算余弦相似度
func cosineSimilarity(a, b []float32) float32 {
	var dotProduct, normA, normB float32
	for i := range a {
		dotProduct += a[i] * b[i]
		normA += a[i] * a[i]
		normB += b[i] * b[i]
	}

	normA = float32(math.Sqrt(float64(normA)))
	normB = float32(math.Sqrt(float64(normB)))

	if normA == 0 || normB == 0 {
		return 0
	}

	return dotProduct / (normA * normB)
}

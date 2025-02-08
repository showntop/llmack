package memo

import (
	"context"
	"fmt"
	"math"
	"sort"
	"sync"

	"github.com/showntop/llmack/vdb"
)

// VDB 实现基于内存的向量存储
type VDB struct {
	vectors map[string][]float64
	mutex   sync.RWMutex
}

// New 创建新的内存向量存储实例
func New() vdb.VDB {
	return &VDB{
		vectors: make(map[string][]float64),
	}
}

// Store 存储向量
func (m *VDB) Store(_ context.Context, id string, vector []float64) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// 创建副本以避免外部修改
	vectorCopy := make([]float64, len(vector))
	copy(vectorCopy, vector)
	m.vectors[id] = vectorCopy
	return nil
}

// SearchQuery 搜索最相似的向量
func (m *VDB) SearchQuery(_ context.Context, query string, options ...vdb.SearchOption) ([]vdb.Document, error) {
	return nil, nil
}

// Search 搜索最相似的向量
func (m *VDB) Search(_ context.Context, vector []float64, options ...vdb.SearchOption) ([]vdb.Document, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	// 获取搜索选项
	searchOptions := &vdb.SearchOptions{
		Topk:      10,
		Threshold: 0.5,
	}
	for _, option := range options {
		option(searchOptions)
	}

	// 计算所有向量的相似度
	scores := make([]vdb.Document, 0, len(m.vectors))
	for id, v := range m.vectors {
		if len(v) != len(vector) {
			return nil, fmt.Errorf("vector dimension mismatch")
		}

		score := cosineSimilarity(vector, v)
		scores = append(scores, vdb.Document{
			ID:     id,
			Score:  []float64{score},
			Vector: v,
		})
	}

	// 按相似度排序
	sort.Slice(scores, func(i, j int) bool {
		return scores[i].Score[0] > scores[j].Score[0]
	})

	k := searchOptions.Topk
	// 返回前k个结果
	if k > len(scores) {
		k = len(scores)
	}
	return scores[:k], nil
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
func cosineSimilarity(a, b []float64) float64 {
	var dotProduct, normA, normB float64
	for i := range a {
		dotProduct += a[i] * b[i]
		normA += a[i] * a[i]
		normB += b[i] * b[i]
	}

	normA = math.Sqrt(normA)
	normB = math.Sqrt(normB)

	if normA == 0 || normB == 0 {
		return 0
	}

	return dotProduct / (normA * normB)
}

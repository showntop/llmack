package milvus

import (
	"context"
	"fmt"

	"github.com/milvus-io/milvus-sdk-go/v2/client"
	"github.com/milvus-io/milvus-sdk-go/v2/entity"

	"github.com/showntop/llmack/vdb"
)

// VDB ...
type VDB struct {
	client     client.Client
	collection string
	dimension  int
}

type Config struct {
	Address    string
	Collection string
	Dimension  int
}

// New 创建新的Milvus向量存储实例
func New(cfg Config) (*VDB, error) {
	c, err := client.NewGrpcClient(context.Background(), cfg.Address)
	if err != nil {
		return nil, fmt.Errorf("connect milvus: %w", err)
	}

	db := &VDB{
		client:     c,
		collection: cfg.Collection,
		dimension:  cfg.Dimension,
	}

	// 确保集合存在
	if err := db.ensureCollection(); err != nil {
		return nil, err
	}

	return db, nil
}

// Search ...
// @param ctx 上下文
// @param vector 查询向量
// @param opts 搜索选项
func (m *VDB) Search(ctx context.Context, vector []float64, opts ...vdb.SearchOption) ([]vdb.Document, error) {
	// 应用搜索选项
	options := &vdb.SearchOptions{
		Topk:      10,
		Threshold: 0.5,
	}
	for _, opt := range opts {
		opt(options)
	}

	// 准备搜索参数
	sp, err := entity.NewIndexFlatSearchParam()
	if err != nil {
		return nil, fmt.Errorf("create search param: %w", err)
	}

	searchResult, err := m.client.Search(
		ctx,
		m.collection,
		[]string{},              // partition names
		"vector_field",          // vector field name
		[]entity.Vector(vector), // search vectors
		"id",                    // output fields
		entity.L2,               // metric type
		options.Topk,            // topK
		sp,                      // search param
	)
	if err != nil {
		return nil, fmt.Errorf("search: %w", err)
	}

	// 转换结果
	var docs []vdb.Document
	for i := 0; i < len(searchResult); i++ {
		score := searchResult[i].Scores
		if score > options.Threshold {
			docs = append(docs, vdb.Document{
				ID:     searchResult[i].ID,
				Score:  score,
				Vector: vector,
			})
		}
	}

	return docs, nil
}

// SearchQuery ...
func (m *VDB) SearchQuery(ctx context.Context, query string, opts ...vdb.SearchOption) ([]vdb.Document, error) {
	// 这里需要先将 query 转换为向量
	// 假设有一个 textToVector 函数可以完成这个转换
	vector, err := m.textToVector(query)
	if err != nil {
		return nil, fmt.Errorf("convert query to vector: %w", err)
	}

	return m.Search(ctx, vector, opts...)
}

// Delete ...
func (m *VDB) Delete(ctx context.Context, id string) error {
	expr := fmt.Sprintf("id == %s", id)
	err := m.client.Delete(
		ctx,
		m.collection,
		"",
		expr,
	)
	if err != nil {
		return fmt.Errorf("delete: %w", err)
	}
	return nil
}

// Close ...
func (m *VDB) Close() error {
	return m.client.Close()
}

// 私有辅助方法
func (m *VDB) ensureCollection() error {
	ctx := context.Background()

	has, err := m.client.HasCollection(ctx, m.collection)
	if err != nil {
		return fmt.Errorf("check collection: %w", err)
	}

	if !has {
		schema := &entity.Schema{
			CollectionName: m.collection,
			Fields: []*entity.Field{
				{
					Name:       "id",
					DataType:   entity.FieldTypeInt64,
					PrimaryKey: true,
					AutoID:     true,
				},
				{
					Name:     "vector_field",
					DataType: entity.FieldTypeFloatVector,
					// Dimension: m.dimension,
				},
			},
		}

		err = m.client.CreateCollection(ctx, schema, 1)
		if err != nil {
			return fmt.Errorf("create collection: %w", err)
		}
	}

	return nil
}

func (m *VDB) textToVector(text string) ([]float64, error) {
	// TODO: 实现文本到向量的转换
	// 这里需要调用外部的嵌入模型服务
	return nil, fmt.Errorf("not implemented")
}

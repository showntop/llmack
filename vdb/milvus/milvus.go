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

type FieldConfig struct {
	Name      string
	Type      entity.FieldType
	IsPrimary bool
	IsAuto    bool
}

type Config struct {
	Address     string
	Collection  string
	Description string
	Dimension   int
	ShareNum    int
	Fields      []FieldConfig
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
	return db, nil
}

// Search ...
// @param ctx 上下文
// @param vector 查询向量
// @param opts 搜索选项
func (m *VDB) Search(ctx context.Context, vector []float64, opts ...vdb.SearchOption) ([]vdb.Document, error) {
	// 应用搜索选项
	options := &vdb.SearchOptions{
		TopK:      10,
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
	sp.AddRadius(options.Threshold)
	searchResult, err := m.client.Search(
		ctx,
		m.collection,
		[]string{},               // partition names
		"",                       // expr,Filter expressions
		[]string{"vector_field"}, // List ofx field names to include in the return.
		//[]entity.Vector{entity.FloatVector(vector)}, // search vectors
		[]entity.Vector{}, // search vectors
		"id",              // vector fields
		entity.L2,         // metric type
		options.TopK,      // topK
		sp,                // search param
	)
	if err != nil {
		return nil, fmt.Errorf("search: %w", err)
	}

	// 转换结果
	var docs []vdb.Document
	for i := 0; i < len(searchResult); i++ {
		score := searchResult[i].Scores
		_ = score
		// if score > options.Threshold {
		// 	docs = append(docs, vdb.Document{
		// 		ID:     searchResult[i].ID,
		// 		Score:  score,
		// 		Vector: vector,
		// 	})
		// }
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
func (m *VDB) EnsureCollection(ctx context.Context, cfg Config) error {
	has, err := m.client.HasCollection(ctx, cfg.Collection)
	if err != nil {
		return fmt.Errorf("check collection: %w", err)
	}
	if has {
		return nil
	}
	schema := entity.NewSchema().WithName(cfg.Collection).WithDescription(cfg.Description)
	for _, f := range cfg.Fields {
		schema = schema.WithField(entity.NewField().WithName(f.Name).WithDataType(f.Type).WithIsPrimaryKey(f.IsPrimary).WithIsAutoID(f.IsAuto).WithDim(int64(m.dimension)))
	}

	err = m.client.CreateCollection(ctx, schema, int32(cfg.ShareNum))
	if err != nil {
		return fmt.Errorf("create collection: %w", err)
	}
	return nil
}

func (m *VDB) textToVector(text string) ([]float64, error) {
	// TODO: 实现文本到向量的转换
	// 这里需要调用外部的嵌入模型服务
	return nil, fmt.Errorf("not implemented")
}

package milvus

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/showntop/llmack/embedding"
	"github.com/showntop/llmack/vdb"

	"github.com/milvus-io/milvus-sdk-go/v2/client"
	"github.com/milvus-io/milvus-sdk-go/v2/entity"
)

// 注册构造函数
func init() {
	vdb.Register("milvus", func(config any) (vdb.VDB, error) {
		cfg, ok := config.(*Config)
		if !ok {
			return nil, fmt.Errorf("invalid config type")
		}
		return New(cfg)
	})
}

// MilvusVDB 封装 Milvus 客户端
type MilvusVDB struct {
	client client.Client
	config *Config
}

// Config Milvus 配置
type Config struct {
	Address        string `json:"address"`
	Client         client.Client
	CollectionName string `json:"collection_name"`
	Dim            int    `json:"dim"`
	Embedder       embedding.Embedder
	Distance       vdb.Distance
}

// New 创建新的 Milvus
func New(config *Config) (*MilvusVDB, error) {
	ctx := context.Background()

	var cc client.Client
	if config.Client != nil {
		cc = config.Client
	} else {
		var err error
		cc, err = client.NewGrpcClient(ctx, config.Address)
		if err != nil {
			return nil, fmt.Errorf("failed to connect to Milvus: %v", err)
		}
	}
	return &MilvusVDB{
		client: cc,
		config: config,
	}, nil
}

// Create 创建集合
func (m *MilvusVDB) Create(ctx context.Context) error {
	if m.config.CollectionName == "" {
		return fmt.Errorf("collection name is required")
	}
	has, err := m.client.HasCollection(ctx, m.config.CollectionName)
	if err != nil {
		return err
	}
	if has {
		return nil
	}

	schema := &entity.Schema{
		CollectionName: m.config.CollectionName,
		Description:    "Collection for document search",
		Fields: []*entity.Field{
			{
				Name:       "id",
				DataType:   entity.FieldTypeVarChar,
				PrimaryKey: true,
				AutoID:     false,
				TypeParams: map[string]string{
					"max_length": "65535",
				},
			},
			{
				Name:     "title",
				DataType: entity.FieldTypeVarChar,
				TypeParams: map[string]string{
					"max_length": "65535",
				},
			},
			{
				Name:     "content",
				DataType: entity.FieldTypeVarChar,
				TypeParams: map[string]string{
					"max_length": "65535",
				},
			},
			{
				Name:     "content_hash",
				DataType: entity.FieldTypeVarChar,
				TypeParams: map[string]string{
					"max_length": "512",
				},
			},
			{
				Name:     "vector",
				DataType: entity.FieldTypeFloatVector,
				TypeParams: map[string]string{
					"dim": fmt.Sprintf("%d", m.config.Dim),
				},
			},
			{
				Name:     "metadata",
				DataType: entity.FieldTypeJSON,
			},
			{
				Name:     "created_at",
				DataType: entity.FieldTypeInt64,
			},
			{
				Name:     "updated_at",
				DataType: entity.FieldTypeInt64,
			},
		},
	}

	if err := m.client.CreateCollection(ctx, schema, 2,
		client.WithMetricsType(metricsType(m.config.Distance)),
	); err != nil {
		return err
	}

	idx, _ := entity.NewIndexIvfFlat(metricsType(m.config.Distance), 128)
	if err := m.client.CreateIndex(ctx, m.config.CollectionName, "vector",
		idx,
		true,
	); err != nil {
		return err
	}
	return nil
}

// Store 存储文档
func (m *MilvusVDB) Store(ctx context.Context, docs ...*vdb.Document) error {
	if len(docs) == 0 {
		return nil
	}

	// 准备数据
	ids := make([]string, len(docs))
	titles := make([]string, len(docs))
	contents := make([]string, len(docs))
	hashes := make([]string, len(docs))
	vectors := make([][]float32, len(docs))
	metadatas := make([][]byte, len(docs))
	createdAts := make([]int64, len(docs))
	updatedAts := make([]int64, len(docs))

	for i, doc := range docs {
		ids[i] = doc.ID
		titles[i] = doc.Title
		contents[i] = doc.Content
		hashes[i] = doc.ContentHash
		vectors[i], _ = m.config.Embedder.Embed(ctx, doc.Content)
		metadataBytes, _ := json.Marshal(doc.Metadata)
		metadatas[i] = (metadataBytes)
		createdAts[i] = doc.CreatedAt.Unix()
		updatedAts[i] = doc.UpdatedAt.Unix()
	}

	// 创建列
	idColumn := entity.NewColumnVarChar("id", ids)
	titleColumn := entity.NewColumnVarChar("title", titles)
	contentColumn := entity.NewColumnVarChar("content", contents)
	hashColumn := entity.NewColumnVarChar("content_hash", hashes)
	vectorColumn := entity.NewColumnFloatVector("vector", m.config.Dim, vectors)
	metadataColumn := entity.NewColumnJSONBytes("metadata", metadatas)
	createdAtColumn := entity.NewColumnInt64("created_at", createdAts)
	updatedAtColumn := entity.NewColumnInt64("updated_at", updatedAts)

	// 插入数据
	_, err := m.client.Upsert(ctx, m.config.CollectionName, "",
		idColumn, titleColumn, contentColumn, hashColumn,
		vectorColumn, metadataColumn, createdAtColumn, updatedAtColumn)
	return err
}

// SearchWithOptions 使用选项搜索
func (m *MilvusVDB) SearchWithOptions(ctx context.Context, vector []float32, opts *vdb.SearchOptions) ([]*vdb.Document, error) {
	// 创建搜索参数
	sp, _ := entity.NewIndexFlatSearchParam()

	// 执行搜索
	results, err := m.client.Search(
		ctx,
		m.config.CollectionName,
		[]string{},
		"",
		[]string{"id", "title", "content", "content_hash", "metadata", "created_at", "updated_at"},
		// []string{"*"},
		[]entity.Vector{entity.FloatVector(vector)},
		"vector",
		metricsType(m.config.Distance),
		opts.TopK,
		sp,
	)
	if err != nil {
		return nil, err
	}

	// 处理结果
	var documents []*vdb.Document
	for _, result := range results {
		for i, score := range result.Scores {
			doc := &vdb.Document{
				Similarity: score,
			}
			for _, field := range result.Fields {
				// 获取字段值
				if field.Name() == "id" {
					idColumn := field.(*entity.ColumnVarChar)
					doc.ID = idColumn.Data()[i]
				} else if field.Name() == "title" {
					titleColumn := field.(*entity.ColumnVarChar)
					doc.Title = titleColumn.Data()[i]
				} else if field.Name() == "content" {
					contentColumn := field.(*entity.ColumnVarChar)
					doc.Content = contentColumn.Data()[i]
				} else if field.Name() == "content_hash" {
					hashColumn := field.(*entity.ColumnVarChar)
					doc.ContentHash = hashColumn.Data()[i]
				} else if field.Name() == "metadata" {
					metadataColumn := field.(*entity.ColumnJSONBytes)
					json.Unmarshal([]byte(metadataColumn.Data()[i]), &doc.Metadata)
				} else if field.Name() == "created_at" {
					createdAtColumn := field.(*entity.ColumnInt64)
					doc.CreatedAt = time.Unix(createdAtColumn.Data()[i], 0)
				} else if field.Name() == "updated_at" {
					updatedAtColumn := field.(*entity.ColumnInt64)
					doc.UpdatedAt = time.Unix(updatedAtColumn.Data()[i], 0)
				}
			}
			documents = append(documents, doc)
		}
	}
	return documents, nil
}

// Search 搜索向量
func (m *MilvusVDB) Search(ctx context.Context, vector []float32, opts ...vdb.SearchOption) ([]*vdb.Document, error) {
	options := &vdb.SearchOptions{
		TopK: 10,
	}
	for _, opt := range opts {
		opt(options)
	}
	return m.SearchWithOptions(ctx, vector, options)
}

// SearchQuery 搜索查询
func (m *MilvusVDB) SearchQuery(ctx context.Context, query string, opts ...vdb.SearchOption) ([]*vdb.Document, error) {
	options := &vdb.SearchOptions{
		TopK: 10,
	}
	for _, opt := range opts {
		opt(options)
	}

	// 这里需要实现文本到向量的转换逻辑
	vector, err := m.config.Embedder.Embed(ctx, query)
	if err != nil {
		return nil, err
	}
	return m.SearchWithOptions(ctx, vector, options)
}

// SearchQueryWithOptions 使用选项搜索查询
func (m *MilvusVDB) SearchQueryWithOptions(ctx context.Context, query string, opts *vdb.SearchOptions) ([]*vdb.Document, error) {
	vector, err := m.config.Embedder.Embed(ctx, query)
	if err != nil {
		return nil, err
	}
	return m.SearchWithOptions(ctx, vector, opts)
}

// Delete 删除文档
func (m *MilvusVDB) Delete(ctx context.Context, id string) error {
	expr := fmt.Sprintf("id == '%s'", id)
	return m.client.Delete(ctx, m.config.CollectionName, expr, "")
}

// Close 关闭客户端连接
func (m *MilvusVDB) Close() error {
	return m.client.Close()
}

func metricsType(distance vdb.Distance) entity.MetricType {
	metricsType := entity.L2
	if distance == vdb.DistanceCosine {
		metricsType = entity.COSINE
	} else if distance == vdb.DistanceL2 {
		metricsType = entity.L2
	}
	return metricsType
}

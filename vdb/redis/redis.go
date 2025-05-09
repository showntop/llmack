package redis

import (
	"context"
	"fmt"
	"strconv"

	"slices"

	"github.com/redis/go-redis/v9"
	"github.com/showntop/llmack/vdb"
)

var Name = "redis"

func init() {
	vdb.Register(Name, New)
}

type VDB struct {
	client    *redis.Client
	index     string
	dimension int
}

type Config struct {
	redis.Options
	redis.FTCreateOptions
	Index       string
	KeyPrefix   string
	FieldSchema []*redis.FieldSchema
}

func New(config any) (vdb.VDB, error) {
	ctx := context.Background()

	cfg, ok := config.(*Config)
	if !ok {
		return nil, fmt.Errorf("invalid config must be *redis.Config, got: %T", config)
	}
	client := redis.NewClient(&cfg.Options)

	// 测试连接
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("connect redis: %w", err)
	}

	db := &VDB{
		client: client,
		index:  cfg.Index,
		// dimension: cfg.FTCreateOptions.,
	}

	// 确保索引存在
	if err := db.ensureIndex(ctx); err != nil {
		return nil, err
	}

	return db, nil
}

func (r *VDB) Store(ctx context.Context, docs []*vdb.Document) error {
	for _, doc := range docs {
		return r.client.Do(ctx, "HSET", r.index, doc.ID, doc.Metadata).Err()
	}
	return nil
}

func (r *VDB) Search(ctx context.Context, vector []float64, opts ...vdb.SearchOption) ([]vdb.Document, error) {

	// 应用搜索选项
	options := &vdb.SearchOptions{
		TopK:      10,
		Threshold: 0.5,
	}
	for _, opt := range opts {
		opt(options)
	}

	// 构建向量搜索查询
	query := fmt.Sprintf("*=>[KNN %d @vector $BLOB AS score]", options.TopK)

	// 执行搜索
	res := r.client.Do(ctx, "FT.SEARCH", r.index, query,
		"PARAMS", "2", "BLOB", vectorToBytes(vector),
		"RETURN", "4", "id", "query", "answer", "score",
		"LIMIT", "0", strconv.Itoa(options.TopK),
	)

	if res.Err() != nil {
		return nil, fmt.Errorf("search: %w", res.Err())
	}

	// 解析结果
	results, err := res.Slice()
	if err != nil {
		return nil, fmt.Errorf("parse results: %w", err)
	}

	var docs []vdb.Document
	// Redis 返回格式: [total_results doc1_id doc1_fields... doc2_id doc2_fields...]
	for i := 1; i < len(results); i += 2 {
		fields := results[i+1].([]interface{})
		score, _ := strconv.ParseFloat(fields[7].(string), 64)

		if score > options.Threshold {
			id, _ := strconv.ParseInt(fields[1].(string), 10, 64)
			docs = append(docs, vdb.Document{
				ID:       strconv.FormatInt(id, 10),
				Title:    fields[3].(string),
				Metadata: fields[5].(map[string]any),
			})
		}
	}

	return docs, nil
}

func (r *VDB) SearchQuery(ctx context.Context, query string, opts ...vdb.SearchOption) ([]vdb.Document, error) {
	vector, err := r.textToVector(query)
	if err != nil {
		return nil, fmt.Errorf("convert query to vector: %w", err)
	}

	return r.Search(ctx, vector, opts...)
}

func (r *VDB) Delete(ctx context.Context, id string) error {
	err := r.client.Del(ctx, fmt.Sprintf("doc:%s", id)).Err()
	if err != nil {
		return fmt.Errorf("delete: %w", err)
	}
	return nil
}

func (r *VDB) Close() error {
	return r.client.Close()
}

// 私有辅助方法

func (r *VDB) ensureIndex(ctx context.Context) error {
	// 检查索引是否存在
	res := r.client.FT_List(ctx)
	if res.Err() != nil {
		return fmt.Errorf("check index: %w", res.Err())
	}

	indices, err := res.Result()
	if err != nil {
		return fmt.Errorf("parse indices: %w", err)
	}

	// 如果索引不存在，创建索引
	indexExists := slices.Contains(indices, r.index)
	if indexExists {
		return nil
	}

	// schemas should match DocumentToHashes configured in IndexerConfig.
	schemas := []*redis.FieldSchema{
		{
			FieldName: "content",
			FieldType: redis.SearchFieldTypeText,
			Weight:    1,
		},
		{
			FieldName: "vector_content",
			FieldType: redis.SearchFieldTypeVector,
			VectorArgs: &redis.FTVectorArgs{
				// FLAT index: https://redis.io/docs/latest/develop/interact/search-and-query/advanced-concepts/vectors/#flat-index
				// Choose the FLAT index when you have small datasets (< 1M vectors) or when perfect search accuracy is more important than search latency.
				FlatOptions: &redis.FTFlatOptions{
					Type:           "FLOAT32", // BFLOAT16 / FLOAT16 / FLOAT32 / FLOAT64. BFLOAT16 and FLOAT16 require v2.10 or later.
					Dim:            1024,      // keeps same with dimensions of Embedding
					DistanceMetric: "COSINE",  // L2 / IP / COSINE
				},
				// HNSW index: https://redis.io/docs/latest/develop/interact/search-and-query/advanced-concepts/vectors/#hnsw-index
				// HNSW, or hierarchical navigable small world, is an approximate nearest neighbors algorithm that uses a multi-layered graph to make vector search more scalable.
				HNSWOptions: nil,
			},
		},
		{
			FieldName: "extra_field_number",
			FieldType: redis.SearchFieldTypeNumeric,
		},
	}

	// 创建索引
	createIndex := r.client.FTCreate(ctx, r.index,
		&redis.FTCreateOptions{
			OnHash: true,
			Prefix: []any{},
		},
		schemas...,
	)

	if createIndex.Err() != nil {
		return fmt.Errorf("create index: %w", createIndex.Err())
	}

	return nil
}

func (r *VDB) textToVector(text string) ([]float64, error) {
	// TODO: 实现文本到向量的转换
	return nil, fmt.Errorf("not implemented")
}

// 辅助函数：将向量转换为字节
func vectorToBytes(vector []float64) []byte {
	// TODO: 实现向量到字节的转换
	return nil
}

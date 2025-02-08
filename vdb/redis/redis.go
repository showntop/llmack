package redis

import (
	"context"
	"fmt"
	"strconv"

	"github.com/redis/go-redis/v9"
	"github.com/showntop/llmack/vdb"
)

type VDB struct {
	client    *redis.Client
	index     string
	dimension int
}

type Config struct {
	Address   string
	Password  string
	DB        int
	Index     string
	Dimension int
}

func New(cfg Config) (*VDB, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.Address,
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	// 测试连接
	if err := client.Ping(context.Background()).Err(); err != nil {
		return nil, fmt.Errorf("connect redis: %w", err)
	}

	db := &VDB{
		client:    client,
		index:     cfg.Index,
		dimension: cfg.Dimension,
	}

	// 确保索引存在
	if err := db.ensureIndex(); err != nil {
		return nil, err
	}

	return db, nil
}

func (r *VDB) Search(ctx context.Context, vector []float64, opts ...vdb.SearchOption) ([]vdb.Document, error) {
	// 应用搜索选项
	options := &vdb.SearchOptions{
		Topk:      10,
		Threshold: 0.5,
	}
	for _, opt := range opts {
		opt(options)
	}

	// 构建向量搜索查询
	query := fmt.Sprintf("*=>[KNN %d @vector $BLOB AS score]", options.Topk)

	// 执行搜索
	res := r.client.Do(ctx, "FT.SEARCH", r.index, query,
		"PARAMS", "2", "BLOB", vectorToBytes(vector),
		"RETURN", "4", "id", "query", "answer", "score",
		"LIMIT", "0", strconv.Itoa(options.Topk),
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
				ID:     strconv.FormatInt(id, 10),
				Query:  fields[3].(string),
				Answer: fields[5].(string),
				Score:  []float64{score},
				Vector: vector,
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

func (r *VDB) ensureIndex() error {
	ctx := context.Background()

	// 检查索引是否存在
	res := r.client.Do(ctx, "FT._LIST")
	if res.Err() != nil {
		return fmt.Errorf("check index: %w", res.Err())
	}

	indices, err := res.Slice()
	if err != nil {
		return fmt.Errorf("parse indices: %w", err)
	}

	// 如果索引不存在，创建索引
	indexExists := false
	for _, idx := range indices {
		if idx.(string) == r.index {
			indexExists = true
			break
		}
	}

	if !indexExists {
		// 创建索引
		err := r.client.Do(ctx, "FT.CREATE", r.index,
			"ON", "HASH",
			"PREFIX", "1", "doc:",
			"SCHEMA",
			"id", "NUMERIC", "SORTABLE",
			"query", "TEXT",
			"answer", "TEXT",
			"vector", "VECTOR", "FLOAT32",
			fmt.Sprintf("DIM %d", r.dimension),
			"DISTANCE_METRIC", "L2",
		).Err()

		if err != nil {
			return fmt.Errorf("create index: %w", err)
		}
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

package pgvector

import (
	"context"
	"crypto/md5"
	"fmt"
	"strings"

	"github.com/showntop/llmack/embedding"
	"github.com/showntop/llmack/vdb"

	"github.com/jackc/pgx/v5"
	"github.com/pgvector/pgvector-go"
	pgxvec "github.com/pgvector/pgvector-go/pgx"
)

var Name = "pgvector"

func init() {
	vdb.Register(Name, New)
}

type VectorDB struct {
	db     *pgx.Conn
	config *Config
}

type Config struct {
	DNS      string
	Table    string
	Schema   string
	Columns  []string
	Embedder embedding.Embedder
	Distance vdb.Distance
	Index    *IndexConfig
}

type IndexConfig struct {
	Name   string
	Type   string
	Params map[string]any // 支持更多索引参数
}

func New(config any) (vdb.VDB, error) {
	ctx := context.Background()
	cfg, ok := config.(*Config)
	if !ok {
		return nil, fmt.Errorf("invalid config must be *Config, got: %T", config)
	}
	if cfg.Table == "" {
		return nil, fmt.Errorf("table is required")
	}
	if cfg.Schema == "" {
		cfg.Schema = "public"
	}

	db, err := pgx.Connect(ctx, cfg.DNS)
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}
	_, err = db.Exec(ctx, "CREATE EXTENSION IF NOT EXISTS vector")
	if err != nil {
		return nil, fmt.Errorf("create extension: %w", err)
	}

	err = pgxvec.RegisterTypes(ctx, db)
	if err != nil {
		return nil, fmt.Errorf("register types: %w", err)
	}

	// table must exists

	return &VectorDB{
		db:     db,
		config: cfg,
	}, nil
}

func (v *VectorDB) Create(ctx context.Context) error {
	columns := []string{
		"id bigserial PRIMARY KEY",   // primary id
		"title varchar(4092)",        // document title
		"content text",               // document content
		"content_hash varchar(1024)", // document content hash
		fmt.Sprintf("embedding vector(%d)", v.config.Embedder.Dimension()), // embedding vector
		"metadata jsonb",          // document metadata
		"source_id varchar(1024)", // document source id
		"created_at timestamp with time zone default now()", // document created at
		"updated_at timestamp with time zone default now()", // document updated at
	}
	query := fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (%s)", v.config.Table, strings.Join(columns, ", "))
	_, err := v.db.Exec(ctx, query)
	if err != nil {
		return fmt.Errorf("create table: %w", err)
	}
	// create unique index
	query = fmt.Sprintf("CREATE UNIQUE INDEX IF NOT EXISTS idx_source_id ON %s (source_id)", v.config.Table)
	_, err = v.db.Exec(ctx, query)
	if err != nil {
		return fmt.Errorf("create unique index: %w", err)
	}
	if v.config.Index != nil {
		op := getVectorOp(v.config.Distance)
		indexParams := ""
		if v.config.Index.Params != nil {
			params := []string{}
			for k, v := range v.config.Index.Params {
				params = append(params, fmt.Sprintf("%s = %v", k, v))
			}
			if len(params) > 0 {
				indexParams = fmt.Sprintf("WITH (%s)", strings.Join(params, ", "))
			}
		}
		query = fmt.Sprintf(
			"CREATE INDEX IF NOT EXISTS %s ON %s USING %s (embedding %s) %s",
			v.config.Index.Name, v.config.Table, v.config.Index.Type, op, indexParams,
		)
		_, err = v.db.Exec(ctx, query)
		if err != nil {
			return fmt.Errorf("create index: %w", err)
		}
	}
	return nil
}

func (v *VectorDB) Store(ctx context.Context, docs ...*vdb.Document) error {
	columns := []string{"source_id", "title", "content", "content_hash", "embedding", "metadata"}
	query := fmt.Sprintf("INSERT INTO %s (%s) VALUES ($1, $2, $3, $4, $5, $6) ON CONFLICT (source_id) DO UPDATE SET title = $2, content = $3, content_hash = $4, embedding = $5, metadata = $6",
		v.config.Table, strings.Join(columns, ", "))

	// 100 batch insert into table
	batch := &pgx.Batch{}
	for _, doc := range docs {
		contentHash := fmt.Sprintf("%x", md5.Sum([]byte(doc.Content)))
		embedding, err := v.config.Embedder.Embed(ctx, doc.Content)
		if err != nil {
			return fmt.Errorf("embed: %w", err)
		}
		metadata := map[string]any{}
		metadata["source"] = doc.Metadata["source"]
		batch.Queue(query, doc.ID, doc.Title, doc.Content, contentHash, pgvector.NewVector(embedding), metadata)
	}
	br := v.db.SendBatch(ctx, batch)
	defer br.Close()

	_, err := br.Exec()
	if err != nil {
		return fmt.Errorf("exec batch %w", err)
	}

	return nil
}

func (v *VectorDB) SearchWithOptions(ctx context.Context, embedding []float32, options *vdb.SearchOptions) ([]*vdb.Document, error) {
	columns := v.config.Columns
	if len(columns) == 0 {
		// columns = []string{"*"}
		columns = []string{
			fmt.Sprintf("embedding %s $1 as similarity", getDistanceOp(v.config.Distance)),
			"title",
			"content",
			"content_hash",
			"embedding",
			"metadata",
			"source_id",
			"created_at",
			"updated_at",
		}
	}
	tx, err := v.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}

	if v.config.Index != nil {
		if v.config.Index.Type == "ivfflat" && v.config.Index.Params["probes"] != nil {
			tx.Exec(ctx, "SET LOCAL ivfflat.probes = $1;", v.config.Index.Params["probes"])
		} else if v.config.Index.Type == "hnsw" && v.config.Index.Params["ef_search"] != nil {
			tx.Exec(ctx, "SET LOCAL hnsw.ef_search = $1;", v.config.Index.Params["ef_search"])
		}
	}
	query := fmt.Sprintf("SELECT %s FROM %s WHERE TRUE ORDER BY embedding %s $2 LIMIT $3;",
		strings.Join(columns, ", "), v.config.Table, getDistanceOp(v.config.Distance))
	rows, err := tx.Query(ctx, query, pgvector.NewVector(embedding), pgvector.NewVector(embedding), options.TopK)
	if err != nil {
		tx.Rollback(ctx)
		return nil, fmt.Errorf("query: %w", err)
	}
	defer rows.Close()

	docs := make([]*vdb.Document, 0)
	for rows.Next() {
		var doc vdb.Document
		var id int64
		err = rows.Scan(&id, &doc.Title, &doc.Content, &doc.ContentHash, &doc.Embedding, &doc.Metadata, &doc.ID, &doc.CreatedAt, &doc.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("scan: %w", err)
		}
		docs = append(docs, &doc)
	}

	err = tx.Commit(ctx)
	if err != nil {
		tx.Rollback(ctx)
		return nil, fmt.Errorf("commit: %w", err)
	}

	return docs, nil
}

// vector similarity search
func (v *VectorDB) Search(ctx context.Context, embedding []float32, opts ...vdb.SearchOption) ([]*vdb.Document, error) {
	options := &vdb.SearchOptions{}
	for _, opt := range opts {
		opt(options)
	}

	return v.SearchWithOptions(ctx, embedding, options)
}

func (v *VectorDB) SearchQueryWithOptions(ctx context.Context, query string, options *vdb.SearchOptions) ([]*vdb.Document, error) {
	if v.config.Embedder == nil {
		return nil, fmt.Errorf("embedder is not set")
	}
	vector, err := v.config.Embedder.Embed(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("embed: %w", err)
	}
	return v.SearchWithOptions(ctx, vector, options)
}

func (v *VectorDB) SearchQuery(ctx context.Context, query string, opts ...vdb.SearchOption) ([]*vdb.Document, error) {
	if v.config.Embedder == nil {
		return nil, fmt.Errorf("embedder is not set")
	}
	vector, err := v.config.Embedder.Embed(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("embed: %w", err)
	}
	return v.Search(ctx, vector, opts...)
}

func (v *VectorDB) Delete(ctx context.Context, id string) error {
	// TODO: implement
	return nil
}

func (v *VectorDB) Close() error {
	return v.db.Close(context.Background())
}

func getVectorOp(distance vdb.Distance) string {
	switch distance {
	case vdb.DistanceL2:
		return "vector_l2_ops"
	case vdb.DistanceCosine:
		return "vector_cosine_ops"
	default:
		return "vector_ip_ops"
	}
}

func getDistanceOp(distance vdb.Distance) string {
	switch distance {
	case vdb.DistanceL2:
		return "<->"
	case vdb.DistanceCosine:
		return "<=>"
	default:
		return "<#>"
	}
}

package rag

import (
	"context"

	"github.com/showntop/llmack/log"
	"github.com/showntop/llmack/vdb"
)

// Indexer ...
type Indexer struct {
	vdb      vdb.VDB
	scalarDB ScalarDB // object
}

// NewIndexer ...
func NewIndexer(name string, config any) (*Indexer, error) {
	vdb, err := vdb.NewVDB(name, config) // new vdb
	if err != nil {
		return nil, err
	}

	return &Indexer{vdb: vdb}, nil
}

// Retrieve TODO: 实现检索逻辑
func (r *Indexer) Retrieve(ctx context.Context, query string, opts ...SearchOption) ([]*vdb.Document, error) {
	searchOpts := &SearchOptions{}
	for _, opt := range opts {
		opt(searchOpts)
	}
	log.InfoContextf(ctx, "Retrieve knowledge: %v query: %s options: %+v", searchOpts.LibraryID, query, searchOpts)

	vdbSearchOpts := &vdb.SearchOptions{
		TopK:      searchOpts.TopK,
		Threshold: searchOpts.ScoreThreshold,
	}
	docs, err := r.vdb.SearchQueryWithOptions(ctx, query, vdbSearchOpts)
	if err != nil {
		return nil, err
	}

	return docs, nil
}

func (r *Indexer) Index(ctx context.Context, docs []*vdb.Document, opts *SearchOptions) ([]*vdb.Document, error) {
	log.InfoContextf(ctx, "Index documents %+v with options %+v", docs, opts)
	// 将知识库中的数据转换为向量
	if err := r.vdb.Store(ctx, docs...); err != nil {
		return nil, err
	}

	return nil, nil
}

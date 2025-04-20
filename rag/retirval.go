package rag

import (
	"context"

	"github.com/showntop/llmack/log"
	"github.com/showntop/llmack/vdb"
)

// Retrival ...
type Retrival struct {
	vdb      vdb.VDB
	scalarDB ScalarDB // object
}

// NewRetrival ...
func NewRetrival(name string, config any) (*Retrival, error) {
	// new vdb
	vdb, err := vdb.NewVDB(name, config)
	if err != nil {
		return nil, err
	}

	return &Retrival{vdb: vdb}, nil
}

// Retrieve TODO: 实现检索逻辑
func (r *Retrival) Retrieve(ctx context.Context, query string, opts *Options) ([]KnowledgeEntity, error) {
	log.InfoContextf(ctx, "Retrieve knowledge: %v query: %s options: %+v", opts.LibraryID, query, opts)

	// return r.vdb.Search(ctx, query, opts)
	return nil, nil
}

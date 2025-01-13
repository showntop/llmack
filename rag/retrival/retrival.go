package retrival

import (
	"context"

	"github.com/showntop/llmack/log"
	"github.com/showntop/llmack/rag"
	"github.com/showntop/llmack/vdb"
)

// Retrival ...
type Retrival struct {
	vdb      vdb.VDB
	scalarDB rag.ScalarDB // object
}

// NewRetrival ...
func NewRetrival(vdb1 vdb.VDB) *Retrival {
	return &Retrival{vdb: vdb1} // TODO 依赖注入
}

// Retrieve TODO: 实现检索逻辑
func (r *Retrival) Retrieve(ctx context.Context, query string, opts *rag.Options) ([]rag.KnowledgeEntity, error) {
	log.InfoContextf(ctx, "Retrieve knowledge: %v query: %s options: %+v", opts.LibraryID, query, opts)

	// return r.vdb.Search(ctx, query, opts)
	return nil, nil
}

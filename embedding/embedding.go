package embedding

import "context"

// Embedder ...
type Embedder interface {
	Embed(ctx context.Context, text string) ([]float32, error)
	Dimension() int
}

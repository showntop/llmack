package embedding

import (
	"context"

	"github.com/sashabaranov/go-openai"
)

// OpenAIEmbedder OpenAI embedder
type OpenAIEmbedder struct {
	client *openai.Client
	model  openai.EmbeddingModel
}

// NewOpenAIEmbedder 创建一个新的OpenAI embedder实例
func NewOpenAIEmbedder(apiKey string, model openai.EmbeddingModel) *OpenAIEmbedder {
	if model == "" {
		model = openai.AdaEmbeddingV2 // 默认使用 text-embedding-ada-002
	}

	return &OpenAIEmbedder{
		client: openai.NewClient(apiKey),
		model:  model,
	}
}

// Embed 将文本转换为向量
func (e *OpenAIEmbedder) Embed(ctx context.Context, text string) ([]float32, error) {
	resp, err := e.client.CreateEmbeddings(ctx, openai.EmbeddingRequest{
		Input: []string{text},
		Model: e.model,
	})

	if err != nil {
		return nil, err
	}

	if len(resp.Data) == 0 {
		return []float32{}, nil
	}

	return resp.Data[0].Embedding, nil
}

// BatchEmbed 批量将多个文本转换为向量
func (e *OpenAIEmbedder) BatchEmbed(ctx context.Context, texts []string) ([][]float32, error) {
	resp, err := e.client.CreateEmbeddings(ctx, openai.EmbeddingRequest{
		Input: texts,
		Model: e.model,
	})

	if err != nil {
		return nil, err
	}

	embeddings := make([][]float32, len(resp.Data))
	for i, data := range resp.Data {
		embeddings[i] = data.Embedding
	}

	return embeddings, nil
}

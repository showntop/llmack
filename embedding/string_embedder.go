package embedding

import "context"

// StringEmbedder ...
type StringEmbedder struct {
}

// NewStringEmbedder 创建一个新的StringEmbedder实例
var NewStringEmbedder = func() Embedder {
	return StringEmbedder{}
}

// Embed 将文本转换为固定维度的向量
func (s StringEmbedder) Embed(ctx context.Context, text string) ([]float32, error) {
	// 1. 固定向量长度
	vectorSize := s.Dimension()
	result := make([]float32, vectorSize)

	// 2. 获取字符串的字节
	bytes := []byte(text)

	// 3. 使用字节生成向量
	for i := range vectorSize {
		var sum float32 = 0
		// 对每个位置，使用一组字节计算一个float值
		for j := i; j < len(bytes); j += vectorSize {
			sum += float32(bytes[j%len(bytes)])
		}
		result[i] = sum
	}

	return result, nil
}

// BatchEmbed 批量将多个文本转换为向量
func (s StringEmbedder) BatchEmbed(ctx context.Context, texts []string) ([][]float32, error) {
	results := make([][]float32, len(texts))
	for i, text := range texts {
		embedding, err := s.Embed(ctx, text)
		if err != nil {
			return nil, err
		}
		results[i] = embedding
	}
	return results, nil
}

// Dimension ...
func (s StringEmbedder) Dimension() int {
	return 32
}

package embedding

// import (
// 	"fmt"

// 	"github.com/sugarme/tokenizer"
// 	"github.com/sugarme/transformer"
// )

// // TransformerEmbedder 实现基于Transformer的文本嵌入
// type TransformerEmbedder struct {
// 	model      *transformer.Model
// 	tokenizer  *tokenizer.Tokenizer
// 	maxLength  int
// 	dim        int
// 	deviceType string // "cpu" or "cuda"
// }

// // NewTransformerEmbedder 创建新的Transformer嵌入器
// func NewTransformerEmbedder(modelPath, tokenizerPath string, dimension int) (*TransformerEmbedder, error) {
// 	// 加载分词器
// 	tok, err := tokenizer.NewFromFile(tokenizerPath)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to load tokenizer: %v", err)
// 	}

// 	// 加载模型
// 	model, err := transformer.NewModel(modelPath)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to load model: %v", err)
// 	}

// 	return &TransformerEmbedder{
// 		model:      model,
// 		tokenizer:  tok,
// 		maxLength:  512,
// 		dim:        dimension,
// 		deviceType: "cpu",
// 	}, nil
// }

// // Embed 生成文本的向量表示
// func (e *TransformerEmbedder) Embed(text string) ([]float64, error) {
// 	// 对文本进行分词
// 	encoding, err := e.tokenizer.Encode(text, true)
// 	if err != nil {
// 		return nil, fmt.Errorf("tokenization failed: %v", err)
// 	}

// 	// 截断或填充到固定长度
// 	if len(encoding.Ids) > e.maxLength {
// 		encoding.Ids = encoding.Ids[:e.maxLength]
// 		encoding.AttentionMask = encoding.AttentionMask[:e.maxLength]
// 	}

// 	// 转换为模型输入格式
// 	input := transformer.NewInput(
// 		encoding.Ids,
// 		encoding.AttentionMask,
// 		nil, // token type ids
// 	)

// 	// 获取模型输出
// 	output, err := e.model.Forward(input)
// 	if err != nil {
// 		return nil, fmt.Errorf("model inference failed: %v", err)
// 	}

// 	// 处理输出，获取句子嵌入
// 	embeddings := e.poolOutput(output)
// 	return embeddings, nil
// }

// // Dimension 返回嵌入向量的维度
// func (e *TransformerEmbedder) Dimension() int {
// 	return e.dim
// }

// // poolOutput 对模型输出进行池化，得到句子嵌入
// func (e *TransformerEmbedder) poolOutput(output *transformer.Output) []float64 {
// 	// 这里实现具体的池化策略
// 	// 例如：取[CLS]标记的输出，或者对所有token的输出进行平均池化
// 	// 示例实现：取最后一层[CLS]标记的输出
// 	lastLayerOutput := output.LastHiddenState
// 	clsEmbedding := lastLayerOutput[0] // [CLS]标记通常在第一个位置

// 	// 转换为float64
// 	result := make([]float64, len(clsEmbedding))
// 	for i, v := range clsEmbedding {
// 		result[i] = float64(v)
// 	}

// 	return result
// }

// // SetDevice 设置运行设备
// func (e *TransformerEmbedder) SetDevice(deviceType string) {
// 	e.deviceType = deviceType
// }

// // Close 清理资源
// func (e *TransformerEmbedder) Close() error {
// 	if e.model != nil {
// 		e.model.Close()
// 	}
// 	return nil
// }

package embedding

// import (
// 	"fmt"

// 	"github.com/owulveryck/onnx-go"
// 	"github.com/owulveryck/onnx-go/backend/x/gorgonnx"
// )

// // OnnxEmbedder 实现基于ONNX的文本嵌入
// type OnnxEmbedder struct {
// 	model *onnx.Model
// 	dim   int
// }

// // NewOnnxEmbedder 创建新的ONNX嵌入器
// func NewOnnxEmbedder(modelPath string, dimension int) (*OnnxEmbedder, error) {
// 	backend := gorgonnx.NewGraph()
// 	model := onnx.NewModel(backend)

// 	// 加载模型
// 	if err := model.Load(modelPath); err != nil {
// 		return nil, fmt.Errorf("failed to load model: %v", err)
// 	}

// 	return &OnnxEmbedder{
// 		model: model,
// 		dim:   dimension,
// 	}, nil
// }

// // Embed 生成文本的向量表示
// func (e *OnnxEmbedder) Embed(text string) ([]float64, error) {
// 	// 这里需要根据具体的模型实现文本预处理和推理
// 	// 示例实现：
// 	input := preprocessText(text)
// 	output, err := e.model.Predict(input)
// 	if err != nil {
// 		return nil, fmt.Errorf("prediction failed: %v", err)
// 	}

// 	return normalizeOutput(output), nil
// }

// // Dimension 返回嵌入向量的维度
// func (e *OnnxEmbedder) Dimension() int {
// 	return e.dim
// }

// // 文本预处理函数
// func preprocessText(text string) []float32 {
// 	// 实现文本预处理逻辑
// 	// 例如：分词、编码等
// 	return nil
// }

// // 输出归一化
// func normalizeOutput(output interface{}) []float64 {
// 	// 实现输出处理逻辑
// 	// 将模型输出转换为标准化的float64切片
// 	return nil
// }

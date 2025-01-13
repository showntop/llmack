package vdb

// SearchOption 搜索选项
type SearchOption func(*SearchOptions)

// SearchOptions 搜索配置
type SearchOptions struct {
	Topk      int     // 返回结果数量限制
	Threshold float64 // 相似度阈值
}

// WithTopk 设置返回结果数量
func WithTopk(topk int) SearchOption {
	return func(o *SearchOptions) {
		o.Topk = topk
	}
}

// WithThreshold 设置相似度阈值
func WithThreshold(threshold float64) SearchOption {
	return func(o *SearchOptions) {
		o.Threshold = threshold
	}
}

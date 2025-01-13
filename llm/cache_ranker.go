package llm

// Scorer 打分
type Scorer interface {
	Eval(string) float64
}

// ExactMatchScorer 精确匹配排序
type ExactMatchScorer struct {
}

// Eval 精确匹配排序
func (r *ExactMatchScorer) Eval(query string) float64 {
	return 1.0
}

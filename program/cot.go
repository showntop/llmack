package program

// cot chain of thought
type cot struct {
	*predictor
}

// COT ...
func COT() *predictor {
	prefix := "Reasoning: Let's think step by step in order to"
	desc := "${reasoning}"

	p := &cot{}
	p.predictor = Predictor()
	p.predictor.WithOutputField("reasoning", desc, prefix)
	return p.predictor
}

func (p *cot) WithRationaleField(infos ...any) *predictor {
	p.predictor.WithOutputField(infos...)
	return p.predictor
}

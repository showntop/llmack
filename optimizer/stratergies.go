package optimizer

import "github.com/showntop/llmack/prompt"

type Strategy interface {
	Mutate(p *prompt.Prompt) *prompt.Prompt
}

// Example mutation strategy
type TemplateModifier struct {
	// rules []Rule
}

func (tm *TemplateModifier) Mutate(p *prompt.Prompt) *prompt.Prompt {
	// newPrompt := p.Clone()
	newPrompt := p
	// Apply various mutation rules
	// - Modify instruction clarity
	// - Adjust few-shot examples
	// - Add/remove constraints
	return newPrompt
}

package program

import (
	"reflect"
	"strings"

	"github.com/showntop/llmack/llm"
	"github.com/showntop/llmack/prompt"
)

// RawAdapter ...
type RawAdapter struct {
	target any
}

// Format ...
func (ada *RawAdapter) Format(p *predictor, inputs map[string]any, _ any) ([]llm.Message, error) {
	var err error
	p.Instruction, err = prompt.Render(p.Instruction, inputs)
	if err != nil {
		return nil, err
	}
	// @TODO handle error
	userPromptBuilder := strings.Builder{}
	userPromptBuilder.WriteString(p.Instruction)
	userPromptBuilder.WriteByte('\n')

	messages := []llm.Message{
		llm.UserTextPromptMessage(userPromptBuilder.String()),
	}
	return messages, nil
}

// Parse ...
func (ada *RawAdapter) Parse(completion string, target any) error {
	// 给 target 赋值为 completion
	reflect.ValueOf(target).Elem().SetString(completion)
	return nil
}

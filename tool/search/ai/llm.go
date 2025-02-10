package ai

import (
	"context"

	"github.com/showntop/llmack/llm"
	oaic "github.com/showntop/llmack/llm/openai-c"
	promptx "github.com/showntop/llmack/prompt"
)

func generate(ctx context.Context, prompt string, inputs map[string]any) string {

	formatter := promptx.NewTemplateFormatter(prompt, false)
	query := formatter.Format(inputs, true)
	// core.SetLogger(&core.WrapLogger{})
	result, err := llm.NewInstance(oaic.Name).Invoke(context.Background(), []llm.Message{
		llm.SystemPromptMessage(" "), llm.UserTextPromptMessage(query),
	}, nil,
		llm.WithModel("hunyuan-standard"),
	)
	if err != nil {
		panic(err)
	}

	final := ""
	for v := result.Stream().Next(); v != nil; v = result.Stream().Next() {
		final += string(v.Delta.Message.Content())
	}

	return final
}

func streamx(ctx context.Context, prompt string, inputs map[string]any) *llm.Stream {

	formatter := promptx.NewTemplateFormatter(prompt, false)
	query := formatter.Format(inputs, true)

	result, err := llm.NewInstance(oaic.Name).Invoke(context.Background(), []llm.Message{
		llm.SystemPromptMessage(" "), llm.UserTextPromptMessage(query),
	}, nil,
		llm.WithModel("hunyuan-standard"),
	)
	if err != nil {
		panic(err)
	}

	return result.Stream()
}

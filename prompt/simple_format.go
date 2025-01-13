package prompt

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/showntop/llmack/llm"
	"github.com/showntop/llmack/log"
)

// SimplePromptFormatter ...
type SimplePromptFormatter struct {
}

// Format ...
func (p *SimplePromptFormatter) Format(preset string,
	inputs map[string]any, query string, contexts string) ([]llm.Message, []string) {
	if preset != "" {
		formatter := NewTemplateFormatter(preset, false)
		preset = formatter.Format(inputs, true)
	}
	log.DefaultLogger().InfoContextf(context.Background(), "contexts: %v", contexts)
	frame := p.getPromptFrame()

	sysPrompt := ""
	variables := []string{}
	for _, v := range frame["system_prompt_orders"].([]any) {
		vname := v.(string)
		log.DefaultLogger().InfoContextf(context.Background(), "vname: %v", vname)
		if "context" == vname && contexts != "" {
			if c, ok := frame[vname].(string); ok {
				sysPrompt += c
			}
		}
		if "query" == vname && query != "" {
			if c, ok := frame[vname].(string); ok {
				sysPrompt += c
			}
		}
		if "preset" == vname {
			sysPrompt += preset + "\n"
		}

		variables = append(variables, "#"+v.(string)+"#")
	}

	sysPrompt = strings.ReplaceAll(sysPrompt, "{{#context#}}", contexts)

	sysMessage := llm.SystemPromptMessage(sysPrompt)

	stops, ok := frame["stops"].(string)
	if ok {
		return []llm.Message{sysMessage}, []string{stops}
	}
	return []llm.Message{sysMessage}, nil
}

const simplePromptFrame = `{
  "context": "用户在与一个客观的助手对话。助手会尽量尊重找到的材料，给出全面专业的解释，但不会过度演绎。以下为材料内容：\n\n'''\n{{#context#}}\n'''\n，以下为用户的要求：\n",
  "system_prompt_orders": [
    "context",
    "preset"
  ],
  "query": "{{#query#}}",
  "stops": null
}`

func (p *SimplePromptFormatter) getPromptFrame() map[string]any {
	// os.ReadFile("./simple_prompt.json")
	prompt := make(map[string]any)
	json.Unmarshal([]byte(simplePromptFrame), &prompt)
	return prompt
}

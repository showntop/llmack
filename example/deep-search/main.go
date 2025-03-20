package main

import (
	"context"
	"fmt"
	"os"

	"github.com/joho/godotenv"

	"github.com/showntop/llmack/engine"
	"github.com/showntop/llmack/example/agentic-search/tools"
	"github.com/showntop/llmack/llm"
	"github.com/showntop/llmack/llm/qwen"
	"github.com/showntop/llmack/log"
	"github.com/showntop/llmack/tool"
	"github.com/showntop/llmack/tool/search"
)

func init() {
	log.SetLogger(&log.WrapLogger{})

	godotenv.Load()
	llm.WithSingleConfig(map[string]any{
		// "api_key":  os.Getenv("deepseek_api_key2"),
		// "base_url": "https://api.lkeap.cloud.tencent.com/v1",
		"api_key": os.Getenv("qwen_api_key"),
		// "base_url": os.Getenv("hunyuan_base_url"),
		// "api_key":  os.Getenv("hunyuan_api_key"),
	})

	tool.WithConfig(map[string]any{
		"serper": map[string]any{
			"api_key": os.Getenv("serper_api_key"),
		},
		"minimax": map[string]any{
			"api_key": os.Getenv("minmax_api_key"),
		},
		"siliconflow": map[string]any{
			"api_key": os.Getenv("siliconflow_api_key"),
		},
	})
}

func main() {
	settings := engine.DefaultSettings()
	// settings.PresetPrompt = prompt.CreatePictureBookPrompt
	// settings.LLMModel.Provider = deepseek.Name
	// settings.LLMModel.Name = "deepseek-v3"

	// settings.LLMModel.Provider = openaic.Name
	// settings.LLMModel.Name = "hunyuan-large"
	settings.LLMModel.Provider = qwen.Name
	settings.LLMModel.Name = "qwen-plus"
	// settings.LLMModel.Provider = hunyuan.Name
	settings.PresetPrompt = Propmt
	settings.Agent.Mode = "ReAct"
	settings.Tools = append(settings.Tools, tools.Subqueries, search.DuckDuckGo)
	eng := engine.NewAgentEngine(settings, engine.WithLogger(&log.WrapLogger{}))
	esm := eng.Execute(context.Background(), engine.Input{
		Query: "",
		Inputs: map[string]any{
			"question": "各大平台 SFT 的流程是什么",
		},
	})
	finalAnswer := ""
	for evt := esm.Next(); evt != nil; evt = esm.Next() {
		if evt.Error != nil {
			panic(evt.Error)
		}
		if cv, ok := evt.Data.(*llm.Chunk); ok {
			finalAnswer += cv.Delta.Message.Content()
		}
	}

	fmt.Println(finalAnswer)
}

var Propmt = `
[Goals]
To answer the question gaven by the user, it will be put below.
Question: {{question}}

[instructions]
- For greetings, casual conversation, general knowledge questions answer directly.
- Take advantage of your multi-step reasoning skills to answer.
- provide deep, unexpected insights, identifying hidden patterns and connections, and creating "aha moments.".
- break conventional thinking, establish unique cross-disciplinary connections, and bring new perspectives to the user.
- If uncertain, use reflect
`

// - to archive the goal of answer the question more comprehensively, you can break down the original question into up to four sub-questions. Return as list of str.
// If this is a very simple question and no decomposition is necessary, then keep the only one original question in the list.
// <EXAMPLE>
// Example input:
// "Explain deep learning"

// Example output:
// [
//     "What is deep learning?",
//     "What is the difference between deep learning and machine learning?",
//     "What is the history of deep learning?"
// ]
// </EXAMPLE>

// Provide your response in list of str format:

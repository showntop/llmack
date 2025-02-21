package main

import (
	"context"
	"fmt"
	"os"

	"github.com/joho/godotenv"

	"github.com/showntop/llmack/engine"
	"github.com/showntop/llmack/llm"
	"github.com/showntop/llmack/llm/deepseek"
	"github.com/showntop/llmack/log"
	"github.com/showntop/llmack/tool"

	"github.com/showntop/llmack/tool/datetime"
	"github.com/showntop/llmack/tool/search"
)

func init() {
	godotenv.Load()
	llm.WithSingleConfig(map[string]any{
		"api_key":  os.Getenv("deepseek_api_key2"),
		"base_url": "https://api.lkeap.cloud.tencent.com/v1",
		// "api_key": os.Getenv("qwen_api_key"),
	})

	tool.WithConfig(map[string]any{
		"serper": map[string]any{
			"api_key": os.Getenv("serper_api_key"),
		},
	})
}

func main() {
	settings := engine.DefaultSettings()
	settings.PresetPrompt = "你是一个AI写作助手，使用网络信息写文章"
	settings.LLMModel.Provider = deepseek.Name
	// settings.LLMModel.Provider = qwen.Name
	// settings.LLMModel.Name = "qwen-plus"
	settings.LLMModel.Name = "deepseek-v3"
	settings.Tools = append(settings.Tools, datetime.GetDate, search.Serper)
	eng := engine.NewAgentEngine(settings, engine.WithLogger(&log.WrapLogger{}))
	esm := eng.Execute(context.Background(), engine.Input{
		Query: `主题：科技能改变世界吗？`,
	})
	for evt := esm.Next(); evt != nil; evt = esm.Next() {
		if evt.Error != nil {
			panic(evt.Error)
		}
		if cv, ok := evt.Data.(*llm.Chunk); ok {
			_ = cv
			fmt.Println("main chunk:", cv.Delta.Message)
		} else {
			// fmt.Println("main event name:", evt.Name, ", data:", evt.Data)
		}
	}
}

package main

import (
	"context"
	"fmt"
	"os"

	"github.com/joho/godotenv"

	"github.com/showntop/llmack/engine"
	"github.com/showntop/llmack/llm"
	"github.com/showntop/llmack/llm/deepseek"
	"github.com/showntop/llmack/llm/qwen"
	"github.com/showntop/llmack/log"
	"github.com/showntop/llmack/tool"
	"github.com/showntop/llmack/tool/image"
)

func init() {
	log.SetLogger(&log.WrapLogger{})

	godotenv.Load()
	llm.WithSingleConfig(map[string]any{
		// "api_key":  os.Getenv("deepseek_api_key2"),
		// "base_url": "https://api.lkeap.cloud.tencent.com/v1",
		"api_key": os.Getenv("qwen_api_key"),
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
	settings.PresetPrompt = "你是一个AI绘本专家"
	settings.LLMModel.Provider = deepseek.Name
	settings.LLMModel.Provider = qwen.Name
	settings.LLMModel.Name = "qwen-plus"
	// settings.LLMModel.Name = "deepseek-v3"
	settings.Tools = append(settings.Tools, image.SiliconflowImageGenerate)
	eng := engine.NewAgentEngine(settings, engine.WithLogger(&log.WrapLogger{}))
	esm := eng.Execute(context.Background(), engine.Input{
		Query: "使用工具，制作绘本，主题：月亮和星星",
	})
	for evt := esm.Next(); evt != nil; evt = esm.Next() {
		if evt.Error != nil {
			panic(evt.Error)
		}
		// fmt.Println("main event name:", evt.Name, ", data:", evt.Data)
		if cv, ok := evt.Data.(*llm.Chunk); ok {
			_ = cv
			fmt.Println("main chunk:", cv.Delta.Message)
		} else {
			// fmt.Println("main event name:", evt.Name, ", data:", evt.Data)
		}
	}
}

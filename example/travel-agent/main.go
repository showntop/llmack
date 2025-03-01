package main

import (
	"context"
	"fmt"
	"os"

	"github.com/joho/godotenv"

	"github.com/showntop/llmack/engine"
	"github.com/showntop/llmack/llm"
	openaic "github.com/showntop/llmack/llm/openai-c"
	"github.com/showntop/llmack/log"

	"github.com/showntop/llmack/tool/weather"
)

func init() {
	log.SetLogger(&log.WrapLogger{})

	godotenv.Load()
	llm.WithSingleConfig(map[string]any{
		// "api_key": os.Getenv("deepseek_api_key2"),
		// "base_url": "https://api.lkeap.cloud.tencent.com/v1",
		// "base_url": "https://dashscope.aliyuncs.com/compatible-mode/v1",
		"base_url": os.Getenv("hunyuan_base_url"),
		"api_key":  os.Getenv("hunyuan_api_key"),
		// "api_key":  os.Getenv("zhipu_api_key"),
	})
}

func main() {
	settings := engine.DefaultSettings()
	settings.PresetPrompt = "你是一个AI助手，请帮我处理以下问题：\n{{query}}"
	// settings.LLMModel.Provider = deepseek.Name
	settings.LLMModel.Provider = openaic.Name
	settings.LLMModel.Name = "hunyuan-large"
	// settings.LLMModel.Name = "deepseek-v3"
	// settings.LLMModel.Provider = zhipu.Name
	// settings.LLMModel.Name = "glm-4v-flash"
	settings.Tools = append(settings.Tools, weather.QueryWeather)
	settings.Agent.Mode = "ReAct"
	eng := engine.NewAgentEngine(settings, engine.WithLogger(&log.WrapLogger{}))
	esm := eng.Execute(context.Background(), engine.Input{
		Inputs: map[string]any{
			"query": "我要去北京旅游，规划一下旅程。",
		},
	})
	for evt := esm.Next(); evt != nil; evt = esm.Next() {
		if evt.Error != nil {
			panic(evt.Error)
		}
		if cv, ok := evt.Data.(*llm.Chunk); ok {
			fmt.Println("main chunk:", cv.Delta.Message)
		} else {
			fmt.Println("main event name:", evt.Name, ", data:", evt.Data)
		}
	}
}

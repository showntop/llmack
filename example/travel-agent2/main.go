package main

import (
	"context"
	"fmt"
	"os"

	"github.com/joho/godotenv"

	"github.com/showntop/llmack/engine"
	"github.com/showntop/llmack/llm"
	"github.com/showntop/llmack/llm/qwen"
	"github.com/showntop/llmack/log"

	"github.com/showntop/llmack/tool/datetime"
	"github.com/showntop/llmack/tool/weather"
)

func init() {
	godotenv.Load()
	llm.WithSingleConfig(map[string]any{
		// "api_key":  os.Getenv("deepseek_api_key2"),
		// "base_url": "https://api.lkeap.cloud.tencent.com/v1",
		"api_key": os.Getenv("qwen_api_key"),
	})
}

func main() {
	settings := engine.DefaultSettings()
	settings.PresetPrompt = "你是一个AI助手"
	// settings.LLMModel.Provider = deepseek.Name
	settings.LLMModel.Provider = qwen.Name
	settings.LLMModel.Name = "qwen-plus"
	settings.Tools = append(settings.Tools, datetime.GetDate, weather.QueryWeather)
	eng := engine.NewAgentEngine(settings, engine.WithLogger(&log.WrapLogger{}))
	esm := eng.Execute(context.Background(), engine.Input{
		Query: "我明天要去北京旅游，根据日期，看一下天气，规划一下旅程。",
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

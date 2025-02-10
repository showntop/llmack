package main

import (
	"context"
	"fmt"
	"os"

	"github.com/joho/godotenv"

	"github.com/showntop/llmack/engine"
	"github.com/showntop/llmack/llm"
	"github.com/showntop/llmack/llm/hunyuan"
	"github.com/showntop/llmack/log"
)

func init() {
	godotenv.Load()
	llm.WithSingleConfig(map[string]any{
		"base_url": os.Getenv("hunyuan_base_url"),
		"api_key":  os.Getenv("hunyuan_api_key"),
	})
}

func main() {
	settings := engine.DefaultSettings()
	settings.PresetPrompt = "你是一个AI助手，请帮我查询一下天气，然后规划今天的旅程。"
	settings.LLMModel.Provider = hunyuan.Name
	settings.LLMModel.Name = "hunyuan-turbo"
	eng := engine.NewAgentEngine(settings, engine.WithLogger(&log.WrapLogger{}))
	esm := eng.Stream(context.Background(), engine.Input{
		Query: "你好",
	})
	for evt := esm.Next(); evt != nil; evt = esm.Next() {
		if evt.Error != nil {
			panic(evt.Error)
		}
		if cv, ok := evt.Data.(*llm.Chunk); ok {
			fmt.Println("main chunk:", cv.Delta.Message.Content())
		} else {
			fmt.Println("main event name:", evt.Name, "data:", evt.Data)
		}
	}
}

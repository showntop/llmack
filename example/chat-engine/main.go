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
	settings.LLMModel.Provider = openaic.Name
	settings.LLMModel.Name = "hunyuan-turbo"
	eng := engine.NewChatEngine(settings, engine.WithLogger(&log.WrapLogger{}))
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

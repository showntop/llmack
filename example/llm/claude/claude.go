package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/joho/godotenv"
	"github.com/showntop/llmack/llm"
	"github.com/showntop/llmack/llm/anthropic"
)

func init() {
	godotenv.Load()
}

func main() {
	ctx := context.Background()
	llm.WithSingleConfig(map[string]any{
		"api_key":  os.Getenv("claude_api_key"),
		"base_url": "http://v2.open.venus.oa.com/llmproxy",
	})

	resp, err := llm.NewInstance(anthropic.Name).Invoke(ctx,
		// []llm.Message{llm.UserPromptMessage("Prove that all entire functions that are also injective take the form f (z) = az + 6 with a, b € C, and a ‡ 0.")},
		[]llm.Message{llm.NewUserTextMessage("你好")},
		llm.WithStream(true),
		llm.WithModel("claude-3-7-sonnet-20250219"))
	if err != nil {
		panic(err)
	}
	// fmt.Println(resp.Result())
	for it := resp.Stream().Take(); it != nil; it = resp.Stream().Take() {
		xxx, _ := json.Marshal(it)
		fmt.Println(string(xxx))
		// fmt.Println("final: ", it.Delta.Message)
	}
}

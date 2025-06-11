package main

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/joho/godotenv"

	"github.com/showntop/llmack/llm"
	"github.com/showntop/llmack/llm/doubao"
)

func init() {
	godotenv.Load()
}

func main() {
	ctx := context.Background()
	llm.WithSingleConfig(map[string]any{
		// "api_key":  os.Getenv("doubao_api_key"),
		"base_url": "https://ark.cn-beijing.volces.com/api/v3",
	})

	resp, err := llm.NewInstance(doubao.Name).Invoke(ctx,
		// []llm.Message{llm.UserPromptMessage("Prove that all entire functions that are also injective take the form f (z) = az + 6 with a, b € C, and a ‡ 0.")},
		[]llm.Message{llm.NewUserTextMessage("你好")},
		llm.WithModel("doubao-1.5-ui-tars-250328"),
		llm.WithStream(true),
	)
	if err != nil {
		panic(err)
	}
	for it := resp.Stream().Take(); it != nil; it = resp.Stream().Take() {
		xxx, _ := json.Marshal(it)
		fmt.Println(string(xxx))
		// fmt.Println("final: ", it.Delta.Message)
	}
}

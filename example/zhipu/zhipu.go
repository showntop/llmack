package main

import (
	"context"
	"fmt"
	"os"

	"github.com/joho/godotenv"

	"github.com/showntop/llmack/llm"
	"github.com/showntop/llmack/llm/zhipu"
)

func init() {
	godotenv.Load()
}

func main() {
	ctx := context.Background()
	llm.WithSingleConfig(map[string]any{
		"api_key": os.Getenv("zhipu_api_key"),
	})

	resp, err := llm.NewInstance(zhipu.Name).Invoke(ctx, []llm.Message{
		llm.UserPromptMessage("你好"),
	}, []llm.PromptMessageTool{}, llm.WithModel("GLM-4-Flash"))
	if err != nil {
		panic(err)
	}
	fmt.Println(resp.Result())
}

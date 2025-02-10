package main

import (
	"context"
	"fmt"
	"os"

	"github.com/joho/godotenv"

	"github.com/showntop/llmack/llm"
	openaic "github.com/showntop/llmack/llm/openai-c"
	"github.com/showntop/llmack/log"
)

func init() {
	godotenv.Load()
}

func main() {
	runWithCache()
}

func runWithCache() {
	ctx := context.Background()

	llm.WithSingleConfig(map[string]any{
		"base_url": os.Getenv("hunyuan_base_url"),
		"api_key":  os.Getenv("hunyuan_api_key"),
	})

	instance := llm.NewInstance(openaic.Name,
		llm.WithCache(llm.NewMemoCache()),
		llm.WithLogger(&log.WrapLogger{}),
	)

	resp, err := instance.Invoke(ctx,
		[]llm.Message{llm.UserTextPromptMessage("你好")},
		llm.WithModel("hunyuan"),
		llm.WithStream(true),
	)
	if err != nil {
		panic(err)
	}
	fmt.Println(resp.Result())

	resp, err = instance.Invoke(ctx,
		[]llm.Message{llm.UserTextPromptMessage("你好")},
		llm.WithModel("hunyuan"),
		llm.WithStream(true),
	)
	if err != nil {
		panic(err)
	}
	// stream := resp.Stream()
	// for v := stream.Next(); v != nil; v = stream.Next() {
	// 	fmt.Println(string(v.Delta.Message.Content().Data))
	// }
	fmt.Println(resp.Result())
}

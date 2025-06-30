package main

import (
	"context"
	"fmt"

	"github.com/showntop/llmack/llm"
	_ "github.com/showntop/llmack/llm/ollama"
)

func main() {
	ctx := context.Background()

	// llm.WithConfigs(map[string]any{
	// 	"ollama": map[string]any{
	// 		"endpoint": "http://127.0.0.1:4000",
	// 	},
	// })

	llm.SetSingleConfig(map[string]any{
		"base_url": "http://127.0.0.1:11434",
	})

	resp, err := llm.New("ollama").Invoke(ctx, []llm.Message{
		llm.NewUserTextMessage("你好"),
	}, llm.WithModel("llama3.1"), llm.WithStream(true))
	if err != nil {
		panic(err)
	}
	// stream := resp.Stream()
	// for v := stream.Next(); v != nil; v = stream.Next() {
	// 	fmt.Println(string(v.Delta.Message.Content().Data))
	// }
	fmt.Println(resp.Result())
}

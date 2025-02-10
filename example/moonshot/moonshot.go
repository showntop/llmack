package main

import (
	"context"
	"fmt"
	"os"

	"github.com/joho/godotenv"

	"github.com/showntop/llmack/llm"
	"github.com/showntop/llmack/llm/moonshot"
)

func init() {
	godotenv.Load()
}

func main() {
	ctx := context.Background()
	llm.WithSingleConfig(map[string]any{
		"api_key": os.Getenv("moonshot_api_key"),
	})

	resp, err := llm.NewInstance(moonshot.Name).Invoke(ctx, []llm.Message{
		llm.UserTextPromptMessage("你好"),
	}, llm.WithModel("moonshot-v1-8k"))
	if err != nil {
		panic(err)
	}
	// stream := resp.Stream()
	// for v := stream.Next(); v != nil; v = stream.Next() {
	// 	fmt.Println(string(v.Delta.Message.Content().Data))
	// }
	fmt.Println(resp.Result())

}

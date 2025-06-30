package main

import (
	"context"
	"fmt"
	"os"

	"github.com/joho/godotenv"
	"github.com/showntop/llmack/llm"
	openaic "github.com/showntop/llmack/llm/openai-c"
)

func init() {
	godotenv.Load()
}

func main() {
	ctx := context.Background()

	llm.SetSingleConfig(map[string]any{
		"base_url": os.Getenv("hunyuan_base_url"),
		"api_key":  os.Getenv("hunyuan_api_key"),
	})

	resp, err := llm.New(openaic.Name).Invoke(ctx, []llm.Message{
		llm.NewUserMultipartMessage(
			llm.MultipartContentImageURL("https://img.tukuppt.com/bg_grid/05/37/54/v40ZCaqERa.jpg!/fh/350"),
			llm.MultipartContentText("给这张图片添加一个太阳"),
		),
	},
		llm.WithModel("hunyuan-vision"),
		llm.WithStream(true),
	)

	if err != nil {
		panic(err)
	}
	for it := resp.Stream().Take(); it != nil; it = resp.Stream().Take() {
		if len(it.Choices) > 0 {
			fmt.Println("final: ", it.Choices[0].Delta.String())
		}
	}
}

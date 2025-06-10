package main

import (
	"context"
	"fmt"
	"os"

	"github.com/joho/godotenv"
	"github.com/showntop/llmack/agent"
	"github.com/showntop/llmack/llm"
	"github.com/showntop/llmack/llm/deepseek"
	"github.com/showntop/llmack/log"
)

func init() {
	godotenv.Load()

	log.SetLogger(&log.WrapLogger{})

	// llm.WithSingleConfig(map[string]any{
	// 	"api_key": os.Getenv("qwen_api_key"),
	// })
	llm.WithSingleConfig(map[string]any{
		"api_key": os.Getenv("deepseek_api_key"),
	})
}

func main() {

	androidAgent := agent.NewAndroidAgent("android use agent",
		// agent.WithModel(llm.NewInstance(qwen.Name, llm.WithDefaultModel("qwen-vl-max-latest"))),
		agent.WithModel(llm.NewInstance(deepseek.Name, llm.WithDefaultModel("deepseek-chat"))),
	)

	response := androidAgent.Invoke(context.Background(), "open wechat")
	if response.Error != nil {
		panic(response.Error)
	}
	fmt.Println(response.Completion())
}

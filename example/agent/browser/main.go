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
	"github.com/showntop/llmack/pkg/browser"
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

	browserAgent := agent.NewBrowserAgent("browser agent",
		// agent.WithModel(llm.NewInstance(qwen.Name, llm.WithDefaultModel("qwen-vl-max-latest"))),
		agent.WithModel(llm.NewInstance(deepseek.Name, llm.WithDefaultModel("deepseek-chat"))),
		agent.WithBrowserConfig(&browser.BrowserConfig{
			"headless": false,
		}),
	)

	response := browserAgent.Invoke(context.Background(), "去美团外卖查找所有烤鸭店的电话号码、名称、地址")
	if response.Error != nil {
		panic(response.Error)
	}
	fmt.Println(response.Completion())
}

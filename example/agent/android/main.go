package main

import (
	"context"
	"fmt"
	"os"

	"github.com/joho/godotenv"
	"github.com/showntop/llmack/agent"
	"github.com/showntop/llmack/llm"
	"github.com/showntop/llmack/llm/claude"
	"github.com/showntop/llmack/log"
)

func init() {
	godotenv.Load()

	log.SetLogger(&log.WrapLogger{})

	llm.WithConfigs(map[string]any{
		"doubao": map[string]any{
			"base_url": "https://ark.cn-beijing.volces.com/api/v3",
			"api_key":  os.Getenv("doubao_api_key"),
		},
		"claude": map[string]any{
			"base_url": "http://v2.open.venus.oa.com/llmproxy",
			"api_key":  os.Getenv("claude_api_key"),
		},
	})

	// llm.WithSingleConfig(map[string]any{
	// 	"base_url": "https://ark.cn-beijing.volces.com/api/v3",
	// 	"api_key":  os.Getenv("doubao_api_key"),
	// 	// "base_url": "http://v2.open.venus.oa.com/llmproxy",
	// 	// "api_key":  os.Getenv("claude_api_key"),
	// })

	// llm.WithSingleConfig(map[string]any{
	// 	"api_key": os.Getenv("qwen_api_key"),
	// })
	// llm.WithSingleConfig(map[string]any{
	// 	"api_key": os.Getenv("deepseek_api_key"),
	// })
}

func main() {

	androidAgent := agent.NewMobileAgent("mobile use agent",
		// agent.WithModel(llm.NewInstance(claude.Name, llm.WithDefaultModel("doubao-1-5-ui-tars-250428"))),
		agent.WithModel(llm.NewInstance(claude.Name, llm.WithDefaultModel("claude-3-7-sonnet-20250219"))),
		// agent.WithModel(llm.NewInstance(qwen.Name, llm.WithDefaultModel("qwen-vl-max-latest"))),
		// agent.WithModel(llm.NewInstance(deepseek.Name, llm.WithDefaultModel("deepseek-chat"))),
	)

	response := androidAgent.Invoke(context.Background(), "open browser", agent.WithMaxIterationNum(10))
	if response.Error != nil {
		panic(response.Error)
	}
	fmt.Println(response.Completion())
}

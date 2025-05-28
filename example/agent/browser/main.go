package main

import (
	"context"
	"fmt"
	"os"

	"github.com/joho/godotenv"
	"github.com/showntop/llmack/agent"
	"github.com/showntop/llmack/llm"
	"github.com/showntop/llmack/llm/qwen"
	"github.com/showntop/llmack/log"
)

func init() {
	godotenv.Load()

	log.SetLogger(&log.WrapLogger{})

	llm.WithSingleConfig(map[string]any{
		"api_key": os.Getenv("qwen_api_key"),
	})

}

func main() {

	browserAgent := agent.NewBrowserAgent("browser agent",
		agent.WithModel(llm.NewInstance(qwen.Name, llm.WithDefaultModel("qwen-vl-max-latest"))),
	)

	response := browserAgent.Invoke(context.Background(), "Find todays DOW stock price")
	if response.Error != nil {
		panic(response.Error)
	}
	fmt.Println(response.Completion())
}

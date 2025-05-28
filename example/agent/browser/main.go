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

	llm.WithSingleConfig(map[string]any{
		"api_key": os.Getenv("qwen_api_key"),
	})

	llm.SetLogger(&log.WrapLogger{})

}

func main() {

	browserAgent := agent.NewBrowserAgent("browser agent",
		agent.WithModel(llm.NewInstance(qwen.Name, llm.WithDefaultModel("qwen-vl-max-latest"))),
	)

	response := browserAgent.Invoke(context.Background(), "what is the weather in tokyo")
	if response.Error != nil {
		panic(response.Error)
	}
	fmt.Println(response.Completion())
}

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
	"github.com/showntop/llmack/memory"
)

var (
	model = llm.New(deepseek.Name, llm.WithDefaultModel("deepseek-chat"))
)

func init() {
	godotenv.Load()

	log.SetLogger(&log.WrapLogger{})
	llm.SetSingleConfig(map[string]any{
		"api_key": os.Getenv("deepseek_api_key"),
	})
}

func main() {
	sessionID := "session_1"

	agentx := agent.NewAgent("",
		agent.WithMemory(memory.NewFlatMemory(nil)),
		agent.WithModel(model),
	)

	response := agentx.Invoke(context.Background(),
		"My name is John Doe and I like to hike in the mountains on weekends.",
		agent.WithStream(false),
		agent.WithSessionID(sessionID),
	)
	if response.Error != nil {
		panic(response.Error)
	}
	fmt.Println(response.Answer)

	response2 := agentx.Invoke(context.Background(),
		"What are my hobbies?",
		agent.WithStream(false),
		agent.WithSessionID(sessionID),
	)
	if response2.Error != nil {
		panic(response2.Error)
	}
	fmt.Println(response2.Answer)
}

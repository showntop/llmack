package main

import (
	"context"
	"fmt"
	"os"

	"github.com/joho/godotenv"
	"github.com/showntop/llmack/llm"
	"github.com/showntop/llmack/llm/deepseek"
	"github.com/showntop/llmack/program"
)

func init() {
	godotenv.Load()

	llm.WithConfigs(map[string]any{
		deepseek.Name: map[string]any{
			"api_key": os.Getenv("deepseek_api_key"),
		},
	})
	program.SetLLM(deepseek.Name, "deepseek-chat")
}

func main() {
	var target any
	err := program.COT().
		WithInstruction("hello？").
		WithInputField("hello", "world").
		WithOutputField("hello", "world").
		Result(context.Background(), &target)
	if err != nil {
		panic(err)
	}
	fmt.Println(target)
}

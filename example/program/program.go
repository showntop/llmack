package main

import (
	"context"
	"fmt"
	"os"
	"reflect"

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
	var target struct {
		Out1 string `json:"out1"`
		Out2 bool   `json:"out2"`
	}
	err := program.COT().
		WithAdapter(&program.MarkableAdapter{}).
		WithInstruction("hello？").
		WithInputField("hello", "world").
		WithOutputField("out1", "description", "marker", reflect.String).
		WithOutputField("out2", "description", "marker", reflect.Bool).
		Result(context.Background(), &target)
	if err != nil {
		panic(err)
	}
	fmt.Println(target)
}

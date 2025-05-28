package main

import (
	"context"

	"github.com/playwright-community/playwright-go"
	"github.com/showntop/llmack/agent"
	"github.com/showntop/llmack/llm"
	"github.com/showntop/llmack/llm/qwen"
)

func main() {
	playwright.Install()

	browserAgent := agent.NewBrowserAgent("browser agent",
		agent.WithModel(llm.NewInstance(qwen.Name, llm.WithDefaultModel("qwen-vl-max-latest"))),
	)

	browserAgent.Invoke(context.Background(), "what is the weather in tokyo")
}

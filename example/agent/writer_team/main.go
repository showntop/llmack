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
	"github.com/showntop/llmack/tool/search"
)

var (
	model = llm.NewInstance(deepseek.Name, llm.WithDefaultModel("deepseek-chat"))
)

func init() {
	godotenv.Load()

	log.SetLogger(&log.WrapLogger{})
	llm.WithSingleConfig(map[string]any{
		"api_key": os.Getenv("deepseek_api_key"),
	})
}

func main() {
	// Create individual specialized agents
	researcher := agent.NewAgent(
		"Researcher",
		agent.WithDescription("Expert at finding information"),
		agent.WithTools(search.DuckDuckGo),
		agent.WithModel(model),
	)

	writer := agent.NewAgent(
		"Writer",
		agent.WithDescription("Expert at writing clear, engaging content"),
		agent.WithModel(model),
	)

	// Create a team with these agents
	contentTeam := agent.NewTeam(
		// "Content Team",
		"coordinate",
		agent.WithMembers(researcher, writer),
		agent.WithInstructions("You are a team of researchers and writers that work together to create high-quality content."),
		// agent.WithShowMembersResponses(true),
		agent.WithModel(model),
	)
	_ = contentTeam
	// Run the team with a task
	response := contentTeam.Invoke(
		context.Background(),
		"Create a short article about quantum computing",
		agent.WithStream(true),
	)
	if response.Error != nil {
		panic(response.Error)
	}
	fmt.Println("--------------------------------报告内容--------------------------------")
	for chunk := range response.Stream {
		fmt.Print(chunk.Choices[0].Delta.Content())
	}
	// fmt.Println(response.Answer)
	// contentTeam.DebugAssignTask()
	// xxx, err := tool.Spawn("assign_task_to_member").Invoke(context.Background(), map[string]any{
	// 	"member_name":     "Researcher",
	// 	"task":            `Find recent and reliable information about quantum computing, including its definition, applications, and current advancements.`,
	// 	"expected_output": `A summary of key points about quantum computing, including its definition, applications, and recent advancements.`,
	// })
	// if err != nil {
	// 	panic(err)
	// }
	// fmt.Println("xxx:", xxx)
}

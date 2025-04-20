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

var (
	model = llm.NewInstance(deepseek.Name, llm.WithDefaultModel("deepseek-chat"))
	// model = llm.NewInstance(openaic.Name, llm.WithDefaultModel("hunyuan"))
)

func init() {
	godotenv.Load()

	log.SetLogger(&log.WrapLogger{})
	// llm.WithSingleConfig(map[string]any{
	// 	"base_url": os.Getenv("hunyuan_base_url"),
	// 	"api_key":  os.Getenv("hunyuan_api_key"),
	// })
	llm.WithSingleConfig(map[string]any{
		"api_key": os.Getenv("deepseek_api_key"),
	})
}

func main() {

	agent1 := agent.NewAgent(
		"EnglishTranslator",
		agent.WithModel(model),
		agent.WithRole("You only answer in English."),
		agent.WithDescription("You are an English translator, please translate the text I give you into English."),
	)

	agent2 := agent.NewAgent(
		"ChineseTranslator",
		agent.WithModel(model),
		agent.WithRole("You only answer in Chinese."),
		agent.WithDescription("You are a Chinese translator, please translate the text I give you into Chinese."),
	)

	agent3 := agent.NewAgent(
		"FrenchTranslator",
		agent.WithModel(model),
		agent.WithRole("You only answer in French."),
		agent.WithDescription("You are a French translator, please translate the text I give you into French."),
	)

	team := agent.NewTeam(
		agent.TeamModeRoute,
		agent.WithLLM(model),
		agent.WithMembers([]*agent.Agent{agent1, agent2, agent3}),
		agent.WithName("Multi Language Translator Team"),
		agent.WithDescription("You are a language router that directs questions to the appropriate language agent."),
		agent.WithInstructions(
			"Identify the language of the user's question and direct it to the appropriate language agent.",
			"Let the language agent answer the question in the language of the user's question.",
			"If the user asks in a language whose agent is not a team member, respond in English with:",
			"'I can only answer in the following languages: English, Chinese, French. Please ask your question in one of these languages.'",
			"Always check the language of the user's input before routing to an agent.",
			"For unsupported languages like Italian, respond in English with the above message.",
		),
	)

	result, err := team.Run(context.Background(), "How are you?")
	if err != nil {
		panic(err)
		// log.ErrorContext(context.Background(), err)
	}
	fmt.Println(result)
}

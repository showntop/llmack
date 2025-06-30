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
	"github.com/showntop/llmack/storage"
)

var model = llm.New(deepseek.Name, llm.WithDefaultModel("deepseek-chat"))

func init() {
	godotenv.Load()

	log.SetLogger(&log.WrapLogger{})
	llm.SetSingleConfig(map[string]any{
		"api_key": os.Getenv("deepseek_api_key"),
	})

	if err := storage.SetupPostgresStorage(os.Getenv("postgres_dns")); err != nil {
		panic(err)
	}
}

func main() {

	stockSearcher := agent.NewAgent(
		"Stock Searcher",
		agent.WithModel(model),
		agent.WithRole("Searches the web for information on a stock."),
		agent.WithStorage(storage.NewPostgresStorageWithDNS(os.Getenv("postgres_dns"))),
	)

	webSearcher := agent.NewAgent(
		"Web Searcher",
		agent.WithModel(model),
		agent.WithRole("Searches the web for information on a company."),
		agent.WithStorage(storage.NewJSONStorage("tmp/web_searcher")),
	)

	team := agent.NewTeam(
		agent.TeamModeCoordinate,
		agent.WithName("Stock Team"),
		agent.WithModel(model),
		agent.WithInstructions(
			"You can search the stock market for information about a particular company's stock.",
			"You can also search the web for wider company information.",
		),
		agent.WithMembers(stockSearcher, webSearcher),
		agent.WithStorage(storage.NewJSONStorage("tmp/stock_team")),
	)
	response := team.Invoke(context.Background(), "What is the stock price of Apple?")
	if response.Error != nil {
		panic(response.Error)
	}
	for chunk := range response.Stream {
		fmt.Print(chunk.Choices[0].Delta.Content())
	}
}

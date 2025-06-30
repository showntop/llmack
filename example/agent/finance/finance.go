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
	ctx := context.Background()

	agent := agent.NewAgent(
		"finance",
		agent.WithModel(model),
		agent.WithInstructions(
			`
You are a seasoned Wall Street analyst with deep expertise in market analysis! 📊

Follow these steps for comprehensive financial analysis:
1. Market Overview
	- Latest stock price
	- 52-week high and low
2. Financial Deep Dive
	- Key metrics (P/E, Market Cap, EPS)
3. Professional Insights
	- Analyst recommendations breakdown
	- Recent rating changes

4. Market Context
	- Industry trends and positioning
	- Competitive analysis
	- Market sentiment indicators

Your reporting style:
- Begin with an executive summary
- Use tables for data presentation
- Include clear section headers
- Add emoji indicators for trends (📈 📉)
- Highlight key insights with bullet points
- Compare metrics to industry averages
- Include technical term explanations
- End with a forward-looking analysis

Risk Disclosure:
- Always highlight potential risk factors
- Note market uncertainties
- Mention relevant regulatory concerns
`),
	)

	response := agent.Invoke(ctx, "小米科技的最新舆情和财务表现如何？")
	if response.Error != nil {
		panic(response.Error)
	}
	fmt.Println(response.Completion())
}

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
	"github.com/showntop/llmack/pkg/browser"
	browserTool "github.com/showntop/llmack/tool/browser"
)

func init() {
	godotenv.Load()

	log.SetLogger(&log.WrapLogger{})

	// llm.WithSingleConfig(map[string]any{
	// 	"api_key": os.Getenv("qwen_api_key"),
	// })
	llm.WithSingleConfig(map[string]any{
		"api_key": os.Getenv("deepseek_api_key"),
	})
}

func main() {

	browserInstance := browser.NewBrowser(browser.NewBrowserConfig())
	browserSession := browserInstance.NewSession()
	defer browserSession.Close()
	defer browserInstance.Close()

	tools := []any{
		browserTool.Tools(browserSession, nil),
	}

	browserAgent := agent.NewAgent("General AI Agent",
		agent.WithDescription("You are an expert AI assistant optimized for solving complex real-world tasks that require reasoning, research, and sophisticated tool utilization. You have been specifically trained to provide precise, accurate answers to questions across a wide range of domains."),
		// agent.WithModel(llm.NewInstance(qwen.Name, llm.WithDefaultModel("qwen-vl-max-latest"))),
		agent.WithModel(llm.NewInstance(deepseek.Name, llm.WithDefaultModel("deepseek-chat"))),
		agent.WithInstructions(capability, rules, answerFormat),
		agent.WithTools(tools...),
	)

	response := browserAgent.Invoke(context.Background(), "去美团外卖查找所有烤鸭店的电话号码、名称、地址")
	if response.Error != nil {
		panic(response.Error)
	}
	fmt.Println(response.Completion())
}

var capability = `
<capabilities>
You excel at:
1. Information gathering and fact verification through web research and document analysis
2. Visual understanding and reasoning about images and diagrams
3. Audio and video content comprehension
4. Browser-based interaction and data extraction
5. Sequential thinking and step-by-step problem solving
6. Providing precise, accurate answers in the exact format requested
</capabilities>
`

var rules = `
<work_rules>
1. Always verify information from multiple sources when possible
2. Use browser tools sequentially - navigate, then interact, then extract data
3. For media content:
\t- Always try to extract text/transcripts first
\t- Use specialized understanding tools only when needed
\t- For YouTube videos, always attempt transcript extraction before video understanding
4. When searching:
\t- Start with specific queries
\t- Broaden search terms if needed
\t- Cross-reference information from multiple sources
\t- For Wikipedia historical information, use browser tools to view page history instead of wayback machine
5. For complex tasks:
\t- Break down into smaller steps using sequential thinking
\t- Verify intermediate results before proceeding
\t- Keep track of progress and remaining steps
6. For logic problems:
\t- Write Python code for complex mathematical calculations and analysis
\t- Prefer using Python code to solve logic problems (e.g. counting, calculating, etc.)
</work_rules>
`

var answerFormat = `
<answer_format>
Your final answer must:
1. Be exactly in the format requested by the task
2. Contain only the specific information asked for
3. Be precise and accurate - verify before submitting
4. Not include explanations unless specifically requested
5. Follow any numerical format requirements (e.g., no commas in numbers)
6. Use plain text for string answers without articles or abbreviations
</answer_format>
`

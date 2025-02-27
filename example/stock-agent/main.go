package main

import (
	"context"
	"fmt"
	"os"

	"github.com/joho/godotenv"

	"github.com/showntop/llmack/engine"
	"github.com/showntop/llmack/llm"
	"github.com/showntop/llmack/llm/qwen"
	"github.com/showntop/llmack/log"
	"github.com/showntop/llmack/tool"
	"github.com/showntop/llmack/tool/file"
	"github.com/showntop/llmack/tool/search"

	"github.com/showntop/llmack/example/stock-agent/prompt"
)

func init() {
	log.SetLogger(&log.WrapLogger{})

	godotenv.Load()
	llm.WithSingleConfig(map[string]any{
		"api_key": os.Getenv("qwen_api_key"),
		// "api_key":  os.Getenv("deepseek_api_key2"),
		// "base_url": "https://api.lkeap.cloud.tencent.com/v1",
	})

	tool.WithConfig(map[string]any{
		"searxng": map[string]any{
			"base_url": "http://9.134.217.159:8080",
		},
		"minimax": map[string]any{
			"api_key": os.Getenv("minmax_api_key"),
		},
		"siliconflow": map[string]any{
			"api_key": os.Getenv("siliconflow_api_key"),
		},
	})
}

func main() {
	settings := engine.DefaultSettings()
	settings.PresetPrompt = prompt.StockPrompt
	settings.LLMModel.Provider = qwen.Name
	// settings.LLMModel.Name = "deepseek-r1"
	settings.LLMModel.Name = "qwen-plus"
	settings.Agent.Mode = "ReAct"
	settings.Tools = append(settings.Tools, search.Searxng, file.WriteFile)
	eng := engine.NewAgentEngine(settings, engine.WithLogger(&log.WrapLogger{}))
	esm := eng.Execute(context.Background(), engine.Input{
		Inputs: map[string]any{
			"goals": []string{
				`Generate an overview of "Sahara", detailing its business model, history, and competitors. Note any special circumstances such as a high-profile CEO or past scandals.`,
				`Conduct a SWOT analysis for "Sahara", focusing on its Strengths, Weaknesses, Opportunities, and Threats.`,
				`Provide a definitive investment recommendation of BUY, HOLD, or SELL for "Sahara". This recommendation should be based on the totality of your research and analysis, including but not limited to the Company Overview and SWOT Analysis.`,
				`Compile a comprehensive report that integrates the Company Overview, SWOT Analysis, and Investment Recommendation. The report should be at least 1000 words and outputted as a single .txt file.`,
			},
			"instructions": []string{
				`Utilize your TOOLS to conduct research on "Sahara". Extract and store any information relevant for the completion of each independent GOAL.`,
				`The comprehensive report should integrate the Company Overview, SWOT Analysis, and your own commentary or analysis, which will contribute to the Investment Recommendation.`,
				`Use the Write File (or Append File) TOOL to compile and output the comprehensive report on "Sahara" as a single .txt file.`,
				`Use the Read File TOOL to confirm the existence and accessibility of the final report. Ensure that the report is at least 1000 words long before making it available to the user.`,
			},
			"constraints": []string{
				`If you are unsure how you previously did something or want to recall past events, thinking about similar events will help you remember.`,
				`Ensure the tool and args are as per current plan and reasoning`,
				`Exclusively use the tools listed under "TOOLS"`,
				`REMEMBER to format your response as JSON, using double quotes ("") around keys and string values, and commas (,) to separate items in arrays and objects. IMPORTANTLY, to use a JSON object as a string in another JSON object, you need to escape the double quotes.`,
			},
		},
	})
	for evt := esm.Next(); evt != nil; evt = esm.Next() {
		if evt.Error != nil {
			panic(evt.Error)
		}
		// fmt.Println("main event name:", evt.Name, ", data:", evt.Data)
		if cv, ok := evt.Data.(*llm.Chunk); ok {
			_ = cv
			fmt.Print(cv.Delta.Message.Content())
		} else {
			// fmt.Println("main event name:", evt.Name, ", data:", evt.Data)
		}
	}
}

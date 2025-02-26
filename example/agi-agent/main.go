package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"

	"github.com/showntop/llmack/engine"
	"github.com/showntop/llmack/llm"
	"github.com/showntop/llmack/llm/deepseek"
	"github.com/showntop/llmack/log"
	"github.com/showntop/llmack/tool"
	"github.com/showntop/llmack/tool/image"
	"github.com/showntop/llmack/tool/search"

	"github.com/showntop/llmack/example/agi-agent/prompt"
)

func init() {
	log.SetLogger(&log.WrapLogger{})

	godotenv.Load()
	llm.WithSingleConfig(map[string]any{
		"api_key":  os.Getenv("deepseek_api_key2"),
		"base_url": "https://api.lkeap.cloud.tencent.com/v1",
	})

	tool.WithConfig(map[string]any{
		"searxng": map[string]any{
			"api_key": "http://9.134.217.159:8080",
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
	// Goal Based Super AGI Agent
	settings := engine.DefaultSettings()
	settings.PresetPrompt = prompt.AGIPrompt
	settings.LLMModel.Provider = deepseek.Name
	settings.LLMModel.Name = "deepseek-r1"
	settings.Tools = append(settings.Tools, search.Searxng, image.SiliconflowImageGenerate)
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
			"tools": renderTools(
				search.Searxng,
				image.SiliconflowImageGenerate,
			),
		},
		// Query: "你好",
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

func renderTools(tools ...string) string {
	messageTools := make([]*llm.Tool, 0)
	for _, toolName := range tools {
		tool := tool.Spawn(toolName)
		messageTool := &llm.Tool{
			Type: "function",
			Function: &llm.FunctionDefinition{
				Name:        tool.Name,
				Description: tool.Description,
				Parameters: map[string]any{
					"type":       "object",
					"properties": map[string]any{},
					"required":   []string{},
				},
			},
		}

		for _, p := range tool.Parameters {
			properties := messageTool.Function.Parameters["properties"].(map[string]any)
			properties[p.Name] = map[string]any{
				"description": p.LLMDescrition,
				"type":        p.Type,
				"enum":        nil,
			}
			if p.Required {
				messageTool.Function.Parameters["required"] = append(messageTool.Function.Parameters["required"].([]string), p.Name)
			}
		}

		messageTools = append(messageTools, messageTool)
	}

	// 组装成 1. xxx 2. yyy 格式
	final := ""
	for i := 0; i < len(messageTools); i++ {
		final += strconv.Itoa(i+1) + ". "
		final += messageTools[i].Function.Name + ":" + messageTools[i].Function.Description
		rawArgs, _ := json.Marshal(messageTools[i].Function.Parameters)
		final += ", args json schema: " + string(rawArgs)
		final += "\n"
	}
	return final
}

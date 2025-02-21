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
	"github.com/showntop/llmack/workflow"
)

func init() {
	log.SetLogger(&log.WrapLogger{})

	godotenv.Load()
	llm.WithSingleConfig(map[string]any{
		// "api_key":  os.Getenv("deepseek_api_key2"),
		// "base_url": "https://api.lkeap.cloud.tencent.com/v1",
		"api_key": os.Getenv("qwen_api_key"),
	})

	tool.WithConfig(map[string]any{
		"serper": map[string]any{
			"api_key": os.Getenv("serper_api_key"),
		},
	})
}

func main() {
	settings := engine.DefaultSettings()
	// settings.PresetPrompt = "你是一个AI写作助手，使用网络信息写文章"
	// settings.LLMModel.Provider = deepseek.Name
	// settings.LLMModel.Provider = qwen.Name
	// settings.LLMModel.Name = "qwen-plus"
	// settings.LLMModel.Name = "deepseek-v3"
	// settings.Tools = append(settings.Tools, datetime.GetDate, search.Serper)
	settings.Workflow = workflow.NewWorkflow(1, "writing-agent2").Link(
		workflow.Node{ID: "start", Kind: workflow.NodeKindStart},
		workflow.Node{ID: "create_subject", Kind: workflow.NodeKindLLM,
			// Inputs: workflow.Parameters{ // 不填默认使用 scope ？
			// 	"query": workflow.Parameter{Type: "string", Value: "{{query}}"},
			// },
			Metadata: map[string]any{
				"provider": qwen.Name,
				"model":    "qwen-plus",
			},
		},
		workflow.Node{ID: "end", Kind: workflow.NodeKindEnd, Outputs: workflow.Parameters{
			"response": workflow.Parameter{Type: "any", Value: "{{create_subject.response}}"},
		}},
	)
	eng := engine.NewWorkflowEngine(settings, engine.WithLogger(&log.WrapLogger{}))
	esm := eng.Execute(context.Background(), engine.Input{
		// Query: "制作绘本，主题：月亮和星星",
		Query: "你是谁？",
	})
	for evt := esm.Next(); evt != nil; evt = esm.Next() {
		if evt.Error != nil {
			panic(evt.Error)
		}
		// fmt.Println("main event name:", evt.Name, ", data:", evt.Data)
		if cv, ok := evt.Data.(*llm.Chunk); ok {
			_ = cv
			fmt.Println("main chunk:", cv.Delta.Message)
		} else {
			// fmt.Println("main event name:", evt.Name, ", data:", evt.Data)
		}
	}
}

package main

import (
	"context"
	"fmt"
	"os"

	"github.com/joho/godotenv"

	"github.com/showntop/llmack/engine"
	"github.com/showntop/llmack/example/picture-book-agent/prompt"
	"github.com/showntop/llmack/llm"
	openaic "github.com/showntop/llmack/llm/openai-c"
	"github.com/showntop/llmack/log"
	"github.com/showntop/llmack/tool"
	"github.com/showntop/llmack/tool/file"
	"github.com/showntop/llmack/tool/image"
	"github.com/showntop/llmack/tool/user"
)

func init() {
	log.SetLogger(&log.WrapLogger{})

	godotenv.Load()
	llm.WithSingleConfig(map[string]any{
		// "api_key":  os.Getenv("deepseek_api_key2"),
		// "base_url": "https://api.lkeap.cloud.tencent.com/v1",
		// "api_key": os.Getenv("qwen_api_key"),
		"base_url": os.Getenv("hunyuan_base_url"),
		"api_key":  os.Getenv("hunyuan_api_key"),
	})

	tool.WithConfig(map[string]any{
		"serper": map[string]any{
			"api_key": os.Getenv("serper_api_key"),
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
	settings.PresetPrompt = prompt.CreatePictureBookPrompt
	// settings.LLMModel.Provider = deepseek.Name
	// settings.LLMModel.Name = "deepseek-v3"
	// settings.LLMModel.Provider = qwen.Name
	// settings.LLMModel.Name = "qwen-plus"
	// settings.LLMModel.Provider = hunyuan.Name
	settings.LLMModel.Provider = openaic.Name
	settings.LLMModel.Name = "hunyuan-large"
	settings.Agent.Mode = "ReAct"
	settings.Tools = append(settings.Tools, user.Inquery, image.SiliconflowImageGenerate, file.WriteFile)
	eng := engine.NewAgentEngine(settings, engine.WithLogger(&log.WrapLogger{}))
	esm := eng.Execute(context.Background(), engine.Input{
		Inputs: map[string]any{
			"topic": "小猫钓鱼",
		},
	})
	finalAnswer := ""
	for evt := esm.Next(); evt != nil; evt = esm.Next() {
		if evt.Error != nil {
			panic(evt.Error)
		}
		// fmt.Println("main event name:", evt.Name, ", data:", evt.Data)
		if cv, ok := evt.Data.(*llm.Chunk); ok {
			_ = cv
			fmt.Println("main chunk:", cv.Delta.Message)
			finalAnswer += cv.Delta.Message.Content()
		} else {
			// fmt.Println("main event name:", evt.Name, ", data:", evt.Data)
		}
	}

	fmt.Println(finalAnswer)
}

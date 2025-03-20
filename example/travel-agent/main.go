package main

import (
	"context"
	"fmt"
	"os"

	"github.com/joho/godotenv"

	"github.com/showntop/llmack/engine"
	"github.com/showntop/llmack/llm"
	openaic "github.com/showntop/llmack/llm/openai-c"
	"github.com/showntop/llmack/log"

	"github.com/showntop/llmack/tool/file"
	"github.com/showntop/llmack/tool/search"
	"github.com/showntop/llmack/tool/weather"
)

func init() {
	log.SetLogger(&log.WrapLogger{})

	godotenv.Load()
	llm.WithSingleConfig(map[string]any{
		// "api_key": os.Getenv("deepseek_api_key2"),
		// "base_url": "https://api.lkeap.cloud.tencent.com/v1",
		// "base_url": "https://dashscope.aliyuncs.com/compatible-mode/v1",
		"base_url": os.Getenv("hunyuan_base_url"),
		"api_key":  os.Getenv("hunyuan_api_key"),
		// "api_key":  os.Getenv("zhipu_api_key"),
	})
}

func main() {
	settings := engine.DefaultSettings()
	settings.PresetPrompt = prompt
	// settings.LLMModel.Provider = deepseek.Name
	settings.LLMModel.Provider = openaic.Name
	settings.LLMModel.Name = "hunyuan-large"
	// settings.LLMModel.Name = "deepseek-v3"
	// settings.LLMModel.Provider = zhipu.Name
	// settings.LLMModel.Name = "glm-4v-flash"
	settings.Tools = append(settings.Tools, weather.QueryWeather, search.DuckDuckGo, file.WriteFile)
	settings.Agent.Mode = "ReAct"
	eng := engine.NewAgentEngine(settings, engine.WithLogger(&log.WrapLogger{}))
	esm := eng.Execute(context.Background(), engine.Input{
		Inputs: map[string]any{
			"query": "我要去北京旅游，规划一下旅程。",
		},
	})
	for evt := esm.Next(); evt != nil; evt = esm.Next() {
		if evt.Error != nil {
			panic(evt.Error)
		}
		if cv, ok := evt.Data.(*llm.Chunk); ok {
			fmt.Println("main chunk:", cv.Delta.Message)
		} else {
			fmt.Println("main event name:", evt.Name, ", data:", evt.Data)
		}
	}
}

var prompt = `
[Goals]
提供一份2025年3月10日从北京到上海，3月21日返回的旅行计划。
制定一个我可以去的旅游景点、餐馆、酒吧等地方的行程，可以满足我预定的日期和价格。
在Notion上制作详细的行程页面。

[constraints]
如果你不确定你以前是怎么做的，或者想回忆过去的事情，想想类似的事情会帮助你记忆。
确保工具和参数符合当前的计划和推理。
仅使用列出的工具。
记住将您的回复格式为JSON，在键和字符串值周围使用双引号（""），并使用逗号（,）分隔数组和对象中的项。重要的是，要在另一个JSON对象中使用JSON对象作为字符串，需要转义双引号。
`

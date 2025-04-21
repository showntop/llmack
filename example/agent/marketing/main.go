package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/joho/godotenv"
	"github.com/showntop/llmack/agent"
	"github.com/showntop/llmack/example/agent/agents"
	"github.com/showntop/llmack/llm"
	"github.com/showntop/llmack/llm/deepseek"
	"github.com/showntop/llmack/log"
)

var (
	// 将日期按 6 天一个周期进行划分
	startDate, _ = time.Parse("2006-01-02", "2025-03-25")
	endDate, _   = time.Parse("2006-01-02", "2025-04-01")
)

var (
	MODEL_NAME_R1 = "ep-20250227113433-vv7hr"
	MODEL_NAME_V3 = "ep-20250227112432-tlpgl"
)

var (
	model = llm.NewInstance(deepseek.Name, llm.WithDefaultModel(MODEL_NAME_V3))
)

func init() {
	godotenv.Load()

	log.SetLogger(&log.WrapLogger{})
	// llm.WithSingleConfig(map[string]any{
	// 	"base_url": os.Getenv("hunyuan_base_url"),
	// 	"api_key":  os.Getenv("hunyuan_api_key"),
	// })
	llm.WithSingleConfig(map[string]any{
		"api_key":  os.Getenv("deepseek_api_key3"),
		"base_url": os.Getenv("deepseek_base_url3"),
	})
}

func main() {
	ctx := context.Background()
	// workflow
	// 1. 数据分析 Aagent one
	// response := agents.DataAnalyst(startDate, endDate).Invoke(ctx,
	// 	"分析当前账户的数据，识别优化空间。",
	// 	agent.WithStream(true),
	// )
	// if response.Error != nil {
	// 	panic(response.Error)
	// }
	// for chunk := range response.Stream {
	// 	fmt.Print(chunk.Choices[0].Delta.Content())
	// }

	// 2. 数据分析 B agent two
	// response := agents.DataAnalystB(ctx, startDate, endDate).Invoke(ctx,
	// 	"分析当前账户的数据，识别优化空间。",
	// 	agent.WithStream(true),
	// )
	// if response.Error != nil {
	// 	panic(response.Error)
	// }
	// for chunk := range response.Stream {
	// 	fmt.Print(chunk.Choices[0].Delta.Content())
	// }

	// 3. 数据分析 C a Team of agent
	response3 := agents.DataAnalystC(ctx, startDate, endDate).Invoke(ctx,
		"分析当前账户的数据，识别优化空间。",
		agent.WithStream(true),
	)
	if response3.Error != nil {
		panic(response3.Error)
	}
	for chunk := range response3.Stream {
		fmt.Print(chunk.Choices[0].Delta.Content())
	}
}

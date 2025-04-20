package main

import (
	"context"
	"fmt"

	"github.com/showntop/llmack/agent"
)

var dataAnalyst = agent.NewAgent(
	"DataAnalyst",
	agent.WithModel(model),
	agent.WithDescription("你是一个数据分析师，擅长从数据中挖掘有价值的信息。"),
	agent.WithInstructions(
		"使用中文进行回答。",
	),
	agent.WithTools(
		fetchMyAudienceData(nil, startDate, endDate),
		fetchMyCreativeData(nil, startDate, endDate),
		fetchMyRegionData(nil, startDate, endDate),
		fetchMyPlacementData(nil, startDate, endDate),

		fetchPeerAudienceData(nil, startDate, endDate),
		fetchPeerTopCreativeData(nil, startDate, endDate),
		fetchPeerRegionData(nil, startDate, endDate),
		fetchPeerPlacementData(nil, startDate, endDate),
	),
)

func InvokeDataAnalyst(ctx context.Context) {
	response := dataAnalyst.Invoke(ctx,
		"分析当前账户的数据，识别优化空间。\n\n提供详细的账户数据分析报告，包括潜在的优化建议",
		agent.WithStream(true),
	)
	if response.Error != nil {
		panic(response.Error)
	}
	for chunk := range response.Stream {
		fmt.Print(chunk.Choices[0].Delta.Content())
	}
	// fmt.Println(response.Completion())
}

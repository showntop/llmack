package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/showntop/llmack/agent"
	"github.com/showntop/llmack/tool/search"
)

func planAndGenerateReport() {
	startDate, _ := time.Parse("2006-01-02", "2025-03-25")
	endDate, _ := time.Parse("2006-01-02", "2025-04-01")

	dataAnalyst := agent.NewAgent(
		"DataAnalyst",
		agent.WithModel(model),
		agent.WithDescription("你是一个数据分析师，擅长从数据中挖掘有价值的信息。"),
		agent.WithInstructions(
			"",
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

	webSearcher := agent.NewAgent(
		"WebSearcher",
		agent.WithModel(model),
		agent.WithDescription("你是一个网络搜索专家，擅长从网络上搜索相关信息。"),
		agent.WithInstructions(
			"",
		),
		agent.WithTools(
			search.Serper,
		),
	)
	_ = webSearcher

	// reviewer
	reviewer := agent.NewAgent(
		"Reviewer",
		agent.WithModel(model),
		agent.WithDescription("你是一个专业的营销方案评估者，擅长评估营销方案。"),
		agent.WithInstructions(
			"结合用户输入和营销方案，给出下一步继续深入分析的方向。",
			`
可参考的评估思路：
	- 创意深入分析方向, 思路: 通过受众-版位-创意-地域组合交叉分析, TOP创意维度随时间衰减情况, TOP 创意类的消耗占比和高消耗低CVR的创意类做细致的对比分析
	- 受众深入分析方向, 思路: 人群分层, 受众-版位-地域关联分析, 受众维度随时间的衰减情况等
	- 竞争分析, 思路: 受众-版位-地域及其组合交叉维度, 和同类优质账户的对比情况, 从竞争方面入手发现机会
	- 不要局限在上文中提到的思路, 可以发散思维, 从其他角度进行分析
			`,
			// "如果需要数据支持, 请向数据分析师和网络搜索专家获取。",
			"如果需要数据支持, 请使用工具获取。",
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

	reportWriter := agent.NewTeam( // 引导数据挖掘 agent、网络搜索 agent...进行协作，帮助提供更全面资料，以完成分析报告
		agent.TeamModeCoordinate, // 协调
		agent.WithModel(model),
		agent.WithMembers(dataAnalyst, reviewer),
		agent.WithDescription("你是一个专业的营销方案撰写者，擅长撰写营销方案。"),
		agent.WithInstructions(
			"1. 根据用户输入深入分析其意图, 充分从各个渠道获取信息, 汇总并完善。",
			"2. 尊重数据撰写全面、客观、专业的营销方案。",
			"3. 如果需要数据支持, 可以直接使用工具获取。",
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

	result, err := reportWriter.Run(context.Background(), "账户优化空间")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(result)
}

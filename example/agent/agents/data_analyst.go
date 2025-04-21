package agents

import (
	"context"
	"fmt"
	"time"

	"github.com/showntop/llmack/agent"
	"github.com/showntop/llmack/llm"
	"github.com/showntop/llmack/llm/deepseek"
	"github.com/showntop/llmack/tool/search"
	"golang.org/x/sync/errgroup"
)

// DataAnalyst1 使用全量数据进行分析，输出一份详细的数据洞察报告
func DataAnalyst1(startDate, endDate time.Time) *agent.Agent {
	return agent.NewAgent(
		"DataAnalyst",
		agent.WithModel(llm.NewInstance(deepseek.Name, llm.WithDefaultModel(MODEL_NAME_R1))),
		agent.WithDescription("你是一个数据分析师，擅长从数据中挖掘有价值的信息。"),
		agent.WithInstructions(
			"结合用户问题，深入对比本账户及标杆账户的广告投放数据，分析本账户中存在的问题，给出一份详细的数据洞察报告。",
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
}

func InvokeDataAnalyst(ctx context.Context, startDate, endDate time.Time) {
	response := DataAnalyst1(startDate, endDate).Invoke(ctx,
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

// 先分析本账户问题、再分析标杆账户优势，然后对比给出综合报告
func DataAnalystB(ctx context.Context, startDate, endDate time.Time) *agent.Agent {
	a1 := agent.NewAgent(
		"DataAnalystB1",
		agent.WithModel(llm.NewInstance(deepseek.Name, llm.WithDefaultModel(MODEL_NAME_R1))),
		agent.WithDescription("你是一个数据分析师，擅长从数据中挖掘有价值的信息。"),
		agent.WithInstructions(
			"结合用户问题和本账户的广告投放数据，分析账户中存在的问题",
			"使用中文进行回答。",
		),
		agent.WithTools(
			fetchMyAudienceData(nil, startDate, endDate),
			fetchMyCreativeData(nil, startDate, endDate),
			fetchMyRegionData(nil, startDate, endDate),
			fetchMyPlacementData(nil, startDate, endDate),
		),
	)

	a2 := agent.NewAgent(
		"DataAnalystB2",
		agent.WithModel(llm.NewInstance(deepseek.Name, llm.WithDefaultModel(MODEL_NAME_R1))),
		agent.WithDescription("你是一个数据分析师，擅长从数据中挖掘有价值的信息。"),
		agent.WithInstructions(
			"结合用户问题和标杆账户的广告投放数据，分析标杆账户的优势",
			"使用中文进行回答。",
		),
		agent.WithTools(
			fetchPeerAudienceData(nil, startDate, endDate),
			fetchPeerTopCreativeData(nil, startDate, endDate),
			fetchPeerRegionData(nil, startDate, endDate),
			fetchPeerPlacementData(nil, startDate, endDate),
		),
	)

	var answer1 string
	var answer2 string
	eg := errgroup.Group{}
	eg.Go(func() error {
		response := a1.Invoke(ctx, "分析当前账户的数据，识别优化空间。")
		if response.Error != nil {
			return response.Error
		}
		answer1 = response.Completion()
		return nil
	})
	eg.Go(func() error {
		response := a2.Invoke(ctx, "分析当前账户的数据，识别优化空间。")
		if response.Error != nil {
			return response.Error
		}
		answer2 = response.Completion()
		return nil
	})
	if err := eg.Wait(); err != nil {
		panic(err)
	}
	// 合并答案
	answer := "本账户存在的问题：\n<my_account_problem>" + answer1 + "\n</my_account_problem>" + "标杆账户的优势：\n<peer_account_advantage>" + answer2 + "\n</peer_account_advantage>\n"
	a3 := agent.NewAgent(
		"DataAnalystB3",
		agent.WithModel(llm.NewInstance(deepseek.Name, llm.WithDefaultModel(MODEL_NAME_R1))),
		agent.WithDescription("你是一个数据分析师，擅长从数据中挖掘有价值的信息。"),
		agent.WithInstructions(
			"结合用户问题和本账户的广告投放数据及两份数据洞察（一份本账户一份标杆账户），分析账户中存在的问题，给出详细的数据洞察",
			"以下为两份数据洞察：\n"+answer+"\n\n",
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
	// 生成报告
	return a3
}

// A Team，包含数据分析师、网络搜索专家、营销方案评估者、营销方案撰写者
// 数据分析师：进行数据洞察 -> 网络搜索专家：搜索相关信息 -> 营销方案评估者：评估营销方案 -> 营销方案撰写者：撰写营销方案
func DataAnalystC(ctx context.Context, startDate, endDate time.Time) *agent.Team {
	webSearcher := agent.NewAgent(
		"WebSearcher",
		agent.WithModel(modelR1),
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
		agent.WithModel(modelR1),
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

	// writer
	writer := agent.NewAgent(
		"ReportWriter",
		agent.WithModel(modelV3),
		agent.WithDescription("你是一个专业的营销方案撰写者，擅长撰写高质量的营销方案。根据团队成员的数据分析结果，撰写营销方案。"),
		agent.WithInstructions(
			"1. 使用总分总的内容结构呈现报告。",
			"2. 数据脱敏：",
			"\t- 除年龄外, 不允许透露任何具体数字, 结合上下文选择处理方式。",
			"\t- 全模糊表达，消耗占比, CVR, CPA, CTR 可以根据上下文表述为: CVR 较为优秀/优于行业同类账户 或 CVR 有一定改进空间/与同类优秀账户比尚有差距等",
			"\t- 数字模糊表达: 如CTR为3% 可以改为 CTR < 5% 或 CTR > 1%, 请根据上下文选择合适的表述。",
			"\t- 对于0或0%要改为定型描述, 如: CVR为0% 改为CVR 远低于平均水平。",
			"\t- 不允许输出创意类id。",
			"3. 不要建议广告主减少投放、降低减少预算或降低出价。",
			"4. 不要给任何创意内容特征的总结。",
			"5. 不要输出预期收益之类的承诺或者预测。",
			"6. 帮助用户找到潜在机会, 优化投放效果, 最终提高消费和总收益(转化量)。",
			"7. 名词术语替换，行业标杆替换为同类优质账户。",
			"8. 你可以结合下面的准则整合更优的广告优化建议与方向",
			"\t- 核心人群CTR不理想: 可以针对人群特点(性别, 年龄, 学历) 定制素材, 或者进一步优化定向(如采用更加精细的人群包)",
			"\t- 核心地域CTR不理想: 可以针对地域特点(方言, 习惯, 消费水平)定制素材",
			"\t- CVR低: 如果CTR高, 突出落地页优化和素材和落地页的匹配程度, 反之可以建议进一步优化定向(如采用更加精细的人群包)",
			"\t- 版位特性: 结合版位的特点（社交, 视频等）给出更加细化的建议",
			"\t- 给建议的时候, 可以基于行业知识或对基于受众的认知(如:受教育程度越高, 年龄越大, 往往消费能力越强等等)",
			"\t- 创意衰减: 及时补充素材, 优质素材可以利用AIGC工具客隆和改造出来",
			"\t- TOP创意占比过高: 及时补充素材, 保证多样性, 加速无效素材淘汰",
			"\t- 核心人群CPA低于同类优质账户, 可以考虑提价获取更多流量",
			"\t- 优化素材时, 可以参考核心受众的特点(性别, 年龄, 学历) 定制素材",
			"9. 使用中文进行回答。",
		),
	)
	_ = writer
	marketingTeam := agent.NewTeam( // 引导数据挖掘 agent、网络搜索 agent、报告生成 agent 进行协作，帮助提供更全面资料，以完成分析报告
		agent.TeamModeCoordinate, // 协调
		agent.WithModel(modelV3),
		agent.WithAgenticSharedContext(true),
		agent.WithShareMemberInteractions(true),
		agent.WithMembers(DataAnalyst1(startDate, endDate), reviewer, writer),
		agent.WithDescription("你是一个专业的营销方案策划团队Leader，全面协调团队成员规划制定高质量的营销方案。"),
		agent.WithInstructions(
			"根据用户输入深入分析其意图, 充分协调团队各成员进行协作。",
			"关键职责：",
			"1. 全面协调团队各成员进行协作。",
			"2. 根据用户输入深入分析其意图, 给出更全面专业的任务描述。",
			"团队协调指南：",
			"- 团队成员各司其职, 但需要充分协作, 以完成更全面、更深入的分析。",
			"- 确保给每个成员传达清晰的任务和目标。",
			"- 你需要先让 DataAnalyst1 进行数据洞察，然后让 Reviewer 进行评估，参考Reviewer的反馈，让 DataAnalyst1 根据Reviewer的反馈进行数据洞察的调整。",
			"- 只有 Reviewer 评估OK 后, 你才能指派 ReportWriter 进行报告撰写。",
		),
	)

	return marketingTeam
}

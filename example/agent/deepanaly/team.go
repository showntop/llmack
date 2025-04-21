package main

import (
	"context"
	"fmt"
	"log"

	"github.com/showntop/llmack/agent"
	"github.com/showntop/llmack/tool/search"
)

func planAndGenerateReport() {

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
		"Writer",
		agent.WithModel(model),
		agent.WithDescription("你是一个专业的营销方案撰写者，擅长撰写高质量的营销方案。根据团队成员的数据分析结果，撰写营销方案。"),
		agent.WithInstructions(
			"1. 使用总分总的内容结构呈现报告。",
			"2. 数据脱敏：",
			"  - 除年龄外, 不允许透露任何具体数字, 结合上下文选择处理方式。",
			"  - 全模糊表达，消耗占比, CVR, CPA, CTR 可以根据上下文表述为:  CVR 较为优秀/优于行业同类账户 或 CVR 有一定改进空间/与同类优秀账户比尚有差距等",
			"  - 数字模糊表达: 如CTR为3% 可以改为 CTR < 5% 或 CTR > 1%, 请根据上下文选择合适的表述。",
			"  - 对于0或0%要改为定型描述, 如: CVR为0% 改为CVR 远低于平均水平。",
			"  - 不允许输出创意类id。",
			"3. 不要建议广告主减少投放、降低减少预算或降低出价。",
			"4. 不要给任何创意内容特征的总结。",
			"5. 不要给预期收益。",
			"6. 帮助用户找到潜在机会, 优化投放效果, 最终提高消费和总收益(转化量)。",
			"7. 名词术语替换，行业标杆替换为同类优质账户。",
			`8. 你可以结合下面的准则整合更优的广告优化建议与方向
  - 1. 核心人群CTR不理想: 可以针对人群特点(性别, 年龄, 学历) 定制素材, 或者进一步优化定向(如采用更加精细的人群包)
  - 2. 核心地域CTR不理想: 可以针对地域特点(方言, 习惯, 消费水平)定制素材
  - 3. CVR低: 如果CTR高, 突出落地页优化和素材和落地页的匹配程度, 反之可以建议进一步优化定向(如采用更加精细的人群包)
  - 4. 版位特性: 结合版位的特点（社交, 视频等）给出更加细化的建议
  - 5. 给建议的时候, 可以基于行业知识或对基于受众的认知(如:受教育程度越高, 年龄越大, 往往消费能力越强等等)
  - 6. 创意衰减: 及时补充素材, 优质素材可以利用AIGC工具客隆和改造出来
  - 7. TOP创意占比过高: 及时补充素材, 保证多样性, 加速无效素材淘汰
  - 8. 核心人群CPA低于同类优质账户, 可以考虑提价获取更多流量
  - 9. 优化素材时, 可以参考核心受众的特点(性别, 年龄, 学历) 定制素材`,
			"8. 使用中文进行回答。",
		),
	)
	_ = writer
	marketingTeam := agent.NewTeam( // 引导数据挖掘 agent、网络搜索 agent...进行协作，帮助提供更全面资料，以完成分析报告
		agent.TeamModeCoordinate, // 协调
		agent.WithModel(model),
		agent.WithMembers(dataAnalyst, reviewer),
		agent.WithDescription("你是一个专业的营销方案策划团队，擅长协调团队成员规划制定高质量的营销方案。"),
		agent.WithInstructions(
			// "1. 根据用户输入深入分析其意图, 充分从各个渠道获取信息, 汇总并完善。",
			"1. 根据用户输入深入分析其意图, 充分协调团队各成员进行协作。",
			"2. 尊重数据撰写全面、客观、专业的营销方案。",
			"3. 如果需要数据支持, 可以直接使用工具获取。",
			"4. 使用中文进行回答。",
			"5. 你需要总是把数据分析师的发给Reviewer，以便他可以进行深入的评估。并结合他的评估结果进行下一步的决策。",
			"6. 如果Reviewer的评估结果不理想, 你需要重新协调团队成员进行协作。",
			"7. 如果Reviewer的评估结果理想, 你需要把数据分析师的分析结果和Reviewer的评估结果结合, 撰写营销方案。营销方案的撰写要求如下：",
			"  1. 使用总分总的内容结构呈现报告。",
			"  2. 数据脱敏：",
			"    - 除年龄外, 不允许透露任何具体数字, 结合上下文选择处理方式。",
			"    - 全模糊表达，消耗占比, CVR, CPA, CTR 可以根据上下文表述为:  CVR 较为优秀/优于行业同类账户 或 CVR 有一定改进空间/与同类优秀账户比尚有差距等",
			"    - 数字模糊表达: 如CTR为3% 可以改为 CTR < 5% 或 CTR > 1%, 请根据上下文选择合适的表述。",
			"    - 对于0或0%要改为定型描述, 如: CVR为0% 改为CVR 远低于平均水平。",
			"    - 不允许输出创意类id。",
			"    - 不要建议广告主减少投放、降低减少预算或降低出价。",
			"  4. 不要给任何创意内容特征的总结。",
			"  5. 不要给预期收益。",
			"  6. 帮助用户找到潜在机会, 优化投放效果, 最终提高消费和总收益(转化量)。",
			"  7. 名词术语替换，行业标杆替换为同类优质账户。",
			`  8. 可以结合下面的准则整合更优的广告优化建议与方向
    - 1. 核心人群CTR不理想: 可以针对人群特点(性别, 年龄, 学历) 定制素材, 或者进一步优化定向(如采用更加精细的人群包)
    - 2. 核心地域CTR不理想: 可以针对地域特点(方言, 习惯, 消费水平)定制素材
    - 3. CVR低: 如果CTR高, 突出落地页优化和素材和落地页的匹配程度, 反之可以建议进一步优化定向(如采用更加精细的人群包)
    - 4. 版位特性: 结合版位的特点（社交, 视频等）给出更加细化的建议
    - 5. 给建议的时候, 可以基于行业知识或对基于受众的认知(如:受教育程度越高, 年龄越大, 往往消费能力越强等等)
    - 6. 创意衰减: 及时补充素材, 优质素材可以利用AIGC工具客隆和改造出来
    - 7. TOP创意占比过高: 及时补充素材, 保证多样性, 加速无效素材淘汰
    - 8. 核心人群CPA低于同类优质账户, 可以考虑提价获取更多流量
    - 9. 优化素材时, 可以参考核心受众的特点(性别, 年龄, 学历) 定制素材`,
			"8. 使用中文进行回答。",
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

	response := marketingTeam.Invoke(context.Background(), "账户优化空间", agent.WithStream(true))
	if response.Error != nil {
		log.Fatal(response.Error)
	}
	for ccc := range response.Stream {
		fmt.Print(ccc.Choices[0].Delta.Content())
	}
}

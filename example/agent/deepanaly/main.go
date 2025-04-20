package main

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"github.com/showntop/llmack/agent"
	"github.com/showntop/llmack/llm"
	"github.com/showntop/llmack/llm/deepseek"
	"github.com/showntop/llmack/log"
	"github.com/showntop/llmack/program"
	"github.com/showntop/llmack/tool"
	"golang.org/x/sync/errgroup"
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
	// model = llm.NewInstance(openaic.Name, llm.WithDefaultModel("hunyuan"))
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
	// InvokeDataAnalyst(context.Background())
	planAndGenerateReport()
}

func main4() {
	// 将日期按 6 天一个周期进行划分
	startDate, _ := time.Parse("2006-01-02", "2025-03-25")
	endDate, _ := time.Parse("2006-01-02", "2025-04-01")

	result := ComparisonPeriod(context.Background(), startDate, endDate)
	fmt.Println(result)
	// periods := make([][]time.Time, 0)
	// curDate := startDate
	// for curDate.Before(endDate) {
	// 	secondDate := curDate.AddDate(0, 0, 6)
	// 	if secondDate.After(endDate) {
	// 		secondDate = endDate
	// 	}
	// 	periods = append(periods, []time.Time{curDate, secondDate})
	// 	curDate = secondDate
	// }

	// result := Comparison(context.Background(), periods)
	// fmt.Println(result)
}

func main3() {
	startDate, _ := time.Parse("2006-01-02", "2025-03-01")
	endDate, _ := time.Parse("2006-01-02", "2025-03-07")

	result := AnalysePeriodPeer(context.Background(), startDate, endDate)
	fmt.Println(result)
}

func main2() {
	ctx := context.Background()
	// 广告投放需求分析师
	// webSearcher := agent.NewAgent(
	// 	"webSearcher",
	// 	agent.WithModel(model),
	// 	agent.WithRole("You are a web searcher, please search the web for the information I need."),
	// 	agent.WithDescription("You are a web searcher, please search the web for the information I need."),
	// )

	// 获取数据
	// fetchAdvertSummaries()

	// 将日期按 6 天一个周期进行划分
	startDate, _ := time.Parse("2006-01-02", "2025-03-01")
	endDate, _ := time.Parse("2006-01-02", "2025-04-01")

	periods := make([][]time.Time, 0)
	curDate := startDate
	for curDate.Before(endDate) {
		secondDate := curDate.AddDate(0, 0, 6)
		if secondDate.After(endDate) {
			secondDate = endDate
		}
		periods = append(periods, []time.Time{curDate, secondDate})
		curDate = secondDate
	}

	fmt.Println(periods) // 2025-03-01 2025-03-07

	summary := AnalysePeer(ctx, periods)
	fmt.Println(summary)

	return

	// 数据分析师(问题分析)
	dataAnalyst := agent.NewAgent(
		"DataAnalyst",
		agent.WithModel(model),
		agent.WithRole("广告投放数据分析师"),
		agent.WithDescription("你是一个资深的广告投放数据分析师，擅长分析广告投放数据，洞察其中的规律。"),
		agent.WithInstructions(
			"数据主要包括目标受众数据表现、版位数据表现、创意数据表现、地域数据表现",
			"你可以运用结构化思维，从多个维度分析数据",
			"注重数据驱动，使用相关指标支持决策",
			"注意数据之间的共同维度(日期, 浅层转化目标, 深层转化目标)，它们是关联分析的关键",
		),
		agent.WithTools(
			fetchMyAudienceData(nil, startDate, endDate),
			fetchMyPlacementData(nil, startDate, endDate),
			fetchMyCreativeData(nil, startDate, endDate),
			fetchMyRegionData(nil, startDate, endDate),
			// fetchPeerPlacementData(nil, "", ""),
			// fetchPeerCreativeData(nil, "", ""),
			// fetchPeerRegionData(nil, "", ""),
			// fetchPeerTopCreativeData(nil, "", ""),
		),
	)
	response1 := dataAnalyst.Invoke(ctx, "基于本账户下广告投放数据，深度分析其中存在的问题")
	if response1.Error != nil {
		panic(response1.Error)
	}
	for chunk := range response1.Stream {
		fmt.Print(chunk)
	}

	// reportWriter := agent.NewAgent(
	// 	"ReportWriter",
	// 	agent.WithModel(model),
	// 	agent.WithRole("You are a report writer, please write the report I give you."),
	// 	agent.WithDescription("You are a report writer, please write the report I give you."),
	// )

	// team := agent.NewTeam(
	// 	agent.TeamModeCoordinate,
	// 	agent.WithLLM(model),
	// 	agent.WithMembers([]*agent.Agent{reportWriter}),
	// 	agent.WithName("广告营销分析专家"),
	// 	agent.WithDescription("你是一个广告营销分析专家，精通广告投放策略、媒体购买、广告优化等核心技能，擅长分析市场趋势、竞争对手策略和用户行为。"),
	// 	agent.WithInstructions(
	// 		"仔细分析广告投放需求，提出深入的问题以充分理解目标",
	// 		"给出具体、可执行的广告优化建议和方案",
	// 		"注重ROI最大化，使用相关指标支持决策",
	// 	),
	// )

	// result, err := team.Run(ctx, "分析我的广告账户进一步的优化空间")
	// if err != nil {
	// 	panic(err)
	// 	// log.ErrorContext(ctx, err)
	// }
	// fmt.Println(result)
}

func Comparison(ctx context.Context, periods [][]time.Time) string {
	results := make([]string, 0)
	ch := make(chan []string, len(periods))
	wg := errgroup.Group{}
	for _, period := range periods {
		wg.Go(func() error {
			result := ComparisonPeriod(ctx, period[0], period[1]) // [) 左闭右开
			ch <- []string{period[0].Format("20060102"), period[1].Format("20060102"), result}
			return nil
		})
	}
	if err := wg.Wait(); err != nil {
		panic(err)
	}
	close(ch)
	for result := range ch {
		fragment := fmt.Sprintf("以下为日期（%s-%s）的报告：\n%s", result[0], result[1], result[2])
		results = append(results, fragment)
	}

	return strings.Join(results, "\n")
}

func ComparisonPeriod(ctx context.Context, startDate, endDate time.Time) string {
	// 获取本账户数据
	myAudienceData, err := tool.Spawn(fetchMyAudienceData(nil, startDate, endDate)).Invoke(ctx, map[string]any{
		"start_date": startDate.Format("20060102"),
		"end_date":   endDate.Format("20060102"),
	})
	if err != nil {
		panic(err)
	}
	myPlacementData, err := tool.Spawn(fetchMyPlacementData(nil, startDate, endDate)).Invoke(ctx, map[string]any{
		"start_date": startDate.Format("20060102"),
		"end_date":   endDate.Format("20060102"),
	})
	if err != nil {
		panic(err)
	}
	myCreativeData, err := tool.Spawn(fetchMyCreativeData(nil, startDate, endDate)).Invoke(ctx, map[string]any{
		"start_date": startDate.Format("20060102"),
		"end_date":   endDate.Format("20060102"),
	})
	if err != nil {
		panic(err)
	}
	myRegionData, err := tool.Spawn(fetchMyRegionData(nil, startDate, endDate)).Invoke(ctx, map[string]any{
		"start_date": startDate.Format("20060102"),
		"end_date":   endDate.Format("20060102"),
	})
	if err != nil {
		panic(err)
	}
	// 获取同行竞品数据
	peerPlacementData, err := tool.Spawn(fetchPeerPlacementData(nil, startDate, endDate)).Invoke(ctx, map[string]any{
		"start_date": startDate.Format("20060102"),
		"end_date":   endDate.Format("20060102"),
	})
	if err != nil {
		panic(err)
	}
	peerAudienceData, err := tool.Spawn(fetchPeerAudienceData(nil, startDate, endDate)).Invoke(ctx, map[string]any{
		"start_date": startDate.Format("20060102"),
		"end_date":   endDate.Format("20060102"),
	})
	if err != nil {
		panic(err)
	}
	peerRegionData, err := tool.Spawn(fetchPeerRegionData(nil, startDate, endDate)).Invoke(ctx, map[string]any{
		"start_date": startDate.Format("20060102"),
		"end_date":   endDate.Format("20060102"),
	})
	if err != nil {
		panic(err)
	}
	peerTopCreativeData, err := tool.Spawn(fetchPeerTopCreativeData(nil, startDate, endDate)).Invoke(ctx, map[string]any{
		"start_date": startDate.Format("20060102"),
		"end_date":   endDate.Format("20060102"),
	})
	if err != nil {
		panic(err)
	}

	result := program.Predictor(
		program.WithLLMInstance(model),
	).WithInstruction(
		`
<task>
根据本账户广告投放数据和同行标杆广告投放数据，分析本账户存在的问题。
</task>

<instructions>
1. 全面分析给出洞察和问题以及对应的数据佐证，不需要解决方案。
2. 数据之间的共同维度(日期, 浅层转化目标, 深层转化目标)是关联分析的关键
3. 不要解读创意内容
</instructions>

以下为本账户广告投放数据：
<my_data>
本账户受众数据：
{{my_audience_data}}

本账户版位数据：
{{my_placement_data}}

本账户创意数据：
{{my_creative_data}}

本账户地域数据：
{{my_region_data}}

以下为同行标杆账户广告投放数据：
<peer_data>
标杆账户受众数据：
{{peer_audience_data}}

标杆账户版位数据：
{{peer_placement_data}}

标杆账户地域数据：
{{peer_region_data}}

标杆账户TOP创意数据：
{{peer_top_creative_data}}
</peer_data>
		`,
	).Invoke(ctx, map[string]any{
		"my_audience_data":  myAudienceData,
		"my_placement_data": myPlacementData,
		"my_creative_data":  myCreativeData,
		"my_region_data":    myRegionData,

		"peer_audience_data":     peerAudienceData,
		"peer_placement_data":    peerPlacementData,
		"peer_region_data":       peerRegionData,
		"peer_top_creative_data": peerTopCreativeData,
	})

	if result.Error() != nil {
		panic(result.Error())
	}

	return result.Completion()
}

func AnalysePeer(ctx context.Context, periods [][]time.Time) string {
	results := make([]string, 0)
	ch := make(chan []string, len(periods))
	wg := errgroup.Group{}
	for _, period := range periods {
		wg.Go(func() error {
			result := AnalysePeriodPeer(ctx, period[0], period[1]) // [) 左闭右开
			ch <- []string{period[0].Format("20060102"), period[1].Format("20060102"), result}
			return nil
		})
	}
	if err := wg.Wait(); err != nil {
		panic(err)
	}
	close(ch)
	for result := range ch {
		fragment := fmt.Sprintf("以下为日期（%s-%s）的报告：\n%s", result[0], result[1], result[2])
		results = append(results, fragment)
	}
	summary := program.Predictor(
		program.WithLLMInstance(model),
	).WithInstruction(
		`
<task>
我会给你{{report_count}}份关于同行竞品的数据分析挖掘报告，你需要合并它们。
</task>

<instructions>
去除冲突的内容，保留一致的内容。
去除重复的内容。
</instructions>

<reports>
以下为同行竞品广告投放数据分析报告：
<reports>
{{reports}}
</reports>
		`,
	).Invoke(ctx, map[string]any{
		"report_count": len(results),
		"reports":      results,
	})

	return summary.Completion()
}

func AnalysePeriodPeer(ctx context.Context, startDate, endDate time.Time) string {
	// 数据分析师2(竞品分析)(单周期 6 天)
	dataAnalyst2 := agent.NewAgent(
		"DataAnalyst2",
		agent.WithModel(model),
		agent.WithRole("广告投放数据挖掘专家"),
		agent.WithDescription("你是一个资深的广告投放数据挖掘专家，擅长挖掘分析广告投放数据，洞察其中的规律。"),
		agent.WithInstructions(
			"请分析行业优秀标杆账户的广告投放数据，从中学习优秀的投放策略和优化方法",
			"数据日期为"+startDate.Format("20060102")+"-"+endDate.Format("20060102")+"，主要包括目标受众数据表现、版位数据表现、创意数据表现、地域数据表现",
			"你可以运用结构化思维，从多个维度分析数据",
			"注重数据驱动，使用相关指标支持决策",
			"注意数据之间的共同维度(日期, 浅层转化目标, 深层转化目标)，它们是关联分析的关键",
			"请分析行业优秀标杆账户的广告投放数据，从中学习优秀的投放策略和优化方法",
			``,
		),
		agent.WithTools(
			fetchPeerPlacementData(nil, startDate, endDate),
			fetchPeerAudienceData(nil, startDate, endDate),
			fetchPeerRegionData(nil, startDate, endDate),
			fetchPeerTopCreativeData(nil, startDate, endDate),
		),
	)
	response2 := dataAnalyst2.Invoke(ctx, "分析其投放较好的深层原因及其对应的详细数据佐证")
	if response2.Error != nil {
		panic(response2.Error)
	}

	return response2.Completion()
}

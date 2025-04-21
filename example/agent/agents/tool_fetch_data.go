package agents

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/showntop/llmack/tool"
)

var headers = map[string][2][]string{
	"region": {
		{"日期", "地域名称", "浅层转化目标", "深层转化目标", "浅层转化量", "深层转化量", "消耗", "曝光量", "点击量", "转化量", "cpm", "cpa", "ctr", "cvr", "浅层cvr", "深层cvr", "有消耗广告数量"},
		{"dt", "audience_location", "optmization_goal", "second_optmization_goal", "shallow_cv2", "deep_cv2", "cost2", "sv2", "ckv2", "cv2", "cpm", "cpa", "ctr", "cvr", "shallow_cvr", "deep_cvr", "ad_num"},
	},
	"placement": {
		{"日期", "版位", "浅层转化目标", "深层转化目标", "浅层转化量", "深层转化量", "消耗", "曝光量", "点击量", "转化量", "cpm", "cpa", "ctr", "cvr", "浅层cvr", "深层cvr", "有消耗广告数量"},
		{"dt", "placement_group", "optmization_goal", "second_optmization_goal", "shallow_cv2", "deep_cv2", "cost2", "sv2", "ckv2", "cv2", "cpm", "cpa", "ctr", "cvr", "shallow_cvr", "deep_cvr", "ad_num"},
	},
	"creative": {
		{"日期", "创意id", "浅层转化目标", "深层转化目标", "浅层转化量", "深层转化量", "消耗", "曝光量", "点击量", "转化量", "cpm", "cpa", "ctr", "cvr", "浅层cvr", "深层cvr", "有消耗广告数量"},
		{"dt", "fingerprint_id", "optmization_goal", "second_optmization_goal", "shallow_cv2", "deep_cv2", "cost2", "sv2", "ckv2", "cv2", "cpm", "cpa", "ctr", "cvr", "shallow_cvr", "deep_cvr", "ad_num"},
	},
	"audience": {
		{"日期", "性别", "年龄", "学历", "浅层转化目标", "深层转化目标", "浅层转化量", "深层转化量", "消耗", "曝光量", "点击量", "转化量", "cpm", "cpa", "ctr", "cvr", "浅层cvr", "深层cvr", "有消耗广告数量"},
		{"dt", "gender", "age", "education", "optmization_goal", "second_optmization_goal", "shallow_cv2", "deep_cv2", "sv2", "ckv2", "cv2", "cpm", "cpa", "ctr", "cvr", "shallow_cvr", "deep_cvr", "ad_num"},
	},
}

func buildMarkdownTable(kind string, rows []map[string]any) string {
	headers := headers[kind]
	table := "|"
	for _, header := range headers[0] {
		table += header + "|"
	}
	table += "\n"
	table += "|"
	for _, _ = range headers[0] {
		table += "-|"
	}
	table += "\n"
	for _, row := range rows {
		table += "|"
		for _, header := range headers[0] { // TODO: 需要根据headers[1]来确定
			value := ""
			if v, ok := row[header]; ok {
				value = fmt.Sprint(v)
			}
			table += value + "|"
		}
		table += "\n"
	}
	return table
}

func fetchMyAudienceData(accounts []string, startDate, endDate time.Time) string {
	tl := &tool.Tool{
		Name:        "fetchMyAudienceData",
		Description: "获取本账户的目标受众数据",
	}
	tl.Invokex = func(ctx context.Context, params map[string]any) (string, error) {
		content, err := os.ReadFile("example/agent/agents/data/my_audience_data.json")
		if err != nil {
			return "", err
		}
		var rows []map[string]any
		json.Unmarshal(content, &rows)
		newRows := make([]map[string]any, 0)
		for _, row := range rows {
			date := row["日期"].(string)
			dateTime, err := time.Parse("20060102", date)
			if err != nil {
				return "", err
			}
			if dateTime.After(startDate) && dateTime.Before(endDate) {
				newRows = append(newRows, row)
			}
		}
		text := fmt.Sprintf("本账户的受众数据如下：\n%s", buildMarkdownTable("audience", newRows))
		return text, nil
	}
	tool.Register(tl)
	return tl.Name
}

func fetchMyPlacementData(accounts []string, startDate, endDate time.Time) string {
	tl := &tool.Tool{
		Name:        "fetchMyPlacementData",
		Description: "获取本账户的版位数据",
	}
	tl.Invokex = func(ctx context.Context, params map[string]any) (string, error) {
		content, err := os.ReadFile("example/agent/agents/data/my_placement_data.json")
		if err != nil {
			return "", err
		}
		var rows []map[string]any
		json.Unmarshal(content, &rows)
		newRows := make([]map[string]any, 0)
		for _, row := range rows {
			date := row["日期"].(string)
			dateTime, err := time.Parse("20060102", date)
			if err != nil {
				return "", err
			}
			if dateTime.After(startDate) && dateTime.Before(endDate) {
				newRows = append(newRows, row)
			}
		}
		text := fmt.Sprintf("本账户的版位数据如下：\n%s", buildMarkdownTable("placement", newRows))
		return text, nil
	}
	tool.Register(tl)
	return tl.Name
}

func fetchMyCreativeData(accounts []string, startDate, endDate time.Time) string {

	tl := &tool.Tool{
		Name:        "fetchMyCreativeData",
		Description: "获取本账户的创意数据",
	}
	tl.Invokex = func(ctx context.Context, params map[string]any) (string, error) {
		content, err := os.ReadFile("example/agent/agents/data/my_top_creative_data.json")
		if err != nil {
			return "", err
		}
		var rows []map[string]any
		json.Unmarshal(content, &rows)
		newRows := make([]map[string]any, 0)
		for _, row := range rows {
			date := row["日期"].(string)
			dateTime, err := time.Parse("20060102", date)
			if err != nil {
				return "", err
			}
			if dateTime.After(startDate) && dateTime.Before(endDate) {
				newRows = append(newRows, row)
			}
		}
		text := fmt.Sprintf("本账户的创意数据如下：\n%s", buildMarkdownTable("creative", newRows))
		return text, nil
	}
	tool.Register(tl)
	return tl.Name
}

func fetchMyRegionData(accounts []string, startDate, endDate time.Time) string {

	tl := &tool.Tool{
		Name:        "fetchMyRegionData",
		Description: "获取本账户的地域数据",
	}
	tl.Invokex = func(ctx context.Context, params map[string]any) (string, error) {
		content, err := os.ReadFile("example/agent/agents/data/my_top_region_data.json")
		if err != nil {
			return "", err
		}
		var rows []map[string]any
		json.Unmarshal(content, &rows)
		newRows := make([]map[string]any, 0)
		for _, row := range rows {
			date := row["日期"].(string)
			dateTime, err := time.Parse("20060102", date)
			if err != nil {
				return "", err
			}
			if dateTime.After(startDate) && dateTime.Before(endDate) {
				newRows = append(newRows, row)
			}
		}
		text := fmt.Sprintf("本账户的地域数据如下：\n%s", buildMarkdownTable("region", newRows))
		return text, nil
	}
	tool.Register(tl)
	return tl.Name
}

func fetchPeerPlacementData(accounts []string, startDate time.Time, endDate time.Time) string {

	tl := &tool.Tool{
		Name:        "fetchPeerPlacementData",
		Description: "获取同行版位上的投放数据",
	}
	tl.Invokex = func(ctx context.Context, params map[string]any) (string, error) {
		content, err := os.ReadFile("example/agent/agents/data/peer_placement_data.json")
		if err != nil {
			return "", err
		}
		var rows []map[string]any
		json.Unmarshal(content, &rows)
		newRows := make([]map[string]any, 0)
		for _, row := range rows {
			date := row["日期"].(string)
			dateTime, err := time.Parse("20060102", date)
			if err != nil {
				return "", err
			}
			if dateTime.After(startDate) && dateTime.Before(endDate) {
				newRows = append(newRows, row)
			}
		}
		text := fmt.Sprintf("标杆账户版位上的投放数据如下：\n%s", buildMarkdownTable("placement", newRows))
		return text, nil
	}
	tool.Register(tl)
	return tl.Name
}

func fetchPeerAudienceData(accounts []string, startDate, endDate time.Time) string {

	tl := &tool.Tool{
		Name:        "fetchPeerAudienceData",
		Description: "获取标杆账户受众投放表现数据",
	}
	tl.Invokex = func(ctx context.Context, params map[string]any) (string, error) {
		content, err := os.ReadFile("example/agent/agents/data/peer_audience_data.json")
		if err != nil {
			return "", err
		}
		var rows []map[string]any
		json.Unmarshal(content, &rows)
		newRows := make([]map[string]any, 0)
		for _, row := range rows {
			date := row["日期"].(string)
			dateTime, err := time.Parse("20060102", date)
			if err != nil {
				return "", err
			}
			if dateTime.After(startDate) && dateTime.Before(endDate) {
				newRows = append(newRows, row)
			}
		}
		text := fmt.Sprintf("标杆账户受众投放表现数据如下：\n%s", buildMarkdownTable("audience", newRows))
		return text, nil
	}
	tool.Register(tl)
	return tl.Name
}

func fetchPeerRegionData(accounts []string, startDate, endDate time.Time) string {

	tl := &tool.Tool{
		Name:        "fetchPeerRegionData",
		Description: "获取同行地域上的投放数据",
	}
	tl.Invokex = func(ctx context.Context, params map[string]any) (string, error) {
		content, err := os.ReadFile("example/agent/agents/data/peer_top_region_data.json")
		if err != nil {
			return "", err
		}
		var rows []map[string]any
		json.Unmarshal(content, &rows)
		newRows := make([]map[string]any, 0)
		for _, row := range rows {
			date := row["日期"].(string)
			dateTime, err := time.Parse("20060102", date)
			if err != nil {
				return "", err
			}
			if dateTime.After(startDate) && dateTime.Before(endDate) {
				newRows = append(newRows, row)
			}
		}
		text := fmt.Sprintf("标杆账户地域上的投放数据如下：\n%s", buildMarkdownTable("region", newRows))
		return text, nil
	}
	tool.Register(tl)
	return tl.Name
}

func fetchPeerTopCreativeData(accounts []string, startDate, endDate time.Time) string {

	tl := &tool.Tool{
		Name:        "fetchPeerTopCreativeData",
		Description: "获取同行优质创意的投放数据",
	}
	tl.Invokex = func(ctx context.Context, params map[string]any) (string, error) {
		content, err := os.ReadFile("example/agent/agents/data/peer_top_creative_data.json")
		if err != nil {
			return "", err
		}
		var rows []map[string]any
		json.Unmarshal(content, &rows)
		newRows := make([]map[string]any, 0)
		for _, row := range rows {
			date := row["日期"].(string)
			dateTime, err := time.Parse("20060102", date)
			if err != nil {
				return "", err
			}
			if dateTime.After(startDate) && dateTime.Before(endDate) {
				newRows = append(newRows, row)
			}
		}
		text := fmt.Sprintf("标杆账户优质创意的投放数据如下：\n%s", buildMarkdownTable("creative", newRows))
		return text, nil
	}
	tool.Register(tl)
	return tl.Name
}

func fetchAdvertData(accounts []string, startDate, endDate time.Time) string {
	tl := &tool.Tool{
		Name:        "fetchAdvertData",
		Description: "获取广告投放数据",
		Parameters: []tool.Parameter{
			{
				Name:          "kind",
				LLMDescrition: "数据类型（包括受众表现、版位表现、创意表现、地域表现、转化表现、落地页表现）",
				Required:      true,
				Type:          tool.String,
				Options:       []string{"audience", "placement", "creative", "region"},
			},
		},
	}
	tl.Invokex = func(ctx context.Context, params map[string]any) (string, error) {
		kind := params["kind"].(string)
		switch kind {
		case "audience":
			content, err := os.ReadFile("my_audience_data.json")
			if err != nil {
				return "", err
			}
			return string(content), nil
		case "placement":
			content, err := os.ReadFile("my_placement_data.json")
			if err != nil {
				return "", err
			}
			return string(content), nil
		case "creative":
			content, err := os.ReadFile("my_top_creative_data.json")
			if err != nil {
				return "", err
			}
			return string(content), nil
		case "region":
			content, err := os.ReadFile("my_top_region_data.json")
			if err != nil {
				return "", err
			}
			return string(content), nil
		}
		return "nothing", nil
	}

	tool.Register(tl)

	return tl.Name
}

// ==================================================================================

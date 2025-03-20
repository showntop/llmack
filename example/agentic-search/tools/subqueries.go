package tools

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/showntop/llmack/tool"
)

const Subqueries = "Subqueries"

func init() {
	t := &tool.Tool{}
	t.Name = Subqueries
	t.Kind = "code"
	t.Description = "Subqueries 子问题/原子问题处理"
	t.Parameters = append(t.Parameters, tool.Parameter{
		Name:          "query",
		LLMDescrition: "original question of user query",
		Type:          tool.String,
		Required:      true,
	}, tool.Parameter{
		Name:          "subqueries",
		LLMDescrition: "subqueries of original query. a json string of []string",
		Type:          tool.String,
		Required:      true,
	})
	t.Invokex = func(ctx context.Context, args map[string]any) (string, error) {
		originalQuery, _ := args["query"].(string)
		_ = originalQuery
		raw, _ := args["subqueries"].(string)
		// fmt.Println(subqueries)
		var subqueries []string
		if err := json.Unmarshal([]byte(raw), &subqueries); err != nil {
			subqueries = strings.Split(raw, "\n")
		}

		for i := range subqueries {
			// 如何解决子问题之间的依赖关系
			query(originalQuery, subqueries[i])
		}
		return "", nil
	}
	tool.Register(t)
}

func query(originalQuery string, subquerie string) {
	// vector db if needed 自由知识库，这里示例就不展示，只展示公网内容。
	// 迭代解决问题，直到OK

}

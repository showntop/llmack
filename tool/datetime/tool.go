package datetime

import (
	"context"
	"time"

	"github.com/showntop/llmack/tool"
)

const GetDate = "GetDate"

func init() {
	t := &tool.Tool{}
	t.Name = "GetDate"
	t.Kind = "code"
	t.Description = "查询当前（今天）日期"
	t.Parameters = append(t.Parameters, tool.Parameter{})
	t.Invokex = func(ctx context.Context, args map[string]any) (string, error) {
		return time.Now().Format("2006-01-02"), nil
	}
	tool.Register(t)
}

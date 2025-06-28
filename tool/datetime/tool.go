package datetime

import (
	"context"
	"time"

	"github.com/showntop/llmack/tool"
)

const GetTime = "GetTime"

func init() {
	t := tool.New(
		tool.WithName(GetTime),
		tool.WithKind("code"),
		tool.WithDescription("获取当前时间"),
		tool.WithFunction(func(ctx context.Context, args string) (string, error) {
			return time.Now().Format("2006-01-02 15:04:05"), nil
		}),
	)
	tool.Register(t)
}

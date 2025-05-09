package mcp

import (
	"github.com/showntop/llmack/tool"
)

const Name = "mcp"

func init() {
	t := tool.New(
		tool.WithName(Name),
		tool.WithDescription("xx"),
	)
	tool.Register(t)
}

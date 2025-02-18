package file

import (
	"context"
	"os"

	"github.com/showntop/llmack/tool"
)

const CreateFile = "CreateFile"

func init() {
	t := &tool.Tool{}
	t.Name = CreateFile
	t.Kind = "code"
	t.Description = "创建文件"
	t.Parameters = append(t.Parameters, tool.Parameter{})
	t.Invokex = func(ctx context.Context, args map[string]any) (string, error) {
		_, err := os.Create(args["path"].(string))
		return "", err
	}
	tool.Register(t)
}

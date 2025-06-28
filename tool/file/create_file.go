package file

import (
	"context"
	"encoding/json"
	"os"

	"github.com/showntop/llmack/tool"
)

const CreateFile = "CreateFile"

func init() {
	t := tool.New(
		tool.WithName(CreateFile),
		tool.WithKind("code"),
		tool.WithDescription("创建文件"),
		tool.WithParameters(tool.Parameter{
			Name:          "path",
			LLMDescrition: "文件路径",
			Type:          tool.String,
			Required:      true,
		}),
		tool.WithFunction(func(ctx context.Context, args string) (string, error) {
			var params struct {
				Path string `json:"path"`
			}
			if err := json.Unmarshal([]byte(args), &params); err != nil {
				return "", err
			}
			_, err := os.Create(params.Path)
			if err != nil {
				return "", err
			}
			return "文件创建成功", nil
		}),
	)
	tool.Register(t)
}

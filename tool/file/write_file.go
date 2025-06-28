package file

import (
	"context"
	"encoding/json"
	"os"

	"github.com/showntop/llmack/tool"
)

const WriteFile = "WriteFile"

func init() {
	t := tool.New(
		tool.WithName(WriteFile),
		tool.WithKind("code"),
		tool.WithDescription("Writes text to a file"),
		tool.WithParameters(
			tool.Parameter{
				Name:          "file_name",
				LLMDescrition: "Name of the file to write. Only include the file name. Don't include path.",
				Type:          tool.String,
				Required:      true,
			},
			tool.Parameter{
				Name:          "content",
				LLMDescrition: "File content to write.",
				Type:          tool.String,
				Required:      true,
			},
		),
		tool.WithFunction(func(ctx context.Context, args string) (string, error) {
			var params struct {
				FileName string `json:"file_name"`
				Content  string `json:"content"`
			}
			if err := json.Unmarshal([]byte(args), &params); err != nil {
				return "", err
			}
			ff, err := os.Create(params.FileName)
			if err != nil {
				return "", err
			}
			defer ff.Close()
			_, err = ff.WriteString(params.Content)
			if err != nil {
				return "", err
			}
			return "文件写入成功", nil
		}),
	)
	tool.Register(t)
}

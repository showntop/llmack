package tools

import (
	"context"
	"os"

	"github.com/showntop/llmack/tool"
)

const Reflect = "Reflect"

func init() {
	t := &tool.Tool{}
	t.Name = Reflect
	t.Kind = "code"
	t.Description = "Writes text to a file"
	t.Parameters = append(t.Parameters, tool.Parameter{
		Name:          "file_name",
		LLMDescrition: "Name of the file to write. Only include the file name. Don't include path.",
		Type:          tool.String,
		Required:      true,
	}, tool.Parameter{
		Name:          "content",
		LLMDescrition: "File content to write.",
		Type:          tool.String,
		Required:      true,
	})
	t.Invokex = func(ctx context.Context, args map[string]any) (string, error) {
		fileName, _ := args["file_name"].(string)
		ff, err := os.Create(fileName)
		if err != nil {
			return "", err
		}
		_, err = ff.WriteString(args["content"].(string))
		return "", err
	}
	tool.Register(t)
}

var reflectPrompt = `
- Critically examine <question>, <context>, <knowledge>, <bad-attempts>, and <learned-strategy> to identify gaps and the problems. 
- Identify gaps and ask key clarifying questions that deeply related to the original question and lead to the answer
- Ensure each reflection:
 - Cuts to core emotional truths while staying anchored to original <question>
 - Transforms surface-level problems into deeper psychological insights
 - Makes the unconscious conscious
`

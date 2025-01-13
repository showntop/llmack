package program

import (
	"encoding/json"
	"fmt"
	"reflect"
	"regexp"
	"strings"

	"github.com/showntop/llmack/llm"
)

// OutAdapter ...
type OutAdapter interface {
	Format(p *predictor, inputs map[string]any, target any) ([]llm.Message, error)
	Parse(string, any) error
}

// MarkableAdapter ...
type MarkableAdapter struct {
}

// Format ...
func (ada *MarkableAdapter) Format(p *predictor, inputs map[string]any) ([]llm.Message, error) {
	sysPromptBuilder := strings.Builder{}
	sysPromptBuilder.WriteString("Your input fields are:\n")
	for name, field := range p.InputFields {
		sysPromptBuilder.WriteByte('-')
		sysPromptBuilder.WriteString("'" + name + "'")
		sysPromptBuilder.WriteString(": ")
		sysPromptBuilder.WriteString(field.Description)
		sysPromptBuilder.WriteByte('\n')
	}
	sysPromptBuilder.WriteByte('\n')

	sysPromptBuilder.WriteString("Your output fields are:\n")
	for _, out := range p.OutputFields {
		sysPromptBuilder.WriteByte('-')
		sysPromptBuilder.WriteString("'" + out.Name + "'")
		sysPromptBuilder.WriteString(": ")
		sysPromptBuilder.WriteString(out.Description)
		sysPromptBuilder.WriteByte('\n')
	}
	sysPromptBuilder.WriteByte('\n')

	sysPromptBuilder.WriteString("All interactions will be structured in the following way, with the appropriate values filled in.\n")
	sysPromptBuilder.WriteByte('\n')

	for name := range p.InputFields {
		sysPromptBuilder.WriteString("[[ ## " + name + " ## ]]")
		sysPromptBuilder.WriteByte('\n')
		sysPromptBuilder.WriteString("{" + name + "}")
		sysPromptBuilder.WriteByte('\n')
		sysPromptBuilder.WriteByte('\n')
	}

	for _, out := range p.OutputFields {
		sysPromptBuilder.WriteString("[[ ## " + out.Name + " ## ]]")
		sysPromptBuilder.WriteByte('\n')
		sysPromptBuilder.WriteString("{" + out.Name + "}")
		sysPromptBuilder.WriteByte('\n')
		sysPromptBuilder.WriteByte('\n')
	}

	sysPromptBuilder.WriteString("In adhering to this structure, your objective is: \n")
	sysPromptBuilder.WriteString(p.Instruction)
	sysPromptBuilder.WriteByte('\n')

	userPromptBuilder := strings.Builder{}
	for name, value := range inputs {
		userPromptBuilder.WriteString("[[ ## " + name + " ## ]]")
		userPromptBuilder.WriteByte('\n')
		userPromptBuilder.WriteString(value.(string))
		userPromptBuilder.WriteByte('\n')
		userPromptBuilder.WriteByte('\n')
	}
	userPromptBuilder.WriteString("Respond with the corresponding output fields, starting with the field ")
	i := 0
	for _, out := range p.OutputFields {
		if i == 0 {
			userPromptBuilder.WriteString("[[ ## " + out.Name + " ## ]]")
		} else if i == 1 {
			userPromptBuilder.WriteString(", then [[ ## " + out.Name + " ## ]]")
		} else {
			userPromptBuilder.WriteString(",[[ ## " + out.Name + " ## ]]")
		}
		i++
	}
	userPromptBuilder.WriteString(", and then ending with the marker for `[[ ## completed ## ]]`.")

	messages := []llm.Message{
		llm.SystemPromptMessage(sysPromptBuilder.String()),
		llm.UserPromptMessage(userPromptBuilder.String()),
	}
	return messages, nil
}

// Parse ...
func (ada *MarkableAdapter) Parse(completion string) map[string]any {
	fieldHeaderPattern := regexp.MustCompile(`\[\[ ## (\w+) ## \]\]`) // Replace with the actual regex pattern

	sections := make([][2]string, 1)
	for _, line := range strings.Split(completion, "\n") {
		match := fieldHeaderPattern.FindStringSubmatch(strings.TrimSpace(line))
		fmt.Println("nnooonnnn: ", line, match)
		if match != nil {
			sections = append(sections, [2]string{match[1], ""})
		} else {
			sections[len(sections)-1][1] += line + "\n"
		}
	}
	// 处理 object
	object := make(map[string]any)
	for i := range sections {
		object[sections[i][0]] = sections[i][1]
	}

	return object
}

// JSONAdapter ...
type JSONAdapter struct {
	target any
}

// Format ...
func (ada *JSONAdapter) Format(p *predictor, inputs map[string]any, target any) ([]llm.Message, error) {
	ada.target = target
	typ := reflect.TypeOf(ada.target)
	typ = typ.Elem()
	typeName := "object"
	if typ.Kind() == reflect.Slice {
		typeName = "array"
	}

	sysPromptBuilder := strings.Builder{}
	sysPromptBuilder.WriteString("Your input fields are:\n")
	for name, field := range p.InputFields {
		sysPromptBuilder.WriteString("- ")
		sysPromptBuilder.WriteString("'" + name + "'")
		sysPromptBuilder.WriteString(": ")
		sysPromptBuilder.WriteString(field.Description)
		sysPromptBuilder.WriteByte('\n')
	}
	sysPromptBuilder.WriteByte('\n')

	sysPromptBuilder.WriteString("Your output fields are:\n")
	for _, out := range p.OutputFields {
		sysPromptBuilder.WriteString("- ")
		sysPromptBuilder.WriteString("'" + out.Name + "'")
		if out.Type > 0 {
			sysPromptBuilder.WriteString("(" + out.Type.String() + ")")
		}
		sysPromptBuilder.WriteString("(如果内容值存在双引号，请使用转义符号转义，例如：\"你好？\"，他说到。)")
		sysPromptBuilder.WriteString(": ")
		sysPromptBuilder.WriteString(out.Description)
		sysPromptBuilder.WriteByte('\n')
	}
	sysPromptBuilder.WriteByte('\n')

	sysPromptBuilder.WriteString("All interactions will be structured in the following way, with the appropriate values filled in.\n")
	sysPromptBuilder.WriteByte('\n')

	sysPromptBuilder.WriteString("Inputs will have the following structure:\n")
	for name := range p.InputFields {
		sysPromptBuilder.WriteString("[[ ## " + name + " ## ]]")
		sysPromptBuilder.WriteByte('\n')
		sysPromptBuilder.WriteString("{" + name + "}")
		sysPromptBuilder.WriteByte('\n')
		sysPromptBuilder.WriteByte('\n')
	}
	sysPromptBuilder.WriteString("Outputs will be a JSON " + typeName + " with the following fields:\n")
	if typeName == "array" {
		sysPromptBuilder.WriteByte('[')
	}
	sysPromptBuilder.WriteByte('{')
	sysPromptBuilder.WriteByte('\n')
	i := 0
	for _, out := range p.OutputFields {
		if i > 0 {
			sysPromptBuilder.WriteByte(',')
			sysPromptBuilder.WriteByte('\n')
		}
		sysPromptBuilder.WriteString(`    "` + out.Name + `"`)
		sysPromptBuilder.WriteByte(':')
		sysPromptBuilder.WriteString(`"{` + out.Name + `}"`)
		if out.Type > 0 {
			sysPromptBuilder.WriteString("        # note the value you produce must be a single " + out.Type.String() + " value")
		}
		sysPromptBuilder.WriteString("        # 注意：如果内容值存在双引号，请使用转义符号转义，例如：\"你好？\"，他说到。")
		i++
	}
	sysPromptBuilder.WriteByte('\n')
	sysPromptBuilder.WriteByte('}')
	if typeName == "array" {
		sysPromptBuilder.WriteByte(']')
	}
	sysPromptBuilder.WriteByte('\n')
	sysPromptBuilder.WriteByte('\n')

	sysPromptBuilder.WriteString("In adhering to this structure, your objective is: \n")
	sysPromptBuilder.WriteString(p.Instruction)
	sysPromptBuilder.WriteByte('\n')

	userPromptBuilder := strings.Builder{}
	for name, value := range inputs {
		userPromptBuilder.WriteString("[[ ## " + name + " ## ]]")
		userPromptBuilder.WriteByte('\n')
		// TODO
		if x, ok := value.(string); ok {
			userPromptBuilder.WriteString(x)
		} else if xs, ok := value.([]string); ok {
			zzz := ""
			for i, x := range xs {
				zzz += fmt.Sprintf("[%d] «%s»\n", i, x)
			}
			userPromptBuilder.WriteString(zzz)
		}
		userPromptBuilder.WriteByte('\n')
		userPromptBuilder.WriteByte('\n')
	}
	userPromptBuilder.WriteString("Respond with a JSON " + typeName + " in the following order of fields:")
	i = 0
	for _, out := range p.OutputFields {
		if i == 0 {
			userPromptBuilder.WriteString("`" + out.Name + "`")
		} else if i == 1 {
			userPromptBuilder.WriteString(", then `" + out.Name + "`")
		} else {
			userPromptBuilder.WriteString(",`" + out.Name + "`")
		}
		i++
	}

	messages := []llm.Message{
		llm.SystemPromptMessage(sysPromptBuilder.String()),
		llm.UserPromptMessage(userPromptBuilder.String()),
	}
	return messages, nil
}

// Parse ...
func (ada *JSONAdapter) Parse(completion string, target any) error {
	completion = strings.TrimPrefix(completion, "```json")
	completion = strings.TrimSuffix(completion, "```")
	// json to target
	return json.Unmarshal([]byte(completion), &target)
}

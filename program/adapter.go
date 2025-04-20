package program

import (
	"encoding/json"
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"

	"github.com/showntop/llmack/llm"
	"github.com/showntop/llmack/prompt"
)

type InputAdapter interface {
	Format(p *predictor, inputs map[string]any, target any) ([]llm.Message, error)
}

type OutputAdapter interface {
	Parse(string, any) error
}

// Adapter ...
type Adapter interface {
	InputAdapter
	OutputAdapter
}

type MarkableAdapter struct {
	MarkableInputAdapter
	MarkableOutputAdapter
}

// MarkableInputAdapter ...
type MarkableInputAdapter struct {
}

// Format ...
func (ada *MarkableInputAdapter) Format(p *predictor, inputs map[string]any, _ any) ([]llm.Message, error) {
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
		llm.NewSystemMessage(sysPromptBuilder.String()),
		llm.NewUserTextMessage(userPromptBuilder.String()),
	}
	return messages, nil
}

type MarkableOutputAdapter struct {
}

// Parse ...
func (ada *MarkableOutputAdapter) Parse(completion string, object any) error {
	fieldHeaderPattern := regexp.MustCompile(`\[\[ ## (\w+) ## \]\]`) // Replace with the actual regex pattern

	sections := make([][2]string, 1)
	for _, line := range strings.Split(completion, "\n") {
		match := fieldHeaderPattern.FindStringSubmatch(strings.TrimSpace(line))
		if match != nil {
			sections = append(sections, [2]string{match[1], ""})
		} else {
			sections[len(sections)-1][1] += line + "\n"
		}
	}
	// 处理 object
	objectValue := reflect.ValueOf(object).Elem()
	for i := range sections {
		fieldType := objectValue.FieldByName(sections[i][0])
		if fieldType.Kind() == reflect.String {
			fieldType.SetString(sections[i][1])
		} else if fieldType.Kind() == reflect.Slice {
			fieldType.Set(reflect.MakeSlice(fieldType.Type(), 0, 0))
			fieldType.Set(reflect.Append(fieldType, reflect.ValueOf(sections[i][1])))
		} else if fieldType.Kind() == reflect.Map {
			fieldType.Set(reflect.MakeMap(fieldType.Type()))
		} else if fieldType.Kind() == reflect.Struct {
			return fmt.Errorf("unsupported field type: %s", fieldType.Kind())
		} else if fieldType.Kind() == reflect.Bool {
			fieldType.SetBool(sections[i][1] == "true")
		} else if fieldType.Kind() == reflect.Int {
			num, err := strconv.Atoi(sections[i][1])
			if err != nil {
				return fmt.Errorf("failed to convert %s to int: %w", sections[i][1], err)
			}
			fieldType.SetInt(int64(num))
		} else if fieldType.Kind() == reflect.Float64 {
			num, err := strconv.ParseFloat(sections[i][1], 64)
			if err != nil {
				return fmt.Errorf("failed to convert %s to float64: %w", sections[i][1], err)
			}
			fieldType.SetFloat(num)
		} else if fieldType.Kind() == reflect.Int64 {
			num, err := strconv.ParseInt(sections[i][1], 10, 64)
			if err != nil {
				return fmt.Errorf("failed to convert %s to int64: %w", sections[i][1], err)
			}
			fieldType.SetInt(num)
		} else if fieldType.Kind() == reflect.Uint64 {
			num, err := strconv.ParseUint(sections[i][1], 10, 64)
			if err != nil {
				return fmt.Errorf("failed to convert %s to uint64: %w", sections[i][1], err)
			}
			fieldType.SetUint(num)
		} else if fieldType.Kind() == reflect.Float32 {
			num, err := strconv.ParseFloat(sections[i][1], 32)
			if err != nil {
				return fmt.Errorf("failed to convert %s to float32: %w", sections[i][1], err)
			}
			fieldType.SetFloat(float64(num))
		} else {
			return fmt.Errorf("unsupported field type: %s", fieldType.Kind())
		}
	}

	return nil
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
		llm.NewSystemMessage(sysPromptBuilder.String()),
		llm.NewUserTextMessage(userPromptBuilder.String()),
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

// RawAdapter ...
type RawAdapter struct {
	RawInputAdapter
	RawOutputAdapter
}

type RawInputAdapter struct {
}

type RawOutputAdapter struct {
}

// Format ...
func (ada *RawInputAdapter) Format(p *predictor, inputs map[string]any, _ any) ([]llm.Message, error) {
	var err error
	p.Instruction, err = prompt.Render(p.Instruction, inputs)
	if err != nil {
		return nil, err
	}
	userPromptBuilder := strings.Builder{}
	userPromptBuilder.WriteString(p.Instruction)
	userPromptBuilder.WriteByte('\n')
	messages := []llm.Message{
		llm.NewSystemMessage(userPromptBuilder.String()),
	}
	return messages, nil
}

// Parse ...
func (ada *RawOutputAdapter) Parse(completion string, target any) error {
	// 给 target 赋值为 completion
	reflect.ValueOf(target).Elem().SetString(completion)
	return nil
}

type DefaultAdapter struct {
	RawInputAdapter
	MarkableOutputAdapter
}

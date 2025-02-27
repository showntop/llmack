package program

import (
	"context"
	"encoding/json"
	"strconv"
	"strings"

	"github.com/showntop/llmack/llm"
	"github.com/showntop/llmack/log"
	"github.com/showntop/llmack/tool"
)

type react struct {
	*predictor
}

type ReactResult struct {
	Tool *struct {
		Name string         `json:"name"`
		Args map[string]any `json:"args"`
	}
	Thoughts struct {
		Text    string `json:"text"`
		Reason  string `json:"reasoning"`
		Plan    any    `json:"plan"`
		Critism string `json:"criticism"`
		Speak   string `json:"speak"`
	}
}

// ReAct ...
func ReAct(opts ...option) *react {
	react := &react{}

	p := &predictor{
		adapter: &RawAdapter{},
		Promptx: Promptx{
			Name:         "ReActAgent",
			Instruction:  ReactPrompt,
			Description:  "ReAct mode Agent for General tasks Solve.",
			InputFields:  make(map[string]*Field),
			OutputFields: make(map[string]*Field),
		},
	}
	for i := 0; i < len(opts); i++ {
		opts[i](p)
	}
	if p.model == nil {
		p.model = defaultLLM
	}
	react.predictor = p
	return react
}

func (rp *react) WithTools(tools ...string) *react {
	messageTools := rp.renderTools(tools...)
	instruction := rp.Instruction
	toolString := ""
	for i := 0; i < len(messageTools); i++ {
		toolString += strconv.Itoa(i+1) + ". "
		toolString += messageTools[i].Function.Name + ":" + messageTools[i].Function.Description
		rawArgs, _ := json.Marshal(messageTools[i].Function.Parameters)
		toolString += ", args json schema: " + string(rawArgs)
		toolString += "\n"
	}
	instruction = strings.ReplaceAll(instruction, "{{tools}}", toolString)
	rp.predictor.Promptx.Instruction = instruction
	return rp
}

func (rp *react) WithInstruction(i string) *react {
	instruction := rp.Instruction
	instruction = strings.ReplaceAll(instruction, "{{instruction}}", i)
	rp.predictor.Promptx.Instruction = instruction
	return rp
}

// Invoke invoke forward for predicte
func (rp *react) Invoke(ctx context.Context, inputs map[string]any) *Result {
	var value Result
	value.p = rp.predictor

	thoughts := []string{}
	answer := ""
	for i := 0; i < 20; i++ {
		result, err := rp.invoke(ctx, inputs, thoughts)
		if err != nil {
			continue
		}
		if result.Tool != nil {
			// TODO check function name valid?
			toolResult, err := tool.Spawn(result.Tool.Name).Invoke(ctx, result.Tool.Args)
			log.InfoContextf(ctx, "AgentEngine invokeTool: %s, %v response: %s error: %v \n", result.Tool.Name, result.Tool.Args, "toolResult", err)
			if err != nil {
				continue
			}
			if toolResult != "" {
				thoughts = append(thoughts, toolResult)
			} else {
				thoughts = append(thoughts, "no result")
			}
		} else { // finish
			answer = result.Thoughts.Text
			break
		}
	}
	value.completion = answer
	return &value
}

func (rp *react) invoke(ctx context.Context, inputs map[string]any, thoughts []string) (*ReactResult, error) {
	var result ReactResult
	// inputs["thoughts"] = thoughts
	messages, err := rp.adapter.Format(rp.predictor, inputs, nil)
	if err != nil {
		return nil, err
	}
	if len(thoughts) > 0 {
		messages = append(messages, llm.AssistantPromptMessage(strings.Join(thoughts, "\n")))
		messages = append(messages, llm.AssistantPromptMessage("continue"))
	}
	response, err := rp.model.Invoke(ctx, messages,
		llm.WithStream(true),
	)
	if err != nil {
		return nil, err
	}
	rawResult := response.Result().Message.Content()
	if err := json.Unmarshal([]byte(rawResult), &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// renderTools ...
func (rp *react) renderTools(tools ...string) []*llm.Tool {
	messageTools := make([]*llm.Tool, 0)
	for _, toolName := range tools {
		tool := tool.Spawn(toolName)
		messageTool := &llm.Tool{
			Type: "function",
			Function: &llm.FunctionDefinition{
				Name:        tool.Name,
				Description: tool.Description,
				Parameters: map[string]any{
					"type":       "object",
					"properties": map[string]any{},
					"required":   []string{},
				},
			},
		}

		for _, p := range tool.Parameters {
			properties := messageTool.Function.Parameters["properties"].(map[string]any)
			properties[p.Name] = map[string]any{
				"description": p.LLMDescrition,
				"type":        p.Type,
				"enum":        nil,
			}
			if p.Required {
				messageTool.Function.Parameters["required"] = append(messageTool.Function.Parameters["required"].([]string), p.Name)
			}
		}

		messageTools = append(messageTools, messageTool)
	}
	return messageTools
}

var ReactPrompt = `
You are an AI assistant to solve complex problems. Your decisions must always be made independently without seeking user assistance.
Play to your strengths as an LLM and pursue simple strategies with no legal complications.

Respond to the human as helpfully and accurately as possible.

{{instruction}}

You have access to the following tools:

{{tools}}

PERFORMANCE EVALUATION:
1. Continuously review and analyze your actions to ensure you are performing to the best of your abilities.
2. Use instruction to decide the flow of execution and decide the next steps for achieving the task.
3. Constructively self-criticize your big-picture behavior constantly.
4. Reflect on past decisions and strategies to refine your approach.
5. Every tool has a cost, so be smart and efficient.

You have context following:
{{thoughts}}

Respond with only valid JSON conforming to the following schema:
{
    \"$schema\": \"http://json-schema.org/draft-07/schema#\",
    \"type\": \"object\",
    \"properties\": {
        \"thoughts\": {
            \"type\": \"object\",
            \"properties\": {
                \"text\": {
                    \"type\": \"string\",
                    \"description\": \"thought\"
                },
                \"reasoning\": {
                    \"type\": \"string\",
                    \"description\": \"short reasoning\"
                },
                \"plan\": {
                    \"type\": \"string\",
                    \"description\": \"- short bulleted\
                    - list that conveys\
- long-term plan\"
                },
                \"criticism\": {
                    \"type\": \"string\",
                    \"description\": \"constructive self-criticism\"
                },
                \"speak\": {
                    \"type\": \"string\",
                    \"description\": \"thoughts summary to say to user\"
                }
            },
            \"required\": [\"text\", \"reasoning\", \"plan\", \"criticism\", \"speak\"],
            \"additionalProperties\": false
        },
        \"tool\": {
            \"type\": \"object\",
            \"properties\": {
                \"name\": {
                    \"type\": \"string\",
                    \"description\": \"tool name\"
                },
                \"args\": {
                    \"type\": \"object\",
                    \"description\": \"tool arguments\"
                }
            },
            \"required\": [\"name\", \"args\"],
            \"additionalProperties\": false
        }
    },
    \"required\": [\"thoughts\", \"tool\"],
    \"additionalProperties\": false
}
`

package program

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/showntop/flatmap"
	"github.com/showntop/llmack/llm"
	"github.com/showntop/llmack/log"
	"github.com/showntop/llmack/prompt"
	"github.com/showntop/llmack/tool"
)

type react struct {
	*predictor
	userInstruction string
}

type ReactResult struct {
	Tool *struct {
		Name string `json:"name"`
		Args string `json:"args"`
	}
	Thoughts *struct {
		SelfReason string `json:"self_reason"`
		Text       string `json:"text"`
		Reason     string `json:"reasoning"`
		Plan       any    `json:"plan"`
		Critism    string `json:"criticism"`
		Speak      string `json:"speak"`
	} `json:"thoughts"`
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
	for i := range opts {
		opts[i](p)
	}
	if p.model == nil {
		p.model = defaultLLM
	}
	react.predictor = p
	return react
}

func (rp *react) WithActions(actions ...any) *react {
	var tools []string
	for _, action := range actions {
		if action, ok := action.(string); ok {
			tools = append(tools, action)
		}
	}
	messageTools := rp.renderTools(tools...)
	messageTools = append(messageTools, &llm.Tool{
		Type: "function",
		Function: &llm.FunctionDefinition{
			Name:        "finish",
			Description: "use this to signal that you have finished all your objectives.",
			Parameters: map[string]any{
				"response": "final response to let people know you have finished your objectives",
				"continue": "finish task or output to stdout",
			},
		},
	})
	instruction := rp.Instruction
	toolString := ""
	for i := range messageTools {
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
	rp.userInstruction = i
	instruction := rp.Instruction
	instruction = strings.ReplaceAll(instruction, "{{instruction}}", i)
	rp.predictor.Promptx.Instruction = instruction
	return rp
}

// Invoke invoke forward for predicte
func (rp *react) Invoke(ctx context.Context, messages []llm.Message, query string, inputs map[string]any) *Response {
	var value Response = Response{p: rp.predictor, stream: make(chan *llm.Chunk, 10000)}
	value.p = rp.predictor

	thoughts := []map[string]any{}
	for i := 0; i < 50; i++ {
		result, err := rp.invoke(ctx, messages, query, inputs, thoughts)
		if err != nil {
			continue
		}
		value.stream <- llm.NewChunk(0, llm.NewAssistantMessage(result.Thoughts.Speak), nil)
		if result.Tool != nil {
			if result.Tool.Name == "finish" {
				thoughts = append(thoughts, map[string]any{
					"thought":     result.Thoughts.Text,
					"tool_name":   result.Tool.Name,
					"tool_args":   result.Tool.Args,
					"tool_result": "finished",
				})
				break
			} else {
				// TODO check function name valid?
				toolResult, err := tool.Spawn(result.Tool.Name).Invoke(ctx, result.Tool.Args)
				log.InfoContextf(ctx, "react agent invoke tool: %s, %v response: %s error: %v \n", result.Tool.Name, result.Tool.Args, toolResult, err)
				if err != nil {
					continue
				}
				thoughts = append(thoughts, map[string]any{
					"thought":     result.Thoughts,
					"tool_name":   result.Tool.Name,
					"tool_args":   result.Tool.Args,
					"tool_result": toolResult,
				})
			}
		}
	}

	iii, _ := prompt.Render(rp.userInstruction, inputs)
	messages = append(messages, llm.NewUserTextMessage(iii))
	messages = append(messages, llm.NewAssistantMessage(rp.renderThoughts(thoughts)))
	messages = append(messages, llm.NewUserTextMessage("continue"))
	response, err := rp.model.Invoke(ctx, messages,
		llm.WithStream(true),
	)
	if err != nil {
		return &value
	}
	stream := response.Stream()
	var answer string
	for chunk := stream.Take(); chunk != nil; chunk = stream.Take() {
		value.stream <- chunk
		answer += chunk.Choices[0].Delta.Content()
	}
	value.completion = answer
	close(value.stream)
	return &value
}

func (rp *react) invoke(ctx context.Context, messages []llm.Message, query string, inputs map[string]any, thoughts []map[string]any) (*ReactResult, error) {
	var result ReactResult
	messages, err := rp.adapter.Format(rp.predictor, inputs, nil)
	if err != nil {
		return nil, err
	}
	if len(thoughts) > 0 {
		messages = append(messages, llm.NewAssistantMessage(rp.renderThoughts(thoughts)))
		messages = append(messages, llm.NewUserTextMessage("Determine next step to do(tool use or output), and respond using the format specified in above."))
	}
	response, err := rp.model.Invoke(ctx, messages,
		llm.WithStream(true),
	)
	if err != nil {
		return nil, err
	}
	rawResult := response.Result().Message.Content()
	rawResult = strings.TrimLeft(rawResult, "```json")
	rawResult = strings.TrimLeft(rawResult, "```")
	rawResult = strings.TrimRight(rawResult, "```")
	rawResult = strings.ReplaceAll(rawResult, "\n", "")
	if err := json.Unmarshal([]byte(rawResult), &result); err != nil {
		log.WarnContextf(ctx, "react agent invoke response: %s error: %v \n", rawResult, err)
		return nil, err
	}
	if result.Thoughts == nil {
		log.WarnContextf(ctx, "react agent invoke response: %s error: %v \n", rawResult, err)
		return nil, fmt.Errorf("no result")
	}
	log.InfoContextf(ctx, "react agent invoke response: %s error: %v \n", rawResult, err)
	return &result, nil
}

func (rp *react) renderThoughts(thoughts []map[string]any) string {
	thoughtsText := "# this is your trajectory: \n"
	for i, t := range thoughts {
		is := strconv.Itoa(i)
		thoughtsText += "\n[ thought " + is + "]\n"
		fmap, err := flatmap.Flatten(t, flatmap.DefaultTokenizer)
		if err != nil {
			raw, _ := json.Marshal(&t)
			thoughtsText += string(raw) + "\n"
			continue
		}
		fmap.Each(func(k string, v interface{}) {
			thoughtsText += k + ": " + fmt.Sprintf("%+v", v) + "\n"
		})
	}
	return thoughtsText
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
				Parameters:  tool.Parameters(),
			},
		}

		messageTools = append(messageTools, messageTool)
	}
	return messageTools
}

var ReactPrompt = `
You are an AI assistant to solve complex problems. Your decisions must always be made independently without seeking user assistance.
Play to your strengths as an LLM and pursue simple strategies with no legal complications.
If you have completed all your tasks or reached end state, make sure to use the "finish" tool.

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

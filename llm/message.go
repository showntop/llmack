package llm

import (
	"encoding/json"
	"fmt"
)

type PromptMessageRole string

const (
	PromptMessageRoleSystem    PromptMessageRole = "system"
	PromptMessageRoleUser      PromptMessageRole = "user"
	PromptMessageRoleAssistant PromptMessageRole = "assistant"
	PromptMessageRoleTool      PromptMessageRole = "tool"
)

type MultipartContent struct {
	Type string
	Data string
}

func MultipartContentImageURL(url string) *MultipartContent {
	return &MultipartContent{
		Type: "image_url",
		Data: url,
	}
}

func MultipartContentText(text string) *MultipartContent {
	return &MultipartContent{
		Type: "text",
		Data: text,
	}
}

// PromptMessage ...
type PromptMessage struct {
	role             PromptMessageRole
	content          string              // string | []*PromptMessageContent
	multiPartContent []*MultipartContent // string | []*PromptMessageContent
	Name             string
}

func (m PromptMessage) Content() string {
	return m.content
}

func (m PromptMessage) MultipartContent() []*MultipartContent {
	return m.multiPartContent
}

func (m PromptMessage) String() string {
	panic("implement it for string")
}

func (m PromptMessage) GetToolCalls() []*ToolCall {
	return nil
}

// MarshalJSON 实现marshal
func (m PromptMessage) MarshalJSON() ([]byte, error) {
	if m.role == PromptMessageRoleAssistant {
		var xxx = map[string]any{"role": m.role, "content": m.content, "name": m.Name}
		return json.Marshal(xxx)
	} else if m.role == PromptMessageRoleUser {
		panic(fmt.Sprintf("user prompt message should not be marshal to json: %v", m))
	}

	panic("implement it")
	return nil, nil
}

// Role ...
func (m PromptMessage) Role() PromptMessageRole {
	return m.role
}

// ToolID ...
func (m PromptMessage) ToolID() string {
	return ""
}

// SystemPromptMessage ...
func SystemPromptMessage(text string) PromptMessage {
	return PromptMessage{
		role:    PromptMessageRoleSystem,
		content: text,
	}
}

// UserTextPromptMessage ...
func UserTextPromptMessage(text string) PromptMessage {
	return PromptMessage{
		role:    PromptMessageRoleUser,
		content: text,
	}
}

// UserMultipartPromptMessage ...
func UserMultipartPromptMessage(contents ...*MultipartContent) PromptMessage {
	return PromptMessage{
		role:             PromptMessageRoleUser,
		multiPartContent: contents,
	}
}

// AssistantPromptMessage ...
func AssistantPromptMessage(text string) *assistantPromptMessage {
	m := &assistantPromptMessage{PromptMessage: PromptMessage{}}
	m.content = text
	m.role = PromptMessageRoleAssistant
	return m
}

// AssistantReasoningMessage ...
func AssistantReasoningMessage(text string) *assistantPromptMessage {
	m := &assistantPromptMessage{ReasoningContent: text}
	m.role = PromptMessageRoleAssistant
	return m
}

func (m *assistantPromptMessage) WithToolCalls(toolCalls []*ToolCall) *assistantPromptMessage {
	m.ToolCalls = toolCalls
	return m
}

func (m *assistantPromptMessage) WithReasoningContent(content string) *assistantPromptMessage {
	m.ReasoningContent = content
	return m
}

// ToolPromptMessage ...
func ToolPromptMessage(text string, toolID string) *toolPromptMessage {
	return &toolPromptMessage{
		PromptMessage: PromptMessage{
			role:    PromptMessageRoleTool,
			content: text,
			Name:    toolID,
		},
		toolID: toolID,
	}
}

// assistantPromptMessage ...
type assistantPromptMessage struct {
	PromptMessage
	ToolCalls        []*ToolCall `json:"tool_calls"`
	ReasoningContent string      `json:"reasoning_content"`
}

func (m assistantPromptMessage) GetToolCalls() []*ToolCall {
	return m.ToolCalls
}

func (m *assistantPromptMessage) String() string {
	if m.ReasoningContent != "" {
		return fmt.Sprintf("%s: reasoning: %s", m.role, m.ReasoningContent)
	}
	raw, _ := json.Marshal(m.ToolCalls)
	return fmt.Sprintf("%s => (content:%s tool_calls: %+v)", m.role, m.content, string(raw))
}

// toolPromptMessage ...
type toolPromptMessage struct {
	PromptMessage
	toolID string
}

// ToolID ...
func (m toolPromptMessage) ToolID() string {
	return m.toolID
}

// ToolCall ...
type ToolCall struct {
	ID       string           `json:"id"`
	Type     string           `json:"type"`
	Index    int              `json:"index"`
	Function ToolCallFunction `json:"function"`
}

// ToolCallFunction ...
type ToolCallFunction struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

// Messages ...
type Messages[T PromptMessage | assistantPromptMessage] []T

// Message ...
type Message interface {
	Content() string
	MultipartContent() []*MultipartContent
	Role() PromptMessageRole
	ToolID() string
	GetToolCalls() []*ToolCall
	String() string
}

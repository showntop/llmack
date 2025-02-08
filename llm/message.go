package llm

import (
	"fmt"
)

type PromptMessageRole string

const (
	PromptMessageRoleSystem    PromptMessageRole = "system"
	PromptMessageRoleUser      PromptMessageRole = "user"
	PromptMessageRoleAssistant PromptMessageRole = "assistant"
	PromptMessageRoleTool      PromptMessageRole = "tool"
)

// PromptMessageContent ...
type PromptMessageContent struct {
	Type int    `json:"type"`
	Data string `json:"data"`
}

// PromptMessage ...
type PromptMessage struct {
	role    PromptMessageRole
	content *PromptMessageContent
	Name    string
}

// MarshalJSON 实现marshal
func (m PromptMessage) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`{"role":"%s","content":"%s"}`, m.role, m.content.Data)), nil
}

// UnmarshalJSON 实现unmarshal
func (m *PromptMessage) UnmarshalJSON(data []byte) error {
	// return json.Unmarshal(data, m)
	return nil
}

func (m PromptMessage) String() string {
	if m.content == nil {
		return fmt.Sprintf("%s: %s", m.role, "")
	}
	return fmt.Sprintf("%s: %s", m.role, m.content.Data)
}

// Content ...
func (m PromptMessage) Content() *PromptMessageContent {
	return m.content
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
		content: TextPromptMessageContent(text),
	}
}

// UserPromptMessage ...
func UserPromptMessage(text string) PromptMessage {
	return PromptMessage{
		role:    PromptMessageRoleUser,
		content: TextPromptMessageContent(text),
	}
}

// AssistantPromptMessage ...
func AssistantPromptMessage(text string) *assistantPromptMessage {
	m := &assistantPromptMessage{PromptMessage: PromptMessage{}}
	m.content = TextPromptMessageContent(text)
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
			content: TextPromptMessageContent(text),
			Name:    toolID,
		},
		toolID: toolID,
	}
}

// TextPromptMessageContent ...
func TextPromptMessageContent(text string) *PromptMessageContent {
	return &PromptMessageContent{Data: text}
}

// assistantPromptMessage ...
type assistantPromptMessage struct {
	PromptMessage
	ToolCalls        []*ToolCall `json:"tool_calls"`
	ReasoningContent string      `json:"reasoning_content"`
}

func (m *assistantPromptMessage) String() string {
	if m.ReasoningContent != "" {
		return fmt.Sprintf("%s: reasoning: %s", m.role, m.ReasoningContent)
	}
	if m.content == nil {
		return fmt.Sprintf("%s: %s", m.role, "")
	}
	return fmt.Sprintf("%s: %s", m.role, m.content.Data)
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
	Function ToolCallFunction `json:"function"`
}

// ToolCallFunction ...
type ToolCallFunction struct {
	Name      string `json:"name"`
	Arguments string `json:"args"`
}

// Messages ...
type Messages[T PromptMessage | assistantPromptMessage] []T

// Message ...
type Message interface {
	Content() *PromptMessageContent
	Role() PromptMessageRole
	ToolID() string
	String() string
}

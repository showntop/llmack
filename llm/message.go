package llm

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
)

type MessageRole string

const (
	MessageRoleSystem    MessageRole = "system"
	MessageRoleUser      MessageRole = "user"
	MessageRoleAssistant MessageRole = "assistant"
	MessageRoleTool      MessageRole = "tool"
)

type MultipartContent struct {
	Type string
	Data any
}

func MultipartContentImageURL(url string) *MultipartContent {
	return &MultipartContent{
		Type: "image_url",
		Data: map[string]any{
			"url": url,
		},
	}
}

func MultipartContentCustom(typ string, content string) *MultipartContent {
	return &MultipartContent{
		Type: typ,
		Data: content,
	}
}

func MultipartContentImageBase64(format string, data []byte) *MultipartContent {
	if len(data) == 0 {
		return &MultipartContent{
			Type: "image_url",
			Data: map[string]any{
				"url": nil,
			},
		}
	}
	header := fmt.Sprintf("data:image/%s;base64,", format)

	return &MultipartContent{
		Type: "image_url",
		Data: map[string]any{
			"url": header + base64.StdEncoding.EncodeToString(data),
		},
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
	role             MessageRole
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
	if m.role == MessageRoleAssistant {
		var xxx = map[string]any{"role": m.role, "content": m.content, "name": m.Name}
		return json.Marshal(xxx)
	} else if m.role == MessageRoleUser {
		panic(fmt.Sprintf("user prompt message should not be marshal to json: %v", m))
	}
	return nil, nil
}

// Role ...
func (m PromptMessage) Role() MessageRole {
	return m.role
}

// ToolID ...
func (m PromptMessage) ToolID() string {
	return ""
}

// NewSystemMessage ...
func NewSystemMessage(text string) PromptMessage {
	return PromptMessage{
		role:    MessageRoleSystem,
		content: text,
	}
}

// NewUserTextMessage ...
func NewUserTextMessage(text string) PromptMessage {
	return PromptMessage{
		role:    MessageRoleUser,
		content: text,
	}
}

// NewUserMultipartMessage ...
func NewUserMultipartMessage(contents ...*MultipartContent) PromptMessage {
	return PromptMessage{
		role:             MessageRoleUser,
		multiPartContent: contents,
	}
}

// NewAssistantMessage ...
func NewAssistantMessage(text string) *AssistantMessage {
	m := &AssistantMessage{PromptMessage: PromptMessage{
		role:    MessageRoleAssistant,
		content: text,
	}}
	return m
}

// NewAssistantReasoningMessage ...
func NewAssistantReasoningMessage(text string) *AssistantMessage {
	m := &AssistantMessage{ReasoningContent: text}
	m.role = MessageRoleAssistant
	return m
}

func (m *AssistantMessage) WithToolCalls(toolCalls []*ToolCall) *AssistantMessage {
	m.ToolCalls = toolCalls
	return m
}

func (m *AssistantMessage) WithReasoningContent(content string) *AssistantMessage {
	m.ReasoningContent = content
	return m
}

// NewToolMessage ...
func NewToolMessage(text string, toolID string) *ToolPromptMessage {
	return &ToolPromptMessage{
		PromptMessage: PromptMessage{
			role:    MessageRoleTool,
			content: text,
			Name:    toolID,
		},
		toolID: toolID,
	}
}

// AssistantMessage ...
type AssistantMessage struct {
	PromptMessage
	ToolCalls        []*ToolCall `json:"tool_calls"`
	ReasoningContent string      `json:"reasoning_content"`
}

func (m AssistantMessage) GetToolCalls() []*ToolCall {
	return m.ToolCalls
}

func (m *AssistantMessage) String() string {
	if m.ReasoningContent != "" {
		return fmt.Sprintf("%s: reasoning: %s", m.role, m.ReasoningContent)
	}
	raw, _ := json.Marshal(m.ToolCalls)
	return fmt.Sprintf("%s => (content:%s tool_calls: %+v)", m.role, m.content, string(raw))
}

// ToolPromptMessage ...
type ToolPromptMessage struct {
	PromptMessage
	toolID string
}

// ToolID ...
func (m ToolPromptMessage) ToolID() string {
	return m.toolID
}

// ToolCall ...
type ToolCall struct {
	ID       string           `json:"id"`
	Type     string           `json:"type"`
	Index    int              `json:"index"`
	Function ToolCallFunction `json:"function"`
}

func (t *ToolCall) String() string {
	return fmt.Sprintf("ToolCall=> id: %s type: %s index: %d function: {name: %s arguments: %s}", t.ID, t.Type, t.Index, t.Function.Name, t.Function.Arguments)
}

// ToolCallFunction ...
type ToolCallFunction struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

// Messages ...
type Messages[T PromptMessage | AssistantMessage] []T

// Message ...
type Message interface {
	Content() string
	MultipartContent() []*MultipartContent
	Role() MessageRole
	ToolID() string
	GetToolCalls() []*ToolCall
	String() string
}

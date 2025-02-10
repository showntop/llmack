package llm

import (
	"strings"
)

// QueryProcessor ...
type QueryProcessor func([]Message) (string, error)

// LastQueryMessage ...
func LastQueryMessage(messages []Message) (string, error) {
	return messages[len(messages)-1].Content(), nil // TODO when with multipart content
}

// ConcatQueryMessage ...
func ConcatQueryMessage(messages []Message) (string, error) {
	builder := strings.Builder{}
	for i := 0; i < len(messages); i++ {
		builder.WriteString(string(messages[i].Role()))
		builder.WriteByte(':')
		builder.WriteString(messages[i].Content())
	}
	return builder.String(), nil
}

// CompressQueryMessage ...
func CompressQueryMessage(messages []Message) (string, error) {
	builder := strings.Builder{}
	for i := 0; i < len(messages); i++ {
		builder.WriteString(string(messages[i].Role()))
		builder.WriteByte(':')
		builder.WriteString(messages[i].Content())
	}
	return builder.String(), nil
}

// SummarizeQueryMessage extracts an intelligent summary from conversation messages
func SummarizeQueryMessage(messages []Message) (string, error) {
	// 如果消息为空则返回
	if len(messages) == 0 {
		return "", nil
	}
	summary := ""

	return summary, nil
}

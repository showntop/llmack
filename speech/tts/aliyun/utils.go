package aliyun

import (
	"strings"

	"github.com/google/uuid"
)

// Header is the basic struct for the TTS header.
type Header struct {
	MessageID string `json:"message_id"`
	TaskID    string `json:"task_id"`
	Namespace string `json:"namespace"`
	Name      string `json:"name"`
	Appkey    string `json:"appkey"`
}

func newHeader(taskID, name string) Header {
	return Header{
		TaskID:    taskID,
		Namespace: "FlowingSpeechSynthesizer",
		Name:      name,
		Appkey:    "N3CyrsLkHXD2yD3c",
		MessageID: Get32UUID(),
	}
}

// Response is the basic struct for the TTS response.
type Response struct {
	Header  Header                 `json:"header"`
	Payload map[string]interface{} `json:"payload,omitempty"`
}

// Request is the basic struct for the TTS Request.
type Request struct {
	Header  Header                 `json:"header"`
	Payload map[string]interface{} `json:"payload,omitempty"`
}

// Get32UUID ...
func Get32UUID() string {
	return strings.ReplaceAll(uuid.NewString(), "-", "")
}

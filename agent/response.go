package agent

import (
	"sync"

	"github.com/showntop/llmack/llm"
)

type TeamRunResponse struct {
	sync.RWMutex // for thread safe

	Reasoning       string         `json:"reasoning"`
	Answer          string         `json:"answer"`
	ToolCalls       []llm.ToolCall `json:"tool_calls"`
	Stream          chan any
	MemberResponses []*AgentRunResponse

	Error error
}

func (t *TeamRunResponse) AddMemberResponse(response *AgentRunResponse) {
	t.Lock()
	defer t.Unlock()

	if t.MemberResponses == nil {
		t.MemberResponses = []*AgentRunResponse{}
	}
	t.MemberResponses = append(t.MemberResponses, response)
}

func (t *TeamRunResponse) String() string {
	return t.Answer
}

type AgentRunResponse struct {
	streamx chan *llm.Chunk

	Reasoning string         `json:"reasoning"`
	Answer    string         `json:"answer"`
	ToolCalls []llm.ToolCall `json:"tool_calls"`
	Error     error

	Stream chan any
}

func (a *AgentRunResponse) Completion() string {
	for chunk := range a.streamx {
		a.Answer += chunk.Choices[0].Message.Content()
	}
	return a.Answer
}

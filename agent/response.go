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
	Stream          chan *llm.Chunk
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
	Reasoning string         `json:"reasoning"`
	Answer    string         `json:"answer"`
	ToolCalls []llm.ToolCall `json:"tool_calls"`
	Error     error

	Stream chan *llm.Chunk
}

func (a *AgentRunResponse) Completion() string {
	return a.Answer
}

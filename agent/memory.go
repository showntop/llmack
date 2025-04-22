package agent

import (
	"context"
	"sync"
)

type TeamMemory struct {
	sync.RWMutex

	SharedContexts SharedContexts
}

type SharedContexts struct {
	Context                string
	TeamMemberInteractions []TeamMemberInteraction
}

type TeamMemberInteraction struct {
	MemberName string
	Task       string
	Response   *AgentRunResponse
}

func (t *TeamMemory) SetSharedContext(ctx context.Context, text string) (string, error) {
	t.Lock()
	defer t.Unlock()

	t.SharedContexts.Context = text
	return t.SharedContexts.Context, nil
}

func (t *TeamMemory) GetSharedContext(ctx context.Context) (string, error) {
	t.RLock()
	defer t.RUnlock()

	return t.SharedContexts.Context, nil
}

func (t *TeamMemory) GetTeamMemberInteractions(ctx context.Context) ([]TeamMemberInteraction, error) {
	t.RLock()
	defer t.RUnlock()

	return t.SharedContexts.TeamMemberInteractions, nil
}

func (t *TeamMemory) AddTeamMemberInteractions(ctx context.Context, memberName string, task string, response *AgentRunResponse) error {
	t.Lock()
	defer t.Unlock()

	t.SharedContexts.TeamMemberInteractions = append(t.SharedContexts.TeamMemberInteractions, struct {
		MemberName string
		Task       string
		Response   *AgentRunResponse
	}{
		MemberName: memberName,
		Task:       task,
		Response:   response,
	})
	return nil
}

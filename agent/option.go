package agent

import (
	"github.com/showntop/llmack/llm"
	"github.com/showntop/llmack/memory"
	"github.com/showntop/llmack/rag"
	"github.com/showntop/llmack/storage"
)

type Option func(any)

func WithID(id string) Option {
	return func(a any) {
		if aa, ok := a.(*Agent); ok {
			aa.ID = id
		} else if at, ok := a.(*Team); ok {
			at.ID = id
		}
	}
}

func WithName(name string) Option {
	return func(a any) {
		if aa, ok := a.(*Agent); ok {
			aa.Name = name
		} else if at, ok := a.(*Team); ok {
			at.Name = name
		}
	}
}

func WithDescription(description string) Option {
	return func(a any) {
		if ax, ok := a.(*Agent); ok {
			ax.Description = description
		} else if at, ok := a.(*Team); ok {
			at.Description = description
		}
	}
}

func WithKnowledge(retrieval *rag.Indexer) Option {
	return func(a any) {
		if aa, ok := a.(*Agent); ok {
			aa.ragrtv = retrieval
		} else if at, ok := a.(*Team); ok {
			at.ragrtv = retrieval
		}
	}
}

func WithInstructions(instructions ...string) Option {
	return func(a any) {
		if aa, ok := a.(*Agent); ok {
			aa.Instructions = instructions
		} else if at, ok := a.(*Team); ok {
			at.Instructions = instructions
		}
	}
}

func WithModel(model *llm.Instance) Option {
	return func(a any) {
		if aa, ok := a.(*Agent); ok {
			aa.llm = model
		} else if at, ok := a.(*Team); ok {
			at.llm = model
		}
	}
}

func WithRole(role string) Option {
	return func(a any) {
		if aa, ok := a.(*Agent); ok {
			aa.Role = role
		} else if at, ok := a.(*Team); ok {
			at.Role = role
		}
	}
}

func WithTools(tools ...any) Option {
	return func(a any) {
		if aa, ok := a.(*Agent); ok {
			aa.Tools = tools
		} else if at, ok := a.(*Team); ok {
			at.Tools = tools
		}
	}
}

func WithLLM(llm *llm.Instance) Option {
	return func(a any) {
		if at, ok := a.(*Team); ok {
			at.llm = llm
		} else if aa, ok := a.(*Agent); ok {
			aa.llm = llm
		}
	}
}

func WithAgenticSharedContext(enable bool) Option {
	return func(a any) {
		if at, ok := a.(*Team); ok {
			at.agenticSharedContext = enable
		}
	}
}

func WithShareMemberInteractions(enable bool) Option {
	return func(a any) {
		if at, ok := a.(*Team); ok {
			at.shareMemberInteractions = enable
		}
	}
}

func WithMode(mode TeamMode) Option {
	return func(a any) {
		if a, ok := a.(*Team); ok {
			a.mode = mode
		}
	}
}

func WithMembers(members ...*Agent) Option {
	return func(a any) {
		if a, ok := a.(*Team); ok {
			a.members = members
		}
	}
}

func WithStorage(storage storage.Storage) Option {
	return func(a any) {
		if at, ok := a.(*Team); ok {
			at.storage = storage
		} else if aa, ok := a.(*Agent); ok {
			aa.storage = storage
		}
	}
}

func WithMemory(memory memory.Memory) Option {
	return func(a any) {
		if at, ok := a.(*Agent); ok {
			at.memory = memory
		}
	}
}

// func WithStream(stream bool) Option {
// 	return func(a any) {
// 		if aa, ok := a.(*Agent); ok {
// 			aa.stream = stream
// 		} else if at, ok := a.(*Team); ok {
// 			at.stream = stream
// 		}
// 	}
// }

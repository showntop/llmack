package memory

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/showntop/llmack/llm"
	"github.com/showntop/llmack/program"
	"github.com/showntop/llmack/tool"
)

type FlatMemory struct {
	model *llm.Instance

	sync.RWMutex
	memoryItems map[string][]*MemoryItem
}

func NewFlatMemory(model *llm.Instance) Memory {
	return &FlatMemory{
		model:       model,
		memoryItems: make(map[string][]*MemoryItem),
	}
}

func (m *FlatMemory) Get(ctx context.Context, sessionID string) ([]*MemoryItem, error) {
	m.RLock()
	defer m.RUnlock()

	fmt.Println("get memory", m.memoryItems[sessionID])
	return m.memoryItems[sessionID], nil
}

func (m *FlatMemory) Add(ctx context.Context, sessionID string, item *MemoryItem) error {
	m.RLock()
	history := m.memoryItems[sessionID]
	m.RUnlock()

	// history = append(history)

	// messages
	prompt += "\t4. Decide to delete an existing memory, using the 'delete_memory' tool."
	prompt += "\t5. Decide to clear all memories, using the 'clear_memory' tool."
	prompt += "You can call multiple tools in a single response if needed. "
	prompt += "Only add or update memories if it is necessary to capture key information provided by the user."
	if len(history) > 0 {
		prompt += "\n\n<existing_memories>\n"
		for _, item := range history {
			prompt += fmt.Sprintf("ID: %d\n", item.ID)
			prompt += fmt.Sprintf("Content: %s\n", item.Content)
			prompt += fmt.Sprintf("CreatedAt: %s\n", item.CreatedAt)
			prompt += "\n"
		}
		prompt += "</existing_memories>\n"
	}

	resp := program.FunCall().
		WithInstruction(prompt).
		WithTools(m.memoryTools(sessionID)...).
		WithInputs(map[string]any{
			"memory_capture_instructions": memoriesToCapture,
			// "existing_memories":   m.existingMemories,
		}).InvokeQuery(ctx, item.Content)

	fmt.Println("resp", resp.Completion())

	return resp.Error()
}

func (m *FlatMemory) memoryTools(sessionID string) []any {
	tool.Register(tool.New(
		tool.WithName("add_memory"),
		tool.WithDescription("Use this function to add a memory to the storage."),
		tool.WithParameters([]tool.Parameter{
			{
				Name:          "memory",
				LLMDescrition: "The memory to add to the storage.",
				Type:          tool.String,
				Required:      true,
			},
			{
				Name:          "topics",
				LLMDescrition: "The topics of the memory (e.g. ['name', 'hobbies', 'location']).",
				Type:          tool.String,
				Required:      true,
			},
		}),
		tool.WithFunction(func(ctx context.Context, args string) (string, error) {
			var params map[string]any
			if err := json.Unmarshal([]byte(args), &params); err != nil {
				return "", fmt.Errorf("failed to unmarshal arguments in json, %v", err)
			}

			topics := []string{}
			if _, ok := params["topics"].(string); ok {
				topics = strings.Split(params["topics"].(string), ",")
			} else if _, ok := params["topics"].([]any); ok {
				for _, topic := range params["topics"].([]any) {
					topics = append(topics, topic.(string))
				}
			}
			m.Lock()
			defer m.Unlock()

			m.memoryItems[sessionID] = append(m.memoryItems[sessionID], &MemoryItem{
				ID:        time.Now().Unix(),
				SessionID: sessionID,
				Content:   params["memory"].(string),
				Topics:    topics,
			})
			fmt.Println("set memory", m.memoryItems[sessionID])
			return "Memory added successfully", nil
		}),
	))
	tool.Register(tool.New(
		tool.WithName("update_memory"),
		tool.WithDescription("Use this function to update an existing memory."),
		tool.WithParameters([]tool.Parameter{
			{
				Name:          "memory",
				LLMDescrition: "The memory to update.",
				Type:          tool.String,
				Required:      true,
			},
		}),
		tool.WithFunction(func(ctx context.Context, args string) (string, error) {
			var params map[string]any
			if err := json.Unmarshal([]byte(args), &params); err != nil {
				return "", fmt.Errorf("failed to unmarshal arguments in json, %v", err)
			}

			m.Lock()
			defer m.Unlock()

			m.memoryItems[sessionID] = append(m.memoryItems[sessionID], &MemoryItem{
				ID:        time.Now().Unix(),
				SessionID: sessionID,
				Content:   params["memory"].(string),
			})
			return "Memory updated successfully", nil
		}),
	))
	tool.Register(tool.New(
		tool.WithName("delete_memory"),
		tool.WithDescription("Use this function to delete an existing memory."),
		tool.WithParameters([]tool.Parameter{
			{
				Name:          "memory",
				LLMDescrition: "The memory to delete.",
				Type:          tool.String,
				Required:      true,
			},
		}),
		tool.WithFunction(func(ctx context.Context, args string) (string, error) {
			var params map[string]any
			if err := json.Unmarshal([]byte(args), &params); err != nil {
				return "", fmt.Errorf("failed to unmarshal arguments in json, %v", err)
			}

			m.Lock()
			defer m.Unlock()

			m.memoryItems[sessionID] = append(m.memoryItems[sessionID], &MemoryItem{
				ID:        time.Now().Unix(),
				SessionID: sessionID,
				Content:   params["memory"].(string),
			})
			return "Memory deleted successfully", nil
		}),
	))
	return []any{
		"add_memory",
		"update_memory",
		"delete_memory",
	}
}

func (m *FlatMemory) FetchHistories(ctx context.Context, sessionID string) ([]*MemoryItem, error) {
	m.RLock()
	defer m.RUnlock()
	return m.memoryItems[sessionID], nil
}

var memoriesToCapture = `
Memories should include details that could personalize ongoing interactions with the user, such as:
	- Personal facts: name, age, occupation, location, interests, preferences, etc.
	- Significant life events or experiences shared by the user
	- Important context about the user's current situation, challenges or goals
	- What the user likes or dislikes, their opinions, beliefs, values, etc.
	- Any other details that provide valuable insights into the user's personality, perspective or needs
`

var prompt = `
You are a MemoryManager that is responsible for manging key information about the user.
You will be provided with a criteria for memories to capture in the <memories_to_capture> section and a list of existing memories in the <existing_memories> section.

## When to add or update memories
- Your first task is to decide if a memory needs to be added, updated, or deleted based on the user's message OR if no changes are needed.
- If the user's message meets the criteria in the <memories_to_capture> section and that information is not already captured in the <existing_memories> section, you should capture it as a memory.
- If the users messages does not meet the criteria in the <memories_to_capture> section, no memory updates are needed.
- If the existing memories in the <existing_memories> section capture all relevant information, no memory updates are needed.

## How to add or update memories
- If you decide to add a new memory, create memories that captures key information, as if you were storing it for future reference.
- Memories should be a brief, third-person statements that encapsulate the most important aspect of the user's input, without adding any extraneous information.
	- Example: If the user's message is 'I'm going to the gym', a memory could be 'John Doe goes to the gym regularly'.
	- Example: If the user's message is 'My name is John Doe', a memory could be 'User's name is John Doe'.
- Don't make a single memory too long or complex, create multiple memories if needed to capture all the information.
- Don't repeat the same information in multiple memories. Rather update existing memories if needed.
- If a user asks for a memory to be updated or forgotten, remove all reference to the information that should be forgotten. Don't say 'The user used to like ...'
- When updating a memory, append the existing memory with new information rather than completely overwriting it.
- When a user's preferences change, update the relevant memories to reflect the new preferences but also capture what the user's preferences used to be and what has changed.

## Criteria for creating memories
Use the following criteria to determine if a user's message should be captured as a memory.

<memories_to_capture>
{{memory_capture_instructions}}
</memories_to_capture>

## Updating memories
You will also be provided with a list of existing memories in the <existing_memories> section. You can:
	1. Decide to make no changes.
	2. Decide to add a new memory, using the 'add_memory' tool.
	3. Decide to update an existing memory, using the 'update_memory' tool.
`

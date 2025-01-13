package engine

import (
	"github.com/showntop/llmack/llm"
	"github.com/showntop/llmack/workflow"
)

// EventName 枚举类型
type EventName string

// 枚举值列表
const (
	EventToast   EventName = "toast"
	EventThought EventName = "thought"
	EventEnd     EventName = "end"
	EventError   EventName = "error"
	Ping         EventName = "ping"
	Stop         EventName = "stop"
)

// Event 消息事件接口
// *llm.Chunk
// ThoughID  string
type Event struct {
	Data     any
	Name     EventName
	Error    error
	Source   string
	SourceID string
}

func (e Event) String() string {
	return string(e.Name)
}

// LLMResultEvent ...
func LLMResultEvent(data *llm.Result) *Event {
	return &Event{
		Name: EventEnd,
		Data: data,
	}
}

// ToastEvent ...
func ToastEvent(data any) *Event {
	return &Event{
		Name: EventToast,
		Data: data,
	}
}

// LLMChunkEvent ...
func LLMChunkEvent(data *llm.Chunk) *Event {
	return &Event{
		Name: EventToast,
		Data: data,
	}
}

// ErrorEvent ...
func ErrorEvent(data error) *Event {
	return &Event{
		Name:  EventError,
		Error: data,
	}
}

// EndEvent ...
func EndEvent(data string) *Event {
	return &Event{
		Name: EventEnd,
		Data: data,
	}
}

// EndEventWithSource ...
func EndEventWithSource(data any, source, sourceID string) *Event {
	return &Event{
		Name: EventEnd,
		Data: data,
	}
}

// WorkflowEvent ...
func WorkflowEvent(data *workflow.Event) *Event {
	return &Event{
		Name: EventToast,
		Data: data,
	}
}

// WorkflowResultEvent ...
func WorkflowResultEvent(data *workflow.Result) *Event {
	return &Event{
		Name: EventToast,
		Data: data,
	}
}

// EventStream ...
type EventStream struct {
	events chan *Event
}

// NewEventStream ...
func NewEventStream() *EventStream {
	return &EventStream{
		events: make(chan *Event, 1),
	}
}

// Next ...
func (s *EventStream) Next() *Event {
	return <-s.events
}

// Push ...
func (s *EventStream) Push(e *Event) {
	s.events <- e
}

// Close ...
func (s *EventStream) Close() {
	close(s.events)
}

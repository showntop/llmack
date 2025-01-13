package engine

import (
	"context"
	"sync"

	"github.com/showntop/llmack/speech"
)

// Turn ...
type Turn struct {
	context context.Context
	engine  *Engine
	sync.Mutex
	ID string

	tts speech.StreamTTS

	pipeline Handler
}

// NewTurn ...
func NewTurn(r *Engine, id string) *Turn {
	var err error
	t := &Turn{ID: id, engine: r}
	t.context = r.Context()
	t.tts, err = r.warmUpTTS()
	if err != nil {
		panic(err)
	}
	t.pipeline = agentHandler(t, ttsHandler(t, outHandler(r, nil)))
	return t
}

// Context ...
func (t *Turn) Context() context.Context {
	return t.context
}

// Handle ...
func (t *Turn) Handle(msg *Message) error {
	return t.pipeline(msg)
}

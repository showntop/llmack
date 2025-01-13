package engine

import (
	"context"
	"fmt"
	"sync/atomic"
)

// Handler ...
type Handler func(msg *Message) error

// Message ...
type Message struct {
	ctx            context.Context
	SessionID      string
	ConversationID string // 会话ID
	TurnID         string // 轮次ID
	ID             string // 消息ID
	Payload        any    // 消息内容（当前）
	Question       string // 消息内容（本轮/文本）
	Answer         string // 回复内容（本轮/文本）
	interrupt      atomic.Bool

	onOut func(m *Message) error
}

func (m *Message) String() string {
	return fmt.Sprintf("Message{ID: %s, Payload: %s}", m.ID, m.Payload)
}

// Context ...
func (m *Message) Context() context.Context {
	// return m.ctx
	return context.Background()
}

// Interrupt ...
func (m *Message) Interrupt() {
	// log.InfoContextf(m.ctx, "interrupt message: %s", m.String())
	m.interrupt.Store(true)
}

// Interrupted ...
func (m *Message) Interrupted() bool {
	return m.interrupt.Load()
}

// OnOut ...
func (m *Message) OnOut() func(m *Message) error {
	return m.onOut
}

// WithOnOut ...
func (m *Message) WithOnOut(f func(m *Message) error) *Message {
	m.onOut = f
	return m
}

// WithTurnID ...
func (m *Message) WithTurnID(id string) *Message {
	m.TurnID = id
	return m
}

// NewMessage ...
func NewMessage(id string, payload any) *Message {
	return &Message{ID: id, Payload: payload}
}

// HandlerConstructor ...
type HandlerConstructor func(r *Engine, next Handler) Handler

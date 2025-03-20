package engine

import (
	"github.com/showntop/llmack/log"
	"github.com/showntop/llmack/rag/retrival"
)

const (
	// BotModeUnknown TODO
	BotModeUnknown = iota
	// BotModeProxy TODO
	BotModeProxy
	// BotModeChat TODO
	BotModeChat
	// BotModeWorkflow TODO
	BotModeWorkflow
	// BotModeSingleAgent TODO
	BotModeSingleAgent
	// BotModeMultiAgent TODO
	BotModeMultiAgent
	// BotModeSpeech TODO
	BotModeSpeech = 99
)

// BuildEngine ...
func BuildEngine(mode int, settings *Settings, opts ...Option) Engine {
	if mode == BotModeProxy { // 朴素模式，不使用大模型推理
		return NewProxyEngine(settings, opts...)
	} else if mode == BotModeChat { // 简单模式，问答
		return NewChatEngine(settings, opts...)
	} else if mode == BotModeWorkflow {
		return NewWorkflowEngine(settings, opts...)
	} else if mode == BotModeSingleAgent {
		return NewAgentEngine(settings, opts...)
	} else if mode == BotModeSpeech {
	}
	return nil
}

// Options ...
type Options struct {
	logger         log.Logger
	Rag            *retrival.Retrival
	Memory         Memory
	ConversationID int64
	MessageID      int64
	Reflect        int64
	Hooks          []Hook
}

// Option ...
type Option func(o *Options)

// WithLogger ...
func WithLogger(logger log.Logger) Option {
	return func(o *Options) {
		o.logger = logger
	}
}

// WithRag ...
func WithRag(rag *retrival.Retrival) Option {
	return func(o *Options) {
		o.Rag = rag
	}
}

// WithMemory ...
func WithMemory(m Memory) Option {
	return func(o *Options) {
		o.Memory = m
	}
}

// WithConversation ...
func WithConversation(m int64) Option {
	return func(o *Options) {
		o.ConversationID = m
	}
}

// WithMessageID ...
func WithMessageID(m int64) Option {
	return func(o *Options) {
		o.MessageID = m
	}
}

// WithHook ...
func WithHook(h Hook) Option {
	return func(o *Options) {
		o.Hooks = append(o.Hooks, h)
	}
}

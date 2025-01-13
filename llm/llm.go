package llm

import (
	"context"
	"fmt"
)

type logger interface {
	ErrorContextf(context.Context, string, ...interface{})
	InfoContextf(context.Context, string, ...interface{})
	WarnContextf(context.Context, string, ...interface{})
	DebugContextf(context.Context, string, ...interface{})
}

// Provider ...
type Provider interface {
	Invoke(context.Context, []Message, []PromptMessageTool, ...InvokeOption) (*Response, error)
}

// Instance ...
type Instance struct {
	name     string
	provider Provider
	opts     *Options
}

var providers = map[string]Provider{}

// Register ...
func Register(name string, provider Provider) {
	providers[name] = provider
}

// Options ...
type Options struct {
	hooks  []Hook
	cache  Cache
	logger logger
	model  string
}

// Option ...
type Option func(*Options)

// WithHook ...
func WithHook(hooks ...Hook) Option {
	return func(options *Options) {
		options.hooks = append(options.hooks, hooks...)
	}
}

// WithCache ...
func WithCache(c Cache) Option {
	return func(options *Options) {
		options.cache = c
	}
}

// WithDefaultModel ...
func WithDefaultModel(m string) Option {
	return func(options *Options) {
		options.model = m
	}
}

// WithLogger ...
func WithLogger(l logger) Option {
	return func(options *Options) {
		options.logger = l
	}
}

// NewInstance ...
func NewInstance(provider string, opts ...Option) *Instance {
	var options Options = Options{
		logger: &NoneLogger{},
	}
	for _, o := range opts {
		o(&options)
	}

	return &Instance{
		name:     provider,
		opts:     &options,
		provider: providers[provider],
	}
}

// Invoke ...
func (mi *Instance) Invoke(ctx context.Context,
	messages []Message, tools []PromptMessageTool, options ...InvokeOption) (*Response, error) {

	if mi.provider == nil {
		return nil, fmt.Errorf("llm provider of %v is not registered", mi.name)
	}
	if mi.opts != nil && mi.opts.hooks != nil {
		for _, hook := range mi.opts.hooks {
			hook.OnBeforeInvoke(ctx)
		}
	}

	response, err := mi.invoke(ctx, messages, tools, options...)

	if mi.opts != nil && mi.opts.hooks != nil {
		for _, hook := range mi.opts.hooks {
			hook.OnAfterInvoke(ctx, err)
		}
	}
	return response, err
}

func (mi *Instance) invoke(ctx context.Context,
	messages []Message, tools []PromptMessageTool, options ...InvokeOption) (*Response, error) {

	updateCache := func(ctx context.Context, result string) {} // nothing todo

	if mi.opts.cache != nil && len(tools) <= 0 { // fetch from cache
		document, hited, err := mi.opts.cache.Fetch(ctx, messages)
		if err != nil {
			return nil, err
		}
		if hited { // stream answer or return answer
			mi.opts.logger.InfoContextf(ctx, "llm cache hitted cache value %+v", document)
			return mi.handleStreamCache(ctx, document.Answer)
		}

		updateCache = func(ctx context.Context, result string) {
			if mi.opts.cache != nil && len(tools) <= 0 { // store cache
				mi.opts.logger.InfoContextf(ctx, "llm cache update cache value %v", result)
				if err := mi.opts.cache.Store(ctx, document, result); err != nil {
					return
				}
			}
		}

	}

	var invokeOpts InvokeOptions
	for i := 0; i < len(options); i++ {
		options[i](&invokeOpts)
	}

	if invokeOpts.Model == "" {
		options = append(options, WithModel(mi.opts.model))
	}

	response, err := mi.provider.Invoke(ctx, messages, tools, options...)
	if err != nil {
		return response, err
	}

	if invokeOpts.Stream {
		return mi.handleStreamResponse(ctx, response, updateCache), nil
	}

	return mi.handleBlockResponse(ctx, response, updateCache), nil
}

func (mi *Instance) handleBlockResponse(ctx context.Context, response *Response,
	updateCache func(context.Context, string)) *Response {
	if response == nil || response.result == nil || response.result.Message == nil {
		return response
	}
	if updateCache != nil {
		updateCache(ctx, response.result.Message.Content().Data)
	}
	return response
}

func (mi *Instance) handleStreamResponse(ctx context.Context, response *Response,
	updateCache func(context.Context, string)) *Response {
	newResp := NewStreamResponse()

	if response == nil || response.stream == nil {
		return newResp
	}
	go func() {
		defer newResp.stream.Close()
		// 倒腾一遍，for metrics and trace ... and cache
		result := ""
		firstChunk := true
		for chunk := response.stream.Next(); chunk != nil; chunk = response.stream.Next() {
			if firstChunk {
				for _, hook := range mi.opts.hooks {
					hook.OnFirstChunk(ctx, nil)
				}
			}
			firstChunk = false
			newResp.stream.Push(chunk)
			result += chunk.Delta.Message.content.Data
		}
		if updateCache != nil {
			updateCache(ctx, result)
		}
	}()
	return newResp
}

func (mi *Instance) handleStreamCache(ctx context.Context, answer string) (*Response, error) {
	respone := NewStreamResponse()
	go func() {
		defer respone.stream.Close()
		mi.opts.logger.InfoContextf(ctx, "llm cache stream cache value %s", answer)
		respone.stream.Push(NewChunk(
			0,
			AssistantPromptMessage(answer),
			nil,
		))
	}()
	return respone, nil
}

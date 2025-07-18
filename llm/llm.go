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
	Invoke(context.Context, []Message, *InvokeOptions) (*Response, error)
}

// Instance ...
type Instance struct {
	name     string
	provider Provider
	opts     *Options
}

type ProviderOptions struct {
	BaseURL string
	ApiKey  string
}
type ProviderConstructor func(*ProviderOptions) Provider

var providers = map[string]ProviderConstructor{}

// Register ...
func Register(name string, provider ProviderConstructor) {
	providers[name] = provider
}

// Options ...
type Options struct {
	baseURL string
	apiKey  string
	hooks   []Hook
	cache   Cache
	logger  logger
	model   string
	*InvokeOptions
}

// Option ...
type Option func(*Options)

// WithInvokeOptions ...
func WithInvokeOptions(o *InvokeOptions) Option {
	return func(options *Options) {
		options.InvokeOptions = o
	}
}

func WithBaseURL(baseURL string) Option {
	return func(options *Options) {
		options.baseURL = baseURL
	}
}

func WithAPIKey(apiKey string) Option {
	return func(options *Options) {
		options.apiKey = apiKey
	}
}

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
	po := &ProviderOptions{
		BaseURL: options.baseURL,
		ApiKey:  options.apiKey,
	}

	return &Instance{
		name:     provider,
		opts:     &options,
		provider: providers[provider](po),
	}
}

// Invoke ...
func (mi *Instance) Invoke(ctx context.Context,
	messages []Message, options ...InvokeOption) (*Response, error) {
	// 删除 options 中 nil
	for i := 0; i < len(options); i++ {
		if options[i] == nil {
			options = append(options[:i], options[i+1:]...)
			i--
		}
	}
	invokeOpts := mi.opts.InvokeOptions // default invoke options
	if invokeOpts == nil {
		invokeOpts = &InvokeOptions{}
	}
	for i := range options {
		if options[i] == nil {
			continue
		}
		options[i](invokeOpts)
	}
	return mi.InvokeWithOptions(ctx, messages, invokeOpts)
}

func (mi *Instance) InvokeWithOptions(ctx context.Context,
	messages []Message, invokeOpts *InvokeOptions) (*Response, error) {
	if mi.provider == nil {
		return nil, fmt.Errorf("llm provider of %v is not registered", mi.name)
	}
	if mi.opts != nil && mi.opts.hooks != nil {
		for _, hook := range mi.opts.hooks {
			ctx = hook.OnBeforeInvoke(ctx)
		}
	}
	response, err := mi.invoke(ctx, messages, invokeOpts)
	if mi.opts != nil && mi.opts.hooks != nil {
		for _, hook := range mi.opts.hooks {
			hook.OnAfterInvoke(ctx, err)
		}
	}
	return response, err
}

func (mi *Instance) invoke(ctx context.Context,
	messages []Message, invokeOpts *InvokeOptions) (*Response, error) {

	updateCache := func(ctx context.Context, result string) {} // nothing todo

	if mi.opts.cache != nil && len(invokeOpts.Tools) <= 0 { // fetch from cache
		document, hited, err := mi.opts.cache.Fetch(ctx, messages)
		if err != nil {
			return nil, err
		}
		if hited { // stream answer or return answer
			mi.opts.logger.InfoContextf(ctx, "llm cache hitted cache value %+v", document)
			return mi.handleStreamCache(ctx, document.Answer)
		}

		updateCache = func(ctx context.Context, result string) {
			if mi.opts.cache != nil && len(invokeOpts.Tools) <= 0 { // store cache
				mi.opts.logger.InfoContextf(ctx, "llm cache update cache value %v", result)
				if err := mi.opts.cache.Store(ctx, document, result); err != nil {
					return
				}
			}
		}

	}

	if invokeOpts.Model == "" {
		invokeOpts.Model = mi.opts.model
	}

	response, err := mi.provider.Invoke(ctx, messages, invokeOpts)
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
	// todo
	// _ = response.Result() // make result
	if updateCache != nil {
		updateCache(ctx, response.result.Message.content) // TODO
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

		for chunk := response.stream.Take(); chunk != nil; chunk = response.stream.Take() {
			if firstChunk {
				for _, hook := range mi.opts.hooks {
					ctx = hook.OnFirstChunk(ctx, nil)
				}
			}
			// if len(chunk.Choices) <= 0 {
			// 	mi.opts.logger.WarnContextf(ctx, "llm stream response chunk choices is empty")
			// 	continue
			// }
			firstChunk = false
			newResp.stream.Push(chunk)
			if len(chunk.Choices) > 0 {
				result += chunk.Choices[0].Delta.ReasoningContent
				result += chunk.Choices[0].Delta.content
			}
		}
		if updateCache != nil {
			updateCache(ctx, result)
		}
		for _, hook := range mi.opts.hooks {
			hook.OnLastChunk(ctx, nil)
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
			NewAssistantMessage(answer),
			nil,
		))
	}()
	return respone, nil
}

package vision

import (
	"context"
	"fmt"
)

// Provider ...
type Provider interface {
	GenerateImage(context.Context, string, ...InvokeOption) (string, error)
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
	hooks []Hook
	model string
}

// Option ...
type Option func(*Options)

// WithHook ...
// func WithHook(hooks ...Hook) Option {
// return func(options *Options) {
// options.hooks = append(options.hooks, hooks...)
// }
// }

// NewInstance ...
func NewInstance(provider string, opts ...Option) *Instance {
	var options Options = Options{}
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
func (mi *Instance) GenerateImage(ctx context.Context,
	prompt string, options ...InvokeOption) (string, error) {

	if mi.provider == nil {
		return "null", fmt.Errorf("llm provider of %v is not registered", mi.name)
	}
	if mi.opts != nil && mi.opts.hooks != nil {
		for _, hook := range mi.opts.hooks {
			ctx = hook.OnBeforeInvoke(ctx)
		}
	}

	// 删除 options 中 nil
	for i := 0; i < len(options); i++ {
		if options[i] == nil {
			options = append(options[:i], options[i+1:]...)
			i--
		}
	}
	response, err := mi.provider.GenerateImage(ctx, prompt, options...)
	if err != nil {
		return response, err
	}

	if mi.opts != nil && mi.opts.hooks != nil {
		for _, hook := range mi.opts.hooks {
			hook.OnAfterInvoke(ctx, err)
		}
	}
	return response, err
}

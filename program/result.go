package program

import (
	"context"
	"fmt"
	"reflect"

	"github.com/showntop/llmack/llm"
	"github.com/showntop/llmack/log"
)

type Result struct {
	p          *predictor
	err        error
	completion string
	stream     chan any
}

func (r *Result) Error() error {
	return r.err
}

func (r *Result) Get(value any) error {
	return r.p.adapter.Parse(r.completion, value)
}

func (r *Result) Completion() string {
	return r.completion
}

func (r *Result) Stream() chan any {
	return r.stream
}

// Result ...
func (p *predictor) Result(ctx context.Context, value any) error {
	if reflect.TypeOf(value).Kind() != reflect.Ptr {
		return fmt.Errorf("predictor result target value must be a pointer")
	}

	messages, err := p.adapter.Format(p, p.inputs, value)
	if err != nil {
		return err
	}
	response, err := p.model.Invoke(ctx, messages,
		llm.WithStream(true),
	)
	if err != nil {
		return err
	}
	completion := response.Result().Message.Content()
	log.InfoContextf(ctx, "response: %s", completion)

	return p.adapter.Parse(completion, value)
}

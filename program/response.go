package program

import (
	"context"
	"fmt"
	"reflect"

	"github.com/showntop/llmack/llm"
	"github.com/showntop/llmack/log"
)

func (p *predictor) Response() *Response {
	return p.reponse
}

func (p *predictor) Stream() chan *llm.Chunk {
	return p.reponse.stream
}

func (p *predictor) Completion() string {
	return p.reponse.message.Content()
}

func (p *predictor) Message() *llm.AssistantMessage {
	return p.reponse.message
}

func (p *predictor) Error() error {
	return p.reponse.err
}

func (p *predictor) ToolCalls() []*llm.ToolCall {
	return p.reponse.toolCalls
}

type Response struct {
	p          *predictor
	err        error
	completion string
	toolCalls  []*llm.ToolCall
	stream     chan *llm.Chunk

	message *llm.AssistantMessage

	HasToolCalls bool
}

func NewResponse() *Response {
	return &Response{
		stream:  make(chan *llm.Chunk, 1000),
		message: &llm.AssistantMessage{},
	}
}

func (r *Response) Error() error {
	return r.err
}

func (r *Response) Get(value any) error {
	return r.p.adapter.Parse(r.completion, value)
}

func (r *Response) Completion() string {
	return r.completion
}

func (r *Response) ToolCalls() []*llm.ToolCall {
	return r.toolCalls
}

func (r *Response) Stream() chan *llm.Chunk {
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

package program

import (
	"context"
	"reflect"

	"github.com/showntop/llmack/llm"
)

// predictor ...
type predictor struct {
	Mode      string
	stream    bool
	model     *llm.Instance
	adapter   Adapter
	inputs    map[string]any
	memory    Memory
	observers []llm.Message
	tools     []any
	Promptx

	invoker Invoker
	reponse *Response
}

type Invoker interface {
	Invokex(ctx context.Context, query string) *predictor
	Invoke(ctx context.Context, inputs map[string]any) *predictor
}

func NewPredictor(opts ...option) *predictor {
	p := &predictor{
		adapter: &DefaultAdapter{},
		Promptx: Promptx{InputFields: make(map[string]*Field), OutputFields: make(map[string]*Field)},
	}
	for i := range opts {
		opts[i](p)
	}
	// default invoker
	p.invoker = p
	if p.model == nil {
		p.model = defaultLLM
	}
	return p
}

// Predictor ...
func Predictor(opts ...option) *predictor {
	p := &predictor{
		adapter: &DefaultAdapter{},
		Promptx: Promptx{InputFields: make(map[string]*Field), OutputFields: make(map[string]*Field)},
	}
	for i := range opts {
		opts[i](p)
	}
	if p.model == nil {
		p.model = defaultLLM
	}
	// default invoker
	p.invoker = p

	return p
}

// WithAdapter ...
func (p *predictor) WithAdapter(adapter Adapter) *predictor {
	p.adapter = adapter
	return p
}

// WithInstruction ...
func (p *predictor) WithInstruction(i string) *predictor {
	p.Promptx.Instruction = i
	return p
}

// WithInputField ...
func (p *predictor) WithInputField(key string, value string) *predictor {
	p.InputFields[key] = &Field{Description: value, Name: key}
	return p
}

// WithInputFields ...
func (p *predictor) WithInputFields(inputs map[string]string) *predictor {
	for k, v := range inputs {
		p.InputFields[k] = &Field{Description: v, Name: k}
	}
	return p
}

// WithOutputField tuple is name, description, marker, type
func (p *predictor) WithOutputField(tuple ...any) *predictor {
	if len(tuple) <= 0 {
		return p
	}
	out := &Field{Name: tuple[0].(string)}
	if len(tuple) >= 2 {
		out.Description = tuple[1].(string)
	}
	if len(tuple) >= 3 {
		out.Marker = tuple[2].(string)
	}
	if len(tuple) >= 4 {
		out.Type = tuple[3].(reflect.Kind)
	}
	p.Promptx.OutputFields[out.Name] = out
	return p
}

func (p *predictor) WithInputs(inputs map[string]any) *predictor {
	p.inputs = inputs
	return p
}

func (p *predictor) WithTools(tools ...any) *predictor {
	p.tools = tools
	return p
}

func (p *predictor) WithStream(stream bool) *predictor {
	p.stream = stream
	return p
}

// InvokeQuery invoke forward for predicte
func (p *predictor) InvokeQuery(ctx context.Context, query string) *predictor {
	p.reponse = NewResponse()
	if p.stream {
		go p.invoker.Invokex(ctx, query)
		return p
	}
	return p.invoker.Invokex(ctx, query)
}

func (p *predictor) Invokex(ctx context.Context, query string) *predictor {
	messages, err := p.adapter.Format(p, p.inputs, p.reponse)
	if err != nil {
		p.reponse.err = err
		return p
	}

	response, err := p.model.Invoke(ctx, messages,
		llm.WithStream(true),
	)
	if err != nil {
		p.reponse.err = err
		return p
	}

	stream := response.Stream()

	// 合并 message
	answerContent := ""
	for chunk := range stream.Next() {
		p.reponse.stream <- chunk
		answerContent += chunk.Choices[0].Message.Content()
	}

	p.reponse.message = llm.NewAssistantMessage(answerContent)
	return p
}

// Invoke invoke forward for predicte
func (p *predictor) Invoke(ctx context.Context, inputs map[string]any) *predictor {
	messages, err := p.adapter.Format(p, inputs, p.reponse)
	if err != nil {
		p.reponse.err = err
		return p
	}
	response, err := p.model.Invoke(ctx, messages,
		llm.WithStream(true),
	)
	if err != nil {
		p.reponse.err = err
		return p
	}
	completion := response.Result().Message.Content()
	p.reponse.message = llm.NewAssistantMessage(completion)
	return p
}

func (r *predictor) FetchHistoryMessages(ctx context.Context) []llm.Message {
	return nil
}

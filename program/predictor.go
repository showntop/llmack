package program

import (
	"context"
	"reflect"

	"github.com/showntop/llmack/llm"
)

// predictor ...
type predictor struct {
	Usage           llm.Usage
	Mode            string
	stream          bool
	toolChoice      any
	model           *llm.Instance
	maxIterationNum int
	adapter         Adapter
	inputs          map[string]any
	observers       []llm.Message
	tools           []any
	Promptx

	resetMessages func(ctx context.Context, messages []llm.Message) []llm.Message
	invoker       Invoker
	reponse       *Response
}

// Invoker 定义了 predictor 的调用方式
type Invoker interface {
	InvokeOnce(ctx context.Context, messages []llm.Message) *predictor
	Invoke(ctx context.Context, messages []llm.Message, query string, inputs map[string]any) *predictor
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
	p.Promptx.Instruction += "\n" + i
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
	p.tools = append(p.tools, tools...)
	return p
}

func (p *predictor) WithStream(stream bool) *predictor {
	p.stream = stream
	return p
}

func (p *predictor) WithToolChoice(toolChoice any) *predictor {
	p.toolChoice = toolChoice
	return p
}

// InvokeQuery invoke forward for predicte
func (p *predictor) InvokeQuery(ctx context.Context, query string) *predictor {
	p.reponse = NewResponse()
	if p.stream {
		go p.invoker.Invoke(ctx, nil, query, p.inputs)
		return p
	}
	return p.invoker.Invoke(ctx, nil, query, p.inputs)
}

// InvokeWithMessages invoke forward for predicte with messages
func (p *predictor) InvokeWithMessages(ctx context.Context, messages []llm.Message) *predictor {
	p.reponse = NewResponse()
	if p.stream {
		go p.invoker.Invoke(ctx, messages, "", p.inputs)
		return p
	}
	return p.invoker.Invoke(ctx, messages, "", p.inputs)
}

// Invoke invoke forward for predicte
func (p *predictor) Invoke(ctx context.Context, messages []llm.Message, query string, inputs map[string]any) *predictor {
	systemMessages, err := p.adapter.Format(p, inputs, p.reponse)
	if err != nil {
		p.reponse.err = err
		return p
	}
	messages = append(systemMessages, messages...)
	if len(query) > 0 {
		messages = append(messages, llm.NewUserTextMessage(query))
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

func (p *predictor) InvokeOnce(ctx context.Context, messages []llm.Message) *predictor {
	panic("not implemented")
}

func (p *predictor) InvokeOnceWithMessages(ctx context.Context, messages []llm.Message) *predictor {
	p.reponse = NewResponse()
	if p.stream {
		go p.invoker.InvokeOnce(ctx, messages)
		return p
	}
	return p.invoker.InvokeOnce(ctx, messages)
}

// FetchHistoryMessages fetch history messages
func (p *predictor) FetchHistoryMessages(ctx context.Context) []llm.Message {
	return nil
}

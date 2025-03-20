package program

import (
	"context"
	"fmt"
	"reflect"

	"github.com/showntop/llmack/llm"
	"github.com/showntop/llmack/log"
)

// predictor ...
type predictor struct {
	model   *llm.Instance
	adapter Adapter
	inputs  map[string]any

	Promptx
}

// NewPredictor ...
func NewPredictor(i string, inputs map[string]string, opts ...option) *predictor {

	p := &predictor{
		adapter: &JSONAdapter{},
		model:   defaultLLM,
		Promptx: Promptx{Instruction: i, InputFields: make(map[string]*Field), OutputFields: make(map[string]*Field)},
	}

	for i := 0; i < len(opts); i++ {
		opts[i](p)
	}
	return p
}

// NewPredictorWithPrompt ...
func NewPredictorWithPrompt(prompt *Promptx, opts ...option) *predictor {

	p := &predictor{
		adapter: &JSONAdapter{},
		model:   defaultLLM,
		Promptx: *prompt,
	}

	for i := 0; i < len(opts); i++ {
		opts[i](p)
	}
	return p
}

// Predictor ...
func Predictor(opts ...option) *predictor {
	p := &predictor{
		adapter: &JSONAdapter{},
		Promptx: Promptx{InputFields: make(map[string]*Field), OutputFields: make(map[string]*Field)},
	}
	for i := 0; i < len(opts); i++ {
		opts[i](p)
	}
	if p.model == nil {
		p.model = defaultLLM
	}

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

// ChatWith ...
func (p *predictor) ChatWith(inputs map[string]any) *predictor {
	p.inputs = inputs
	return p
}

// Prompt ...
func (p *predictor) Prompt() *Promptx {
	return &p.Promptx
}

// Update ...
func (p *predictor) Update(opts ...option) {
	for i := 0; i < len(opts); i++ {
		opts[i](p)
	}
}

// Forward // TODO: implement forward pass
func (p *predictor) Forward(ctx context.Context, inputs map[string]any) (any, error) {
	var value any
	if err := p.ChatWith(inputs).Result(ctx, &value); err != nil {
		return value, err
	}

	return value, nil
}

// Invoke invoke forward for predicte
func (p *predictor) Invoke(ctx context.Context, inputs map[string]any) *Result {
	var value Result
	value.p = p
	messages, err := p.adapter.Format(p, inputs, value)
	if err != nil {
		value.err = err
		return &value
	}
	response, err := p.model.Invoke(ctx, messages,
		llm.WithStream(true),
	)
	if err != nil {
		value.err = err
		return &value
	}
	completion := response.Result().Message.Content()
	value.completion = completion
	return &value
}

type Result struct {
	p          *predictor
	err        error
	completion string
	stream     chan any
}

func (r *Result) Get(value any) error {
	return r.p.adapter.Parse(r.completion, value)
}

func (r *Result) Error() error {
	return r.err
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

func (r *predictor) FetchHistoryMessages(ctx context.Context) []llm.Message {
	return nil
}

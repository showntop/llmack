package program

import (
	"context"
	"reflect"

	"github.com/showntop/llmack/llm"
	"github.com/showntop/llmack/log"
	"github.com/showntop/llmack/prompt"
)

// predictor ...
type predictor struct {
	model   *llm.Instance
	adapter OutAdapter
	inputs  map[string]any

	Promptx
}

type option func(*predictor)

// WithInstruction ...
func WithInstruction(info string) option {
	return func(p *predictor) {
		p.Promptx.Instruction = info
	}
}

// WithOutput ...
func WithOutput(tuple ...any) option {
	return func(p *predictor) {
		if len(tuple) <= 0 {
			return
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
		// 重复 ？
		p.Promptx.OutputFields[out.Name] = out
	}
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
func Predictor() *predictor {
	return &predictor{
		model:   defaultLLM,
		adapter: &JSONAdapter{},
		Promptx: Promptx{InputFields: make(map[string]*Field), OutputFields: make(map[string]*Field)},
	}
}

// WithInstruction ...
func (p *predictor) WithInstruction(i string) *predictor {
	p.Promptx.Instruction = i
	return p
}

// WithAdapter ...
func (p *predictor) WithAdapter(adapter OutAdapter) *predictor {
	p.adapter = adapter
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

// WithOutputField ...
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

// FormatWith implement format for {{$xxx}}
func (p *predictor) FormatWith(inputs map[string]any) *predictor {
	x, err := prompt.Render(p.Promptx.Instruction, inputs)
	// @TODO handle error
	if err != nil {
	}
	p.Promptx.Instruction = x
	return p
}

// Forward // TODO: implement forward pass
func (p *predictor) Forward(ctx context.Context, inputs map[string]any) (any, error) {
	var value any
	if err := p.ChatWith(inputs).Result(ctx, &value); err != nil {
		return value, err
	}

	return value, nil
}

// Result ...
func (p *predictor) Result(ctx context.Context, value any) error {
	if reflect.TypeOf(value).Kind() != reflect.Ptr {
		// return fmt.Errorf("result target value must be a pointer")
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
	completion := response.Result().Message.Content().Data
	log.InfoContextf(ctx, "response: %s", completion)

	return p.adapter.Parse(completion, value)
}

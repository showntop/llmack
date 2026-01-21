package tool

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"sort"
	"strings"

	"github.com/flosch/pongo2/v6"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3gen"
	"github.com/showntop/llmack/log"
	"github.com/showntop/llmack/pkg/structx"
)

type InvokeFunc func(context.Context, string) (string, error)

type Tool struct {
	ID           int64  // 引用的工具ID
	Kind         string // 引用的工具类型
	Name         string
	Description  string
	*ParamsOneOf `json:"parameters,omitempty"` // 参数，可选

	AuthenticationType  string
	AuthenticationValue string
	ServerURL           string `json:"server_url"` // 服务器URL
	Method              string `json:"method"`     // 方法
	Body                string `json:"body"`       // body

	invoke InvokeFunc
}

func (t *Tool) WithParameters(parameters ...Parameter) *Tool {
	t.ParamsOneOf = &ParamsOneOf{
		params1: parameters,
	}
	return t
}

func (t *Tool) WithInvokeFunc(invoke InvokeFunc) *Tool {
	t.invoke = invoke
	return t
}

type Option func(*Tool)

func New(opts ...Option) *Tool {
	t := &Tool{
		Kind:        "code",
		ParamsOneOf: &ParamsOneOf{},
	}
	for _, opt := range opts {
		opt(t)
	}
	return t
}

func WithKind(kind string) Option {
	return func(t *Tool) {
		t.Kind = kind
	}
}

func WithName(name string) Option {
	return func(t *Tool) {
		t.Name = name
	}
}

func WithDescription(description string) Option {
	return func(t *Tool) {
		t.Description = description
	}
}

func WithParameters(parameters ...any) Option {
	return func(t *Tool) {
		if len(parameters) == 0 {
			return
		}
		if t.ParamsOneOf == nil {
			t.ParamsOneOf = &ParamsOneOf{}
		}
		for _, parameter := range parameters {
			if x, ok := parameter.(Parameter); ok {
				t.ParamsOneOf.params1 = append(t.ParamsOneOf.params1, x)
			} else if x, ok := parameter.(*openapi3.Schema); ok {
				t.ParamsOneOf.params2 = x
			}
		}
	}
}

func WithFunction(function func(ctx context.Context, args string) (string, error)) Option {
	return func(t *Tool) {
		t.invoke = function
	}
}

func (t *Tool) Invoke(ctx context.Context, params string) (string, error) {
	if t.Kind == "api" {
		return t.invokeAPI(ctx, params)
	}
	return t.invoke(ctx, params)
}

func (t *Tool) invokeAPI(ctx context.Context, args string) (string, error) {
	var params map[string]any
	if err := json.Unmarshal([]byte(args), &params); err != nil {
		return "", fmt.Errorf("failed to unmarshal arguments in json, %v", err)
	}
	// Extract the URL and request parameters
	url := t.ServerURL
	method := t.Method

	var rawx string
	if t.Body != "" {
		tpl, err := pongo2.FromString(t.Body)
		if err != nil {
			return "", errors.Join(err)
		}
		rawx, _ = tpl.Execute(params)
	} else {
		raw, _ := json.Marshal(params)
		rawx = string(raw)
	}

	log.InfoContextf(ctx, "Send HTTP url:%s method: %s request: %s", url, method, string(rawx))
	req, err := http.NewRequestWithContext(ctx, method, url, strings.NewReader(rawx))
	if err != nil {
		return "", errors.Join(err)
	}
	req.Header.Add("Content-Type", "application/json")
	if t.AuthenticationType != "" {
		req.Header.Add("Authorization", t.AuthenticationValue)
	}
	// Send the HTTP request
	client := http.DefaultClient
	resp, err := client.Do(req)
	if err != nil {
		log.ErrorContextf(ctx, "failed to send HTTP request: %s with error: %s", rawx, err)
		return "", errors.Join(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		log.ErrorContextf(ctx, "failed to send HTTP request: %s with error: %s", rawx, resp.Status)
		return "", errors.New(resp.Status)
	}

	// Read and process the response
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		log.ErrorContextf(ctx, "failed to get HTTP request: %+v with error: %v", req, err)
		return "", errors.Join(err)
	}
	log.InfoContextf(ctx, "get HTTP request: %+v with reponse %s", req, string(bodyBytes))

	return string(bodyBytes), nil
}

// DefaultSchemaCustomizer is the default schema customizer when using reflect to infer tool parameter from tagged go struct.
// Supported struct tags:
// 1. jsonschema: "description=xxx"
// 2. jsonschema: "enum=xxx,enum=yyy,enum=zzz"
// 3. jsonschema: "required"
// 4. can also use json: "xxx,omitempty" to mark the field as not required, which means an absence of 'omitempty' in json tag means the field is required.
// If this DefaultSchemaCustomizer is not sufficient or suitable to your specific need, define your own SchemaCustomizerFn and pass it to WithSchemaCustomizer during InferTool or InferStreamTool.
func DefaultSchemaCustomizer(name string, t reflect.Type, tag reflect.StructTag, schema *openapi3.Schema) error {
	jsonS := tag.Get("jsonschema")
	if len(jsonS) > 0 {
		tags := strings.Split(jsonS, ",")
		for _, t := range tags {
			kv := strings.Split(t, "=")
			if len(kv) == 2 {
				if kv[0] == "description" {
					schema.Description = kv[1]
				}
				if kv[0] == "enum" {
					schema.Enum = append(schema.Enum, kv[1])
				}
			} else if len(kv) == 1 {
				if kv[0] == "required" {
					if schema.Extensions == nil {
						schema.Extensions = make(map[string]any, 1)
					}
					schema.Extensions["x_required"] = true
				}
			}
		}
	}

	json := tag.Get("json")
	if len(json) > 0 && !strings.Contains(json, "omitempty") {
		if schema.Extensions == nil {
			schema.Extensions = make(map[string]any, 1)
		}
		schema.Extensions["x_required"] = true
	}

	if name == "_root" {
		if err := setRequired(schema); err != nil {
			return err
		}
	}

	return nil
}

func setRequired(sc *openapi3.Schema) error { // check if properties are marked as required, set schema required to true accordingly
	if sc.Type != nil && sc.Type.Is(openapi3.TypeObject) && sc.Type.Is(openapi3.TypeArray) {
		return nil
	}
	if sc.Type.Is(openapi3.TypeArray) {
		if sc.Items.Value.Extensions != nil {
			if _, ok := sc.Items.Value.Extensions["x_required"]; ok {
				delete(sc.Items.Value.Extensions, "x_required")
				if len(sc.Items.Value.Extensions) == 0 {
					sc.Items.Value.Extensions = nil
				}
			}
		}

		if err := setRequired(sc.Items.Value); err != nil {
			return fmt.Errorf("setRequired for array failed: %w", err)
		}
	}

	for k, p := range sc.Properties {
		if p.Value.Extensions != nil {
			if _, ok := p.Value.Extensions["x_required"]; ok {
				sc.Required = append(sc.Required, k)
				delete(p.Value.Extensions, "x_required")
				if len(p.Value.Extensions) == 0 {
					p.Value.Extensions = nil
				}
			}

		}
		err := setRequired(p.Value)
		if err != nil {
			return fmt.Errorf("setRequired for nested property %s failed: %w", k, err)
		}
	}

	sort.Strings(sc.Required)

	return nil
}

type ToolFunc[T, D any] func(ctx context.Context, input T) (output D, err error)

func NewWithToolFunc[T, D any](name string, description string, function ToolFunc[T, D]) (*Tool, error) {
	fun := func(ctx context.Context, args string) (string, error) {
		var inst T
		inst = structx.NewInstance[T]()

		if err := json.Unmarshal([]byte(args), &inst); err != nil {
			return "", fmt.Errorf("failed to unmarshal arguments in json, %v", err)
		}

		resp, err := function(ctx, inst)
		if err != nil {
			return "", fmt.Errorf("failed to execute action, %v", err)
		}

		output, err := json.Marshal(resp)
		if err != nil {
			return "", fmt.Errorf("failed to marshal output in json, %v", err)
		}

		return string(output), nil
	}

	schemaCustomizer := DefaultSchemaCustomizer

	sc, err := openapi3gen.NewSchemaRefForValue(structx.NewInstance[T](), nil, openapi3gen.SchemaCustomizer(schemaCustomizer))
	if err != nil {
		return nil, fmt.Errorf("new SchemaRef from T failed: %w", err)
	}

	tool := New(
		WithName(name),
		WithDescription(description),
		WithParameters(sc.Value),
		WithFunction(fun),
	)

	return tool, nil
}

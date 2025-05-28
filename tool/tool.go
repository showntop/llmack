package tool

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/flosch/pongo2/v6"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/showntop/llmack/log"
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
		Kind: "code",
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

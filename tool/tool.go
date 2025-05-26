package tool

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"

	"github.com/flosch/pongo2/v6"
	"github.com/showntop/llmack/log"
)

type Tool struct {
	ID          int64  // 引用的工具ID
	Kind        string // 引用的工具类型
	Name        string
	Description string
	Parameters  []Parameter `json:"parameters,omitempty"` // 参数，可选

	AuthenticationType  string
	AuthenticationValue string
	ServerURL           string `json:"server_url"` // 服务器URL
	Method              string `json:"method"`     // 方法
	Body                string `json:"body"`       // body

	Invokex func(context.Context, map[string]any) (string, error)
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

func WithParameters(parameters []Parameter) Option {
	return func(t *Tool) {
		t.Parameters = parameters
	}
}

func WithFunction(function func(ctx context.Context, args map[string]any) (string, error)) Option {
	return func(t *Tool) {
		t.Invokex = function
	}
}

func (t *Tool) Invoke(ctx context.Context, params map[string]any) (string, error) {
	if t.Kind == "api" {
		return t.invokeAPI(ctx, params)
	}
	return t.Invokex(ctx, params)
}

func (t *Tool) invokeAPI(ctx context.Context, params map[string]any) (string, error) {
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

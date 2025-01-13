package tool

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/expr-lang/expr"
	"github.com/flosch/pongo2/v6"

	"github.com/showntop/llmack/log"
)

// APIToolBundle 结构体，用于存储基于API的工具的schema信息
type APIToolBundle struct {
	Meta
	Type                int
	AuthenticationType  string
	AuthenticationValue string
	PostCode            string      `json:"post_code"`              // 代码
	ServerURL           string      `json:"server_url"`             // 服务器URL
	Method              string      `json:"method"`                 // 方法
	Summary             string      `json:"summary,omitempty"`      // 摘要，可选
	Body                string      `json:"body"`                   // body
	OperationID         string      `json:"operation_id,omitempty"` // 操作ID，可选
	Parameters          []Parameter `json:"parameters,omitempty"`   // 参数，可选
	Author              string      `json:"author"`                 // 作者
	Icon                string      `json:"icon,omitempty"`         // 图标，可选
	// OpenAPI             json.RawMessage `json:"openapi"`                // OpenAPI操作
}

// APITool 结构体，用于表示基于API的工具
type APITool struct {
	// Bundle APIToolBundle
	APIToolBundle
	// PostCode string
}

// ToolName 返回工具名称
func (t *APITool) ToolName() string {
	// 实现 ToolName 方法
	return t.APIToolBundle.Name
}

// Kind 返回工具类型
func (t *APITool) Kind() string {
	return "api"
}

// Name 返回工具名称
func (t *APITool) Name() string {
	// 实现 ToolName 方法
	return t.APIToolBundle.Name
}

// Description 返回工具名称
func (t *APITool) Description() string {
	// 实现 ToolName 方法
	return t.APIToolBundle.Description
}

// Parameters 返回工具名称
func (t *APITool) Parameters() map[string]any {
	xxx := make(map[string]any, len(t.Meta.Parameters))
	for i := 0; i < len(t.Meta.Parameters); i++ {
		xxx[t.Meta.Parameters[i].Name] = t.Meta.Parameters[i]
	}
	return xxx
}

// Invoke 调用工具
func (t *APITool) Invoke(ctx context.Context, params map[string]any) (string, error) {

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

	log.InfoContextf(ctx, "PostCode %s", t.PostCode)
	if t.PostCode != "" {
		code := t.PostCode
		var env map[string]any
		json.Unmarshal(bodyBytes, &env)
		program, err := expr.Compile(code, expr.Env(env))
		if err != nil {
			return "", errors.Join(err)
		}
		output, err := expr.Run(program, env)
		if err != nil {
			return "", errors.Join(err)
		}
		log.InfoContextf(ctx, "PostCode result %s", fmt.Sprint(output))
		return fmt.Sprint(output), nil

	}
	return string(bodyBytes), nil
}

// Stream 调用工具
func (t *APITool) Stream(ctx context.Context, params map[string]any) (<-chan any, error) {
	resultChan := make(chan any, 1)

	raw, _ := json.Marshal(params)
	req, err := http.NewRequestWithContext(ctx, t.Method, t.ServerURL, bytes.NewReader(raw))
	// Send the HTTP request
	client := http.DefaultClient
	resp, err := client.Do(req)
	if err != nil {
		log.ErrorContextf(ctx, "failed to send http request: %+v with error: %v", req, err)
		return nil, errors.Join(err)
	}
	log.InfoContextf(ctx, "send http request: %+v with success", req)
	// if resp.Header.Get("Content-Type") == "text/event-stream" {

	// }

	go func() {
		defer resp.Body.Close()

		scanner := bufio.NewScanner(resp.Body)
		for scanner.Scan() {
			raw := scanner.Bytes()
			log.InfoContextf(ctx, "send http response: %s ", string(raw))
			if len(raw) == 0 {
				continue
			}

			var env map[string]any
			json.Unmarshal(raw, &env)
			program, err := expr.Compile(t.PostCode, expr.Env(env))
			if err != nil {
				close(resultChan)
				return
				// return errors.Join(err)
			}
			output, err := expr.Run(program, env)
			if err != nil {
				close(resultChan)
				// return "", errors.Join(err)
			}
			resultChan <- fmt.Sprint(output)
		}
		close(resultChan)
	}()

	return resultChan, nil
}

// NewAPITool ...
func NewAPITool(m APIToolBundle) Tool {
	t := &APITool{}
	t.APIToolBundle = m
	return t
}

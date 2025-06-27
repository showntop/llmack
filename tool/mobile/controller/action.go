package controller

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/getkin/kin-openapi/openapi3gen"
	"github.com/showntop/llmack/pkg/structx"
	"github.com/showntop/llmack/tool"
)

type ActionFunc[T, D any] func(ctx context.Context, input T) (output D, err error)

type Action struct {
	Name string
	Tool *tool.Tool
}

func NewAction[T, D any](name string, description string, actionFunc ActionFunc[T, D]) (*Action, error) {
	fun := func(ctx context.Context, args string) (string, error) {
		var inst T
		inst = structx.NewInstance[T]()

		if err := json.Unmarshal([]byte(args), &inst); err != nil {
			return "", fmt.Errorf("failed to unmarshal arguments in json, %v", err)
		}

		resp, err := actionFunc(ctx, inst)
		if err != nil {
			return "", fmt.Errorf("failed to execute action, %v", err)
		}

		output, err := json.Marshal(resp)
		if err != nil {
			return "", fmt.Errorf("failed to marshal output in json, %v", err)
		}

		return string(output), nil
	}

	schemaCustomizer := tool.DefaultSchemaCustomizer

	sc, err := openapi3gen.NewSchemaRefForValue(structx.NewInstance[T](), nil, openapi3gen.SchemaCustomizer(schemaCustomizer))
	if err != nil {
		return nil, fmt.Errorf("new SchemaRef from T failed: %w", err)
	}

	tool := tool.New(
		tool.WithName(name),
		tool.WithDescription(description),
		tool.WithParameters(sc.Value),
		tool.WithFunction(fun),
	)

	return &Action{
		Name: name,
		Tool: tool,
	}, nil
}

package tool

import (
	"github.com/getkin/kin-openapi/openapi3"
)

// ParameterType 枚举类型
type ParameterType string

const (
	Array       ParameterType = "array"
	String      ParameterType = "string"
	Number      ParameterType = "number"
	Boolean     ParameterType = "boolean"
	Select      ParameterType = "select"
	SecretInput ParameterType = "secret-input"
	File        ParameterType = "file"
)

// Parameter 用于描述API参数的结构体
type Parameter struct {
	Name             string
	Label            string
	HumanDescription string
	Placeholder      string
	Type             ParameterType
	LLMDescrition    string
	Required         bool
	Default          any
	Min              float64
	Max              float64
	Options          []string
}

type ParamsOneOf struct {
	params1 []Parameter
	params2 *openapi3.Schema
}

func (p *ParamsOneOf) Parameters() any {
	if p == nil {
		return nil
	}
	if len(p.params1) > 0 {
		properties := map[string]any{}
		required := []string{}
		for _, p := range p.params1 {
			properties[p.Name] = map[string]any{
				"description": p.LLMDescrition,
				"type":        p.Type,
			}
			if p.Required {
				required = append(required, p.Name)
			}
		}
		return map[string]any{
			"type":       "object",
			"properties": properties,
			"required":   required,
		}
	}
	return p.params2
}

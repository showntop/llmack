package tool

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

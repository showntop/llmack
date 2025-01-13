// Package prompt handles prompt representation and templating
package prompt

// Prompt 提示词
type Prompt struct {
	Version            string
	InitialInstruction string         // 初始指令，用于生成prompt
	Description        string         // 简短描述任务
	Inputs             map[string]any // 输入
	Outputs            any            // 输出
	Metadata           map[string]any // 元数据
}

// New creates a new prompt
func New() *Prompt {
	p := &Prompt{
		Inputs:   make(map[string]any),
		Metadata: make(map[string]any),
	}
	return p
}

// Render renders the prompt as string
func (p *Prompt) Render(inputs map[string]any) string {
	xxx, err := Render(p.InitialInstruction, inputs)
	if err != nil {
		panic(err)
	}
	return xxx
}

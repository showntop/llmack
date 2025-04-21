package agents

import (
	"github.com/showntop/llmack/llm"
	"github.com/showntop/llmack/llm/deepseek"
)

var (
	MODEL_NAME_R1 = "ep-20250227113433-vv7hr"
	MODEL_NAME_V3 = "ep-20250227112432-tlpgl"
)

var (
	modelV3 = llm.NewInstance(deepseek.Name, llm.WithDefaultModel(MODEL_NAME_V3))
	modelR1 = llm.NewInstance(deepseek.Name, llm.WithDefaultModel(MODEL_NAME_R1))
)

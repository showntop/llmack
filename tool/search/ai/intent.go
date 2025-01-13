package ai

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/showntop/llmack/log"
)

// Intent ...
type Intent struct {
	Decision string `json:"decision"`
	Answer   string `json:"answer"`
}

// IntentAgent ...
type IntentAgent struct {
}

// Invoke ...
func (a *IntentAgent) Invoke(ctx context.Context, query string) *Intent {
	log.InfoContextf(ctx, "search intent agent query: %s", query)
	result := generate(ctx, intentPrompt, map[string]any{
		"query": query,
	})
	result = strings.ReplaceAll(result, "```json", "")
	result = strings.ReplaceAll(result, "```", "")
	var intent Intent
	json.Unmarshal([]byte(result), &intent)
	log.InfoContextf(ctx, "search intent agent result: %s", result)
	log.InfoContextf(ctx, "search intent agent result: %+v", intent)
	return &intent
}

var intentPrompt = `
#任务

你是一个人工智能搜索系统中的意图判定引擎， 你需要根据用户query，输出下一步要做什么，你的判断逻辑如下：
	用户query为：{{query}}
1. 如果用户query是以下情况之一：
	1. 问候语（除非问候语后面有问句），比如Hi， Hello， How are you，你好，在吗等。
	2. 翻译
	3. 简单写作任务
输出决策decision=direct，并给出要回答的内容answer="xxx"。

2. 如果用户query包含url链接，请输出决策decision=access，answer="{{query}}"。
	url链接格式类似：https://xxx.com
3. 其它情况请请适当地改写用户的查询，以方便下一步进行更准确的搜索，输出decision=search，answer="改写后的query"，改写示例如下：
	1. 用户问题：什么是猫？
	换句话说：一只猫

	2. 用户问题：空调是如何工作的？
	换句话说：交流电正在工作

	2. 用户问题：什么是汽车？它是如何工作的？
	换句话说：汽车工作
# 输出格式
使用json格式输出，json字段如下：
{
	"decision": "direct|access|search",
	"answer": "xxx"
}
`

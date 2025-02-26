package workflow

import (
	"github.com/showntop/llmack/llm/qwen"
	"github.com/showntop/llmack/tool/crawl"
	"github.com/showntop/llmack/tool/search"
	"github.com/showntop/llmack/workflow"
)

var intentNode = workflow.Node{ID: "intent", Kind: workflow.NodeKindLLM, // a small model for this task
	Metadata: map[string]any{
		"stream":   false,
		"provider": qwen.Name,
		"model":    "qwen-plus",
		"user_prompt": `
你是一个人工智能搜索系统中的意图判定引擎

#任务
根据用户query，输出下一步要做什么，你的判断逻辑如下：
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
`,
	},
	Outputs: workflow.Parameters{
		"decision": workflow.Parameter{Type: "string"},
		"answer":   workflow.Parameter{Type: "string"},
	},
}

func BuildWorkflow() *workflow.Workflow {
	wf := workflow.NewWorkflow(1, "search-agent").Link(
		workflow.Node{ID: "start", Kind: workflow.NodeKindStart},
		intentNode, // 输出 intent.decision, intent.answer
		workflow.Node{ID: "intentResult", Kind: workflow.NodeKindExpr, Metadata: map[string]any{
			"expr": "fromJSON(intent.message.content)",
		}},
		workflow.Node{ID: "gateway", Kind: workflow.NodeKindGateway, Subref: "exclusive",
			Inputs: workflow.Parameters{
				"decision": workflow.Parameter{Value: "{{intentResult.decision}}"},
				"query":    workflow.Parameter{Value: "{{intentResult.answer}}"},
			},
		},
	)
	// 以下为三个分支
	// 1. branch 1 for direct answer
	wf.LinkWithCondition("gateway", `decision=="direct"`, workflow.Node{ID: "end", Kind: workflow.NodeKindEnd,
		Outputs: workflow.Parameters{
			"answer": workflow.Parameter{Value: "{{intentResult.answer}}"},
		},
	})

	// 2. branch 2 for access website
	wf.LinkWithCondition("gateway", `decision=="access"`,
		workflow.Node{ID: "crawlWebPage", Kind: workflow.NodeKindTool,
			Inputs: workflow.Parameters{
				"link": workflow.Parameter{Value: "{{intentResult.answer}}"},
			},
			Metadata: map[string]any{
				"tool_name": crawl.Jina,
			},
		},
		workflow.Node{ID: "summarizeResult", Kind: workflow.NodeKindLLM,
			Metadata: map[string]any{
				"provider":    qwen.Name,
				"model":       "qwen-plus",
				"user_prompt": `总结以下内容: '''{{crawlWebPage.result}}'''`,
			},
		},
		workflow.Node{ID: "end2", Kind: workflow.NodeKindEnd,
			Outputs: workflow.Parameters{
				"answer": workflow.Parameter{Value: "{{summarizeResult.response}}"},
			},
		})

	// 3. branch 3 for hybrid search and summarize use llm
	wf.LinkWithCondition("gateway", `decision=="search"`,
		workflow.Node{
			ID:   "searchFromEngine",
			Kind: workflow.NodeKindTool,
			Inputs: workflow.Parameters{
				"query": workflow.Parameter{Value: "{{intentResult.answer}}"},
			},
			Metadata: map[string]any{
				"tool_name": search.Serper,
			},
		},
		AnswerWithLLMNode,
		generateRelatedQuestionNode,
		workflow.Node{ID: "end3", Kind: workflow.NodeKindEnd,
			Outputs: workflow.Parameters{
				"answer":  workflow.Parameter{Value: "{{answerWithLLM.response}}"},
				"related": workflow.Parameter{Value: "{{generateRelatedQuestion.response}}"},
			},
		})
	return wf
}

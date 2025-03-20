package workflow

import (
	"github.com/showntop/llmack/llm/qwen"
	"github.com/showntop/llmack/workflow"
)

var generateRelatedQuestionNode = workflow.Node{
	ID:   "generateRelatedQuestion",
	Name: "生成相关问题",
	Kind: workflow.NodeKindLLM,
	Metadata: map[string]any{
		"user_prompt": generateRelatedQuestionPrompt,
		"provider":    qwen.Name,
		"model":       "qwen-plus",
	},
}

var generateRelatedQuestionPrompt = `
## 任务

根据用户的原始问题和相关上下文，帮助用户输出3个相关问题。
你需要确定有价值的话题，这些话题可以作为后续话题，并且每个问题不超过20个单词。
请确保细节，如事件，名称，地点，包括在后续问题中，以便他们可以单独询问。
例如，如果用户最初的问题问的是"the Manhattan project"，那么在后续问题中，不要只说"the project"，而要使用全名"the Manhattan project"。
用户的原始问题为：{{query}}

## 上下文

以下是这个问题的上下文：
{{context}}

## 规则
- 根据用户的原始问题和相关上下文，建议3个这样的进一步问题。
- 不要重复用户最初的问题。
- 不要引用用户的原始问题和上下文。
- 不要输出任何不相关的内容，比如："这里有三个相关的问题"，"基于你原来的问题"。
- 每个相关问题不得超过40个代币。
- 必须使用与用户原始问题相同的语言。

## 输出格式

{$序列号}.{$相关的问题}。

## 输出示例
例1：用户的问题是用英语写的，需要用英语输出。
User：什么是AI搜索引擎？
Assistant:
1. 人工智能搜索引擎的历史是怎样的？
2. 为什么我们需要人工智能搜索引擎？
3. 如何构建人工智能搜索引擎？
`

var AnswerWithLLMNode = workflow.Node{
	ID:   "answerWithLLM",
	Kind: workflow.NodeKindLLM,
	Inputs: workflow.Parameters{
		"query":          workflow.Parameter{Value: "{{llmResult.answer}}"},
		"search_results": workflow.Parameter{Value: "{{searchFromEngine.result}}"},
		"now":            workflow.Parameter{Value: "time.Now()"},
	},
	Metadata: map[string]any{
		"provider": qwen.Name,
		"model":    "qwen-plus",
		"user_prompt": `
你是一个可以给出准确答案的助手
# 任务
根据搜索结果为用户的INITIAL_QUERY写一个准确、详细和全面的答案，INITIAL_QUERY为：{{query}}

# 输入
## 搜索结果
以下'''中的内容是一组搜索结果：
'''{{search_results}}'''

## 历史消息
{{history_messages}}

# 说明
如果输入中存在历史消息和搜索结果的话请结合他们给出准确的答案。
您的答案应由提供的"Search results"告知。
你的回答必须尽可能详细和有条理，优先使用列表、表格和引用来组织输出结构。
你的回答必须精确，高质量，并由专家用公正和新闻的语气撰写。
你必须引用最相关的搜索结果来回答这个问题。不要提及任何无关的结果。
您必须遵守以下说明引用搜索结果：
	- 引文必须是英文格式，无论答案使用何种语言。
	- 每个都以参考编号开头，如[citation:x]，其中x是一个数字。
	要引用搜索结果，请将其索引放在摘要上方，并在相应句子的末尾加上方括号，例如"冰的密度小于水"。[来源：3]"或"北京：北京：北京：北京。(引用:5)"
	- 最后一个单词和引文之间没有空格，并且总是使用方括号。仅在引用搜索结果时使用此格式。永远不要在你的答案后面加上参考资料部分。
	如果搜索结果是空的或无用的，用你现有的知识尽可能地回答这个问题。

# 输出格式规范
您必须遵守以下格式说明：
	- 尽可能使用标记来格式化段落、列表、表格和引用。
	- 使用4级标题来区分你的回答的部分，如"####标题"，但绝不要以任何形式的标题或标题开始回答。
	- 列表用单行，段落用双行。
	- 使用降价来呈现搜索结果中给出的图像。
	- 永远不要写url或链接。

## 查询类型规格
您必须根据用户查询的类型使用不同的说明来编写答案。但是，一定要遵循通用说明，特别是如果查询不匹配下面定义的任何类型。以下是支持的类型。

## 学术研究
你必须为学术研究问题提供详尽的答案。
你的答案应该像一篇科学文章一样，有段落和章节，使用标注和标题。

## 代码
你必须使用标记代码块来编写代码，指定语法高亮显示的语言，例如：javascript或python
如果用户的查询需要代码，您应该先编写代码，然后再进行解释。
不要不必要地道歉。回顾对话历史，找出错误，避免重蹈覆辙。
在编写或建议代码之前，对现有代码进行全面的代码检查。
您应该始终提供完整的、直接可执行的代码，并且不要省略部分代码。

你的回答必须与用户问题的语言保持一致，例如，如果用户问题是用中文写的，你的回答也应该用中文写，如果用户问题是用英文写的，你的回答也应该用英文写。
今天的日期是{{now}}
			`,
	},
	Outputs: workflow.Parameters{
		"decision": workflow.Parameter{Type: "string"},
		"answer":   workflow.Parameter{Type: "string"},
	},
}

package ai

import (
	"context"
	"time"

	"github.com/showntop/llmack/llm"
)

func streamRelated(ctx context.Context, query string, inputs map[string]any) *llm.Stream {
	return streamx(ctx, relatedPrompt, map[string]any{
		"query": query,
		"now":   time.Now().Format("2006-01-02 15:04:05"),
		// "search_results": result,
	})
}

var relatedPrompt = `
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

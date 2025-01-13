package deepseek

// ChatCompletionsResponse ...
type ChatCompletionsResponse struct {
	// 本次请求的 RequestId。
	ID string `json:"id,omitnil,omitempty" name:"id"`
	// Unix 时间戳，单位为秒。
	Created int64 `json:"created,omitnil,omitempty" name:"created"`
	// 回复内容。
	Choices []*Choice `json:"choices,omitnil,omitempty" name:"choices"`
	// Token 统计信息。
	Usage             Usage  `json:"usage,omitnil,omitempty" name:"usage"`
	Model             string `json:"model,omitnil,omitempty" name:"model"`
	SystemFingerprint string `json:"system_fingerprint,omitnil,omitempty" name:"system_fingerprint"`
}

// ToolCallFunction ...
type ToolCallFunction struct {
	// function名称
	Name string `json:"name,omitnil,omitempty" name:"name"`
	// function参数，一般为json字符串
	Arguments string `json:"arguments,omitnil,omitempty" name:"arguments"`
}

// ToolCall ...
type ToolCall struct {
	// 工具调用id
	ID string `json:"id,omitnil,omitempty" name:"id"`
	// 工具调用类型，当前只支持function
	Type string `json:"type,omitnil,omitempty" name:"type"`
	// 具体的function调用
	Function *ToolCallFunction `json:"function,omitnil,omitempty" name:"function"`
}

// Delta ...
type Delta struct {
	// 角色名称。
	Role string `json:"role,omitnil,omitempty" name:"role"`
	// 内容详情。
	Content string `json:"content,omitnil,omitempty" name:"content"`
	// 模型生成的工具调用，仅 hunyuan-functioncall 模型支持
	// 说明：
	// 对于每一次的输出值应该以Id为标识对Type、Name、Arguments字段进行合并。
	//
	// 注意：此字段可能返回 null，表示取不到有效值。
	ToolCalls []*ToolCall `json:"tool_calls,omitnil,omitempty" name:"tool_calls"`
}

// Choice ...
type Choice struct {
	// 结束标志位，可能为 stop 或 sensitive。
	// stop 表示输出正常结束，sensitive 只在开启流式输出审核时会出现，表示安全审核未通过。
	FinishReason string `json:"finish_reason,omitnil,omitempty" name:"finish_reason"`
	// 增量返回值，流式调用时使用该字段。
	// 注意：此字段可能返回 null，表示取不到有效值。
	Delta *Delta `json:"delta,omitnil,omitempty" name:"delta"`
	// 返回值，非流式调用时使用该字段。
	// 注意：此字段可能返回 null，表示取不到有效值。
	Message *Message `json:"message,omitnil,omitempty" name:"message"`
}

// Usage ...
type Usage struct {
	// 输入 Token 数量。
	PromptTokens int64 `json:"prompt_tokens,omitnil,omitempty" name:"prompt_tokens"`
	// 输出 Token 数量。
	CompletionTokens int64 `json:"completion_tokens,omitnil,omitempty" name:"completion_tokens"`
	// 用户 prompt 中，命中上下文缓存的 token 数
	PromptCacheHitTokens int64 `json:"prompt_cache_hit_tokens,omitnil,omitempty" name:"prompt_cache_hit_tokens"`
	// 用户 prompt 中，未命中上下文缓存的 token 数
	PromptCacheMissTokens int64 `json:"prompt_cache_miss_tokens,omitnil,omitempty" name:"prompt_cache_miss_tokens"`
	// 总 Token 数量。
	TotalTokens int64 `json:"total_tokens,omitnil,omitempty" name:"total_tokens"`
}

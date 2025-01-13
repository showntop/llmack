package deepseek

// ChatCompletionsRequest ...
type ChatCompletionsRequest struct {
	Stream            bool       `json:"stream,omitnil,omitempty" name:"stream"`
	Model             string     `json:"model,omitnil,omitempty" name:"model"`
	Messages          []*Message `json:"messages,omitnil,omitempty" name:"messages"`
	TopP              float64    `json:"top_p,omitnil,omitempty" name:"top_p"`
	Temperature       float64    `json:"temperature,omitnil,omitempty" name:"temperature"`
	EnableEnhancement bool       `json:"enable_enhancement,omitnil,omitempty" name:"enable_enhancement"`
	// 默认是false，在值为true且命中搜索时，接口会返回SearchInfo
	SearchInfo bool    `json:"searchInfo,omitnil,omitempty" name:"searchInfo"`
	Citation   bool    `json:"citation,omitnil,omitempty" name:"citation"`
	Stop       bool    `json:"stop,omitnil,omitempty" name:"stop"`
	Tools      []*Tool `json:"tools,omitnil,omitempty" name:"tools"`
	ToolChoice string  `json:"ToolChoice,omitnil,omitempty" name:"ToolChoice"`
}

// Message ...
type Message struct {
	// 角色，可选值包括 system、user、assistant、 tool。
	Role string `json:"role,omitnil,omitempty" name:"role"`
	// 文本内容
	Content string `json:"content,omitnil,omitempty" name:"content"`
	// 多种类型内容（目前支持图片和文本），仅 hunyuan-vision 模型支持
	// 注意：此字段可能返回 null，表示取不到有效值。
	Contents []*Content `json:"contents,omitnil,omitempty" name:"contents"`
	// 当role为tool时传入，标识具体的函数调用
	// 注意：此字段可能返回 null，表示取不到有效值。
	ToolCallID string `json:"ToolCallId,omitnil,omitempty" name:"ToolCallId"`
	// 模型生成的工具调用，仅 hunyuan-functioncall 模型支持
	// 注意：此字段可能返回 null，表示取不到有效值。
	ToolCalls []*ToolCall `json:"tool_calls,omitnil,omitempty" name:"tool_calls"`
}

// Content ...
type Content struct {
	// 内容类型
	// 注意：
	// 当前只支持传入单张图片，传入多张图片时，以第一个图片为准。
	// 注意：此字段可能返回 null，表示取不到有效值。
	Type string `json:"type,omitnil,omitempty" name:"type"`

	// 当 Type 为 text 时使用，表示具体的文本内容
	// 注意：此字段可能返回 null，表示取不到有效值。
	Text string `json:"text,omitnil,omitempty" name:"text"`

	// 图片的url，当 Type 为 image_url 时使用，表示具体的图片内容
	// 如"https://example.com/1.png" 或 图片的base64（注意 "data:image/jpeg;base64," 为必要部分）："data:image/jpeg;base64,/9j/4AAQSkZJRgABAQAA......"
	// 注意：此字段可能返回 null，表示取不到有效值。
	ImageURL *ImageURL `json:"image_url,omitnil,omitempty" name:"image_url"`
}

// ImageURL ...
type ImageURL struct {
	// 图片的 Url（以 http:// 或 https:// 开头）
	// 注意：此字段可能返回 null，表示取不到有效值。
	URL string `json:"url,omitnil,omitempty" name:"url"`
}

// Tool ...
type Tool struct {
	// 工具类型，当前只支持function
	Type string `json:"type,omitnil,omitempty" name:"type"`
	// 具体要调用的function
	Function *ToolFunction `json:"function,omitnil,omitempty" name:"function"`
}

// ToolFunction ...
type ToolFunction struct {
	// function名称，只能包含a-z，A-Z，0-9，\_或-
	Name string `json:"name,omitnil,omitempty" name:"name"`
	// function参数，一般为json字符串
	Parameters string `json:"parameters,omitnil,omitempty" name:"parameters"`
	// function的简单描述
	Description string `json:"description,omitnil,omitempty" name:"description"`
}

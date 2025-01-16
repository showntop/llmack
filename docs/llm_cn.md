# LLM 使用文档

## 概述
LLM模块提供了与大型语言模型交互的统一接口，支持多种LLM提供商，包括OpenAI、Azure OpenAI、DeepSeek等。模块采用插件化设计，易于扩展新的LLM提供商。

## 核心接口

### LLM接口
```go
type Provider interface {
    Invoke(context.Context, []Message, []PromptMessageTool, ...InvokeOption) (*Response, error)
}
```

### 消息类型
```go
type Message interface {
    Content() *PromptMessageContent
    Role() PromptMessageRole
    ToolID() string
    String() string
}
```

## 使用示例

### 基本使用
```go
// 创建LLM实例
llm := NewInstance("openai", WithDefaultModel("gpt-4"))

// 构造消息
messages := []Message{
    SystemPromptMessage("你是一个有帮助的助手"),
    UserPromptMessage("你好，世界"),
}

// 调用LLM
response, err := llm.Invoke(context.Background(), messages, nil)
if err != nil {
    log.Fatal(err)
}

// 处理响应
fmt.Println(response.Result().String())
```

### 流式响应
```go
// 创建LLM实例
llm := NewInstance("openai", WithDefaultModel("gpt-4"))

// 构造消息
messages := []Message{
    SystemPromptMessage("你是一个有帮助的助手"),
    UserPromptMessage("你好，世界"),
}

// 调用LLM
response, err := llm.Invoke(context.Background(), messages, nil, WithStream(true))
if err != nil {
    log.Fatal(err)
}

// 处理流式响应
for chunk := response.Stream().Next(); chunk != nil; chunk = response.Stream().Next() {
    fmt.Print(chunk.Delta.Message.content.Data)
}
```

## 配置选项

### 缓存配置
```go
// 使用内存缓存
llm := NewInstance("openai", 
    WithDefaultModel("gpt-4"),
    WithCache(NewMemoCache()),
)

// 使用Redis缓存
llm := NewInstance("openai",
    WithDefaultModel("gpt-4"),
    WithCache(NewRedisCache(redisClient)),
)
```

### 日志配置
```go
// 使用自定义日志
llm := NewInstance("openai",
    WithDefaultModel("gpt-4"),
    WithLogger(customLogger),
)
```

## 工具扩展
```go
// 定义工具
tool := PromptMessageTool{
    Name: "weather",
    Description: "获取天气信息",
    Parameters: map[string]any{
        "location": "string",
    },
}

// 调用带工具的LLM
response, err := llm.Invoke(context.Background(), messages, []PromptMessageTool{tool})
if err != nil {
    log.Fatal(err)
}
```

## 性能优化

### 批处理
```go
// 启用批处理
llm := NewInstance("openai",
    WithDefaultModel("gpt-4"),
    WithBatchSize(10),
)
```

### 异步调用
```go
// 异步调用
go func() {
    response, err := llm.Invoke(context.Background(), messages, nil)
    if err != nil {
        log.Fatal(err)
    }
    fmt.Println(response.Result().String())
}()
```

## 最佳实践

1. 使用缓存提高性能
2. 合理设置超时时间
3. 监控LLM调用指标
4. 实现重试机制
5. 使用工具扩展功能
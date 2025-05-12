# Agent 和 Team 使用教程

## 简介

Agent 和 Team 是一个强大的 AI 代理系统，允许你创建单个 AI 代理或由多个代理组成的团队来完成复杂的任务。系统支持多种工作模式，包括路由模式、协调模式和协作模式。

## Agent 基础

### 创建 Agent

```go
agent := NewAgent("agent_name", 
    WithRole("role_description"),
    WithDescription("detailed_description"),
    WithGoals([]string{"goal1", "goal2"}),
    WithInstructions([]string{"instruction1", "instruction2"}),
    WithOutputs([]string{"output1", "output2"}),
)
```

### Agent 主要属性

- `Name`: 代理名称
- `Role`: 代理角色
- `Description`: 详细描述
- `Goals`: 目标列表
- `Instructions`: 指令列表
- `Outputs`: 输出格式列表
- `Tools`: 可用工具列表

### 调用 Agent

```go
response := agent.Invoke(ctx, "your task here",
    WithSessionID("session_id"),
    WithRetries(3),
    WithStream(true),
)
```

## Team 功能

### 创建 Team

```go
team := NewTeam(TeamModeCoordinate,
    WithName("team_name"),
    WithDescription("team_description"),
    WithInstructions([]string{"instruction1", "instruction2"}),
)
```

### Team 工作模式

1. **路由模式 (TeamModeRoute)**
   - 团队 leader 根据用户请求选择合适的 agent 处理任务
   - 适用于任务分配明确的场景

2. **协调模式 (TeamModeCoordinate)**
   - 团队 leader 将任务拆分给不同的 agent 处理
   - 最后综合所有 agent 的结果给出答案
   - 适用于需要多步骤协作的复杂任务

3. **协作模式 (TeamModeCollaborate)**
   - 团队成员之间进行深度协作
   - 适用于需要持续交互和反馈的任务

### Team 特性

- **共享上下文**: 通过 `agenticSharedContext` 启用团队成员间的上下文共享
- **成员交互**: 通过 `shareMemberInteractions` 启用成员间的交互记录
- **会话管理**: 支持会话持久化和状态管理

## 使用示例

### 创建单个 Agent

```go
// 创建一个研究助手
researcher := NewAgent("researcher",
    WithRole("Research Assistant"),
    WithDescription("Expert in conducting thorough research and analysis"),
    WithGoals([]string{
        "Gather relevant information",
        "Analyze data thoroughly",
        "Present findings clearly",
    }),
)
```

### 创建协作团队

```go
// 创建一个研究团队
researchTeam := NewTeam(TeamModeCoordinate,
    WithName("Research Team"),
    WithDescription("A team of specialized researchers"),
    WithInstructions([]string{
        "Coordinate research efforts",
        "Ensure comprehensive coverage",
        "Maintain high quality standards",
    }),
)

// 添加团队成员
researchTeam.AddMember(researcher)
researchTeam.AddMember(analyst)
researchTeam.AddMember(writer)
```

## 最佳实践

1. **明确角色定义**
   - 为每个 agent 定义清晰的角色和职责
   - 确保团队成员之间角色互补

2. **任务分解**
   - 在协调模式下，将复杂任务分解为可管理的子任务
   - 为每个子任务设定明确的预期输出

3. **上下文管理**
   - 合理使用共享上下文功能
   - 注意控制上下文大小，避免信息过载

4. **错误处理**
   - 使用重试机制处理临时性错误
   - 实现适当的错误日志和监控

## 注意事项

1. 确保每个 agent 都有明确的职责范围
2. 合理设置重试次数，避免无限循环
3. 注意控制流式输出的缓冲区大小
4. 定期清理过期的会话数据
5. 监控团队成员之间的交互，确保协作效率

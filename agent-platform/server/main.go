package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/joho/godotenv"
	"github.com/showntop/llmack/agent"
	"github.com/showntop/llmack/llm"
	openaic "github.com/showntop/llmack/llm/openai-c"
)

func init() {
	godotenv.Load()

	llm.WithConfigs(map[string]any{
		"doubao": map[string]any{
			"base_url": "https://ark.cn-beijing.volces.com/api/v3",
			"api_key":  os.Getenv("doubao_api_key"),
		},
		"anthropic": map[string]any{
			"base_url": "http://v2.open.venus.oa.com/llmproxy",
			"api_key":  os.Getenv("claude_api_key"),
		},
		openaic.Name: map[string]any{
			"base_url": "http://v2.open.venus.oa.com/llmproxy",
			"api_key":  os.Getenv("claude_api_key"),
		},
		"qwen": map[string]any{
			"api_key": os.Getenv("qwen_api_key"),
		},
	})
}

var claudeModel = llm.NewInstance(openaic.Name,
	llm.WithDefaultModel("claude-3-7-sonnet-20250219"),
	llm.WithInvokeOptions(&llm.InvokeOptions{
		Metadata: map[string]any{
			"thinking_enabled": "true",
			"thinking_tokens":  2048,
		},
	}),
)

// AgentInterface 定义通用 Agent 接口
type AgentInterface interface {
	Invoke(ctx context.Context, message string, maxIterations int) InvokeResult
}

// AIAgent 模拟 AI Agent 结构
type AIAgent struct {
	Name  string
	Model string
}

// MobileAgentWrapper 包装真实的 MobileAgent
type MobileAgentWrapper struct {
	agent *agent.MobileAgent
}

// InvokeResult AI调用结果
type InvokeResult struct {
	Content string
	Error   error
}

// NewAIAgent 创建 AI Agent
func NewAIAgent(name string, model string) *AIAgent {
	return &AIAgent{
		Name:  name,
		Model: model,
	}
}

// NewMobileAgentWrapper 创建 MobileAgent 包装器
func NewMobileAgentWrapper() *MobileAgentWrapper {
	mobileAgent := agent.NewMobileAgent("mobile use agent",
		agent.WithModel(claudeModel),
	)
	return &MobileAgentWrapper{agent: mobileAgent}
}

// Invoke 调用 AI Agent
func (a *AIAgent) Invoke(ctx context.Context, message string, maxIterations int) InvokeResult {
	// 模拟 AI 处理逻辑
	log.Printf("AI Agent [%s] 正在处理: %s", a.Name, message)

	// 模拟处理时间
	time.Sleep(2 * time.Second)

	// 根据消息内容生成智能回复
	response := generateIntelligentResponse(message)

	return InvokeResult{
		Content: response,
		Error:   nil,
	}
}

// Invoke 调用 Mobile Agent
func (m *MobileAgentWrapper) Invoke(ctx context.Context, message string, maxIterations int) InvokeResult {
	log.Printf("Mobile Agent 正在处理 Android 任务: %s", message)

	response := m.agent.Invoke(ctx, message, agent.WithMaxIterationNum(maxIterations))
	if response.Error != nil {
		return InvokeResult{
			Content: "",
			Error:   response.Error,
		}
	}

	return InvokeResult{
		Content: response.Completion(),
		Error:   nil,
	}
}

// generateIntelligentResponse 生成智能回复
func generateIntelligentResponse(message string) string {
	message = strings.ToLower(message)

	if strings.Contains(message, "购物") || strings.Contains(message, "买") || strings.Contains(message, "购买") {
		if strings.Contains(message, "儿童自行车") || strings.Contains(message, "自行车") {
			return `基于您要为3岁小孩购买儿童自行车的需求，我为您进行了详细分析：

**需求分析要点：**
1. 年龄适配：3岁儿童身高约90-100cm
2. 安全性：稳定性和保护措施最重要
3. 学习性：有助于平衡感发展

**推荐建议：**
1. **12寸轮径**最适合3岁儿童
2. **建议选择平衡车或带辅助轮的自行车**
3. **重要安全特性：**
   - 低车架设计，方便上下车
   - 防滑手把和脚踏
   - 可调节座椅高度
   - 配备护膝护肘等防护用具

**品牌推荐：**
- 迪卡侬：性价比高，安全可靠
- 永久儿童：经典品牌，质量稳定
- 好孩子：专业儿童用品，设计贴心

**购买建议：**
- 价格区间：200-500元为宜
- 优先关注用户真实评价中的安全性反馈
- 建议购买时让孩子试骑，确保舒适度

希望这些建议对您有帮助！`
		}
		return `作为您的购物专家助手，我已经为您分析了购物需求。建议您：

1. **明确需求**：确定具体的产品规格和预算范围
2. **多平台比较**：建议对比淘宝、京东、拼多多等平台价格
3. **查看评价**：重点关注真实用户的使用反馈
4. **考虑售后**：选择有良好售后服务的商家

请告诉我您想购买什么具体产品，我可以为您提供更详细的分析建议。`
	} else if strings.Contains(message, "天气") {
		return fmt.Sprintf(`当前时间：%s

很抱歉，我暂时无法获取实时天气数据。建议您：
1. 查看手机自带的天气应用
2. 访问中国天气网或天气通等专业天气网站
3. 关注当地气象台发布的天气预报

如果您需要特定城市的天气信息，我可以为您提供查询建议。`, time.Now().Format("2006-01-02 15:04:05"))
	} else if strings.Contains(message, "编程") || strings.Contains(message, "代码") {
		return `作为您的编程助手，我可以帮助您：

**技术支持范围：**
- 代码编写和优化
- 问题调试和解决
- 架构设计建议
- 最佳实践指导

**支持的编程语言：**
- Go, Python, JavaScript, Java
- C/C++, Rust, TypeScript
- 前端框架：React, Vue, Angular
- 后端框架：Gin, Django, Express

请详细描述您的具体需求，我会为您提供针对性的技术方案。`
	} else if strings.Contains(message, "你好") || strings.Contains(message, "hello") {
		return `您好！我是AI智能助手 🤖

我可以为您提供以下服务：
- 🛒 **购物咨询**：产品分析、比价建议、用户评价解读
- 💻 **编程协助**：代码编写、调试优化、技术方案
- 🔍 **信息搜索**：资料查找、问题解答、知识科普
- 📊 **数据分析**：数据处理、图表生成、趋势分析

我会通过多个步骤来理解您的需求并提供专业建议。请告诉我您需要什么帮助！`
	} else {
		return fmt.Sprintf(`我已收到您的消息：「%s」

作为AI智能助手，我正在分析您的需求：
- ✅ 消息理解完成
- ⚡ 任务类型识别
- 🎯 方案制定中

请稍等，我正在为您准备最合适的解决方案...`, message)
	}
}

// Session 会话管理
type Session struct {
	ID          string             `json:"id"`
	CreatedAt   time.Time          `json:"created_at"`
	Messages    []Message          `json:"messages"`
	Status      string             `json:"status"`
	CurrentStep int                `json:"current_step"`
	Steps       []ExecutionStep    `json:"steps"`
	Context     context.Context    `json:"-"`
	Cancel      context.CancelFunc `json:"-"`
	Mutex       sync.RWMutex       `json:"-"`
	Agent       AgentInterface     `json:"-"`
	AgentType   string             `json:"agent_type"`
}

// Message 消息结构
type Message struct {
	ID        string    `json:"id"`
	Type      string    `json:"type"` // user, agent, system, thinking
	Content   string    `json:"content"`
	Timestamp time.Time `json:"timestamp"`
	Status    string    `json:"status,omitempty"`
	StepID    string    `json:"step_id,omitempty"`
}

// 全局会话管理
var (
	sessions     = make(map[string]*Session)
	sessionMutex sync.RWMutex
)

// enableCORS 启用 CORS
func enableCORS(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
}

// ChatRequest 数据结构定义
type ChatRequest struct {
	Message   string `json:"message"`
	SessionID string `json:"session_id"`
	Stream    bool   `json:"stream"`
}

// ChatResponse TODO
type ChatResponse struct {
	SessionID string          `json:"session_id"`
	Message   string          `json:"message"`
	Status    string          `json:"status"`
	Steps     []ExecutionStep `json:"steps,omitempty"`
	Timestamp time.Time       `json:"timestamp"`
}

// ExecutionStep TODO
type ExecutionStep struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	Status    string    `json:"status"` // pending, running, completed, failed
	Details   string    `json:"details,omitempty"`
	Progress  int       `json:"progress,omitempty"` // 0-100
	Timestamp time.Time `json:"timestamp"`
	Duration  int64     `json:"duration,omitempty"` // 毫秒
}

// StreamMessage TODO
type StreamMessage struct {
	Type      string          `json:"type"`
	SessionID string          `json:"session_id"`
	Content   string          `json:"content,omitempty"`
	Steps     []ExecutionStep `json:"steps,omitempty"`
	Status    string          `json:"status,omitempty"`
	Error     string          `json:"error,omitempty"`
	Timestamp time.Time       `json:"timestamp"`
}

// isAndroidTask 检测是否为Android相关任务
func isAndroidTask(message string) bool {
	message = strings.ToLower(message)
	androidKeywords := []string{
		"@android", "android", "手机", "mobile", "app", "应用",
		"点击", "滑动", "截屏", "安装", "启动应用", "手机操作",
		"淘宝", "京东", "小红书", "微信", "抖音", "支付宝",
		"购物", "下单", "浏览", "搜索商品", "打开应用",
	}

	for _, keyword := range androidKeywords {
		if strings.Contains(message, keyword) {
			return true
		}
	}
	return false
}

// getOrCreateSession 获取或创建会话
func getOrCreateSession(sessionID string) *Session {
	sessionMutex.Lock()
	defer sessionMutex.Unlock()

	if sessionID == "" {
		sessionID = fmt.Sprintf("session-%d", time.Now().Unix())
	}

	session, exists := sessions[sessionID]
	if !exists {
		ctx, cancel := context.WithCancel(context.Background())

		session = &Session{
			ID:        sessionID,
			CreatedAt: time.Now(),
			Messages:  []Message{},
			Status:    "active",
			Context:   ctx,
			Cancel:    cancel,
			Steps:     []ExecutionStep{},
			AgentType: "ai", // 默认为AI Agent
		}
		sessions[sessionID] = session
	}

	return session
}

// createAgentForTask 根据任务创建相应的Agent
func createAgentForTask(message string) (AgentInterface, string) {
	if isAndroidTask(message) {
		log.Printf("检测到Android任务，创建MobileAgent")
		return NewMobileAgentWrapper(), "android"
	}

	log.Printf("创建默认AI Agent")
	return NewAIAgent("AI智能助手", "claude-3-sonnet"), "ai"
}

// analyzeTaskType 分析任务类型并返回相应的执行步骤
func analyzeTaskType(message string) []ExecutionStep {
	message = strings.ToLower(message)
	baseTime := time.Now()

	// Android相关任务的步骤
	if isAndroidTask(message) {
		return []ExecutionStep{
			{ID: "1", Title: "📱 初始化Android设备连接", Status: "pending", Timestamp: baseTime},
			{ID: "2", Title: "📋 分析任务需求", Status: "pending", Timestamp: baseTime},
			{ID: "3", Title: "🔍 获取当前屏幕状态", Status: "pending", Timestamp: baseTime},
			{ID: "4", Title: "🎯 定位目标元素", Status: "pending", Timestamp: baseTime},
			{ID: "5", Title: "⚡ 执行Android操作", Status: "pending", Timestamp: baseTime},
			{ID: "6", Title: "✅ 验证操作结果", Status: "pending", Timestamp: baseTime},
			{ID: "7", Title: "📤 生成任务报告", Status: "pending", Timestamp: baseTime},
		}
	}

	if strings.Contains(message, "购物") || strings.Contains(message, "买") || strings.Contains(message, "购买") {
		return []ExecutionStep{
			{ID: "1", Title: "🧠 理解购物需求", Status: "pending", Timestamp: baseTime},
			{ID: "2", Title: "🔍 分析关键需求点", Status: "pending", Timestamp: baseTime},
			{ID: "3", Title: "📱 启动购物应用分析", Status: "pending", Timestamp: baseTime},
			{ID: "4", Title: "🛒 搜索商品信息", Status: "pending", Timestamp: baseTime},
			{ID: "5", Title: "⭐ 分析用户评价", Status: "pending", Timestamp: baseTime},
			{ID: "6", Title: "📊 比较商品选项", Status: "pending", Timestamp: baseTime},
			{ID: "7", Title: "💡 提供购买建议", Status: "pending", Timestamp: baseTime},
		}
	} else if strings.Contains(message, "搜索") || strings.Contains(message, "查找") {
		return []ExecutionStep{
			{ID: "1", Title: "🧠 理解搜索意图", Status: "pending", Timestamp: baseTime},
			{ID: "2", Title: "🔍 制定搜索策略", Status: "pending", Timestamp: baseTime},
			{ID: "3", Title: "🌐 执行信息搜索", Status: "pending", Timestamp: baseTime},
			{ID: "4", Title: "📊 筛选相关结果", Status: "pending", Timestamp: baseTime},
			{ID: "5", Title: "💡 整理提供答案", Status: "pending", Timestamp: baseTime},
		}
	} else {
		return []ExecutionStep{
			{ID: "1", Title: "🧠 理解用户需求", Status: "pending", Timestamp: baseTime},
			{ID: "2", Title: "🔍 分析任务类型", Status: "pending", Timestamp: baseTime},
			{ID: "3", Title: "📋 制定执行计划", Status: "pending", Timestamp: baseTime},
			{ID: "4", Title: "⚡ 执行核心任务", Status: "pending", Timestamp: baseTime},
			{ID: "5", Title: "✅ 验证结果质量", Status: "pending", Timestamp: baseTime},
			{ID: "6", Title: "📤 生成最终回复", Status: "pending", Timestamp: baseTime},
		}
	}
}

// updateStepStatus 更新步骤状态
func updateStepStatus(session *Session, stepID string, status string, details string, progress int) {
	session.Mutex.Lock()
	defer session.Mutex.Unlock()

	for i := range session.Steps {
		if session.Steps[i].ID == stepID {
			oldStatus := session.Steps[i].Status
			session.Steps[i].Status = status
			session.Steps[i].Details = details
			session.Steps[i].Progress = progress

			// 记录开始时间
			if oldStatus == "pending" && status == "running" {
				session.Steps[i].Timestamp = time.Now()
			}

			// 计算执行时间
			if status == "completed" || status == "failed" {
				session.Steps[i].Duration = time.Since(session.Steps[i].Timestamp).Milliseconds()
			}
			break
		}
	}
}

// addStepMessage 添加步骤相关的消息
func addStepMessage(session *Session, stepID string, msgType string, content string) {
	session.Mutex.Lock()
	defer session.Mutex.Unlock()

	msg := Message{
		ID:        fmt.Sprintf("msg-%d", time.Now().UnixNano()),
		Type:      msgType,
		Content:   content,
		Timestamp: time.Now(),
		StepID:    stepID,
	}
	session.Messages = append(session.Messages, msg)
}

// StepResult 步骤执行结果
type StepResult struct {
	Success bool
	Message string
	Details string
	Error   string
}

// executeStep 执行具体步骤
func executeStep(session *Session, step ExecutionStep, originalMessage string, stepIndex int) StepResult {
	// 根据步骤类型执行不同的逻辑
	switch {
	case strings.Contains(step.Title, "理解") || strings.Contains(step.Title, "分析"):
		// 模拟分析时间，分阶段更新进度
		updateStepStatus(session, step.ID, "running", "正在理解需求...", 25)
		time.Sleep(400 * time.Millisecond)
		updateStepStatus(session, step.ID, "running", "分析关键要点...", 50)
		time.Sleep(400 * time.Millisecond)
		updateStepStatus(session, step.ID, "running", "整理分析结果...", 75)
		time.Sleep(200 * time.Millisecond)
		return StepResult{
			Success: true,
			Message: "需求分析完成，已识别关键要点",
			Details: "成功解析用户意图和需求要点",
		}
	case strings.Contains(step.Title, "搜索") || strings.Contains(step.Title, "查找"):
		// 模拟搜索时间
		updateStepStatus(session, step.ID, "running", "准备搜索策略...", 20)
		time.Sleep(300 * time.Millisecond)
		updateStepStatus(session, step.ID, "running", "执行信息搜索...", 60)
		time.Sleep(600 * time.Millisecond)
		updateStepStatus(session, step.ID, "running", "处理搜索结果...", 90)
		time.Sleep(300 * time.Millisecond)
		return StepResult{
			Success: true,
			Message: "搜索操作完成，找到相关信息",
			Details: "已获取相关数据和信息",
		}
	case strings.Contains(step.Title, "启动") || strings.Contains(step.Title, "打开") || strings.Contains(step.Title, "初始化"):
		// 模拟启动时间
		updateStepStatus(session, step.ID, "running", "正在启动应用...", 30)
		time.Sleep(300 * time.Millisecond)
		updateStepStatus(session, step.ID, "running", "等待应用响应...", 70)
		time.Sleep(300 * time.Millisecond)
		return StepResult{
			Success: true,
			Message: "应用启动成功",
			Details: "相关应用或服务已准备就绪",
		}
	case strings.Contains(step.Title, "比较") || strings.Contains(step.Title, "评价"):
		// 模拟比较分析时间
		updateStepStatus(session, step.ID, "running", "收集比较数据...", 25)
		time.Sleep(400 * time.Millisecond)
		updateStepStatus(session, step.ID, "running", "执行对比分析...", 60)
		time.Sleep(700 * time.Millisecond)
		updateStepStatus(session, step.ID, "running", "生成对比结果...", 85)
		time.Sleep(400 * time.Millisecond)
		return StepResult{
			Success: true,
			Message: "比较分析完成",
			Details: "已完成多维度对比分析",
		}
	case strings.Contains(step.Title, "生成") || strings.Contains(step.Title, "提供"):
		// 模拟生成时间
		updateStepStatus(session, step.ID, "running", "准备生成内容...", 30)
		time.Sleep(400 * time.Millisecond)
		updateStepStatus(session, step.ID, "running", "生成结果中...", 70)
		time.Sleep(600 * time.Millisecond)
		return StepResult{
			Success: true,
			Message: "结果生成完成",
			Details: "已准备最终建议和答案",
		}
	case strings.Contains(step.Title, "Android") || strings.Contains(step.Title, "设备"):
		// Android相关步骤
		updateStepStatus(session, step.ID, "running", "连接Android设备...", 30)
		time.Sleep(500 * time.Millisecond)
		updateStepStatus(session, step.ID, "running", "验证设备状态...", 70)
		time.Sleep(300 * time.Millisecond)
		return StepResult{
			Success: true,
			Message: "Android设备连接成功",
			Details: "设备已就绪，可以执行操作",
		}
	case strings.Contains(step.Title, "屏幕") || strings.Contains(step.Title, "截屏"):
		// 屏幕相关操作
		updateStepStatus(session, step.ID, "running", "获取屏幕信息...", 40)
		time.Sleep(400 * time.Millisecond)
		updateStepStatus(session, step.ID, "running", "分析屏幕内容...", 80)
		time.Sleep(200 * time.Millisecond)
		return StepResult{
			Success: true,
			Message: "屏幕状态获取完成",
			Details: "已获取当前屏幕信息",
		}
	case strings.Contains(step.Title, "元素") || strings.Contains(step.Title, "定位"):
		// UI元素定位
		updateStepStatus(session, step.ID, "running", "扫描界面元素...", 35)
		time.Sleep(400 * time.Millisecond)
		updateStepStatus(session, step.ID, "running", "定位目标元素...", 75)
		time.Sleep(300 * time.Millisecond)
		return StepResult{
			Success: true,
			Message: "界面元素定位完成",
			Details: "已找到可操作的界面元素",
		}
	default:
		// 默认处理时间
		updateStepStatus(session, step.ID, "running", "正在处理...", 40)
		time.Sleep(time.Duration(300+stepIndex*100) * time.Millisecond)
		updateStepStatus(session, step.ID, "running", "即将完成...", 80)
		time.Sleep(200 * time.Millisecond)
		return StepResult{
			Success: true,
			Message: fmt.Sprintf("步骤 %d 执行完成", stepIndex+1),
			Details: "步骤处理成功",
		}
	}
}

func processAIRequest(session *Session, message string) {
	mobileAgent := agent.NewMobileAgent("mobile use agent",
		agent.WithModel(claudeModel),
	)
	response := mobileAgent.Invoke(session.Context, message, agent.WithMaxIterationNum(10))
	for chunk := range response.Stream {
		session.Mutex.Lock()
		session.Messages = append(session.Messages, Message{
			ID:        fmt.Sprintf("msg-%d", time.Now().UnixNano()),
			Type:      "agent",
			Content:   chunk.Choices[0].Delta.Content(),
			Timestamp: time.Now(),
		})
		session.Mutex.Unlock()
	}
	session.Mutex.Lock()
	session.Status = "completed"
	session.Mutex.Unlock()
}

// processAIRequest2 处理AI请求
func processAIRequest2(session *Session, message string) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("AI处理出错: %v", r)
			session.Mutex.Lock()
			session.Status = "error"
			session.Mutex.Unlock()

			// 更新当前步骤为失败状态
			if session.CurrentStep < len(session.Steps) {
				updateStepStatus(session, fmt.Sprintf("%d", session.CurrentStep+1), "failed",
					fmt.Sprintf("处理过程中发生错误: %v", r), 0)
			}
		}
	}()

	// 根据任务创建相应的Agent
	if session.Agent == nil {
		agent, agentType := createAgentForTask(message)
		session.Mutex.Lock()
		session.Agent = agent
		session.AgentType = agentType
		session.Mutex.Unlock()
		log.Printf("创建了 %s 类型的Agent", agentType)
	}

	// 分析任务类型并设置执行步骤
	steps := analyzeTaskType(message)
	session.Mutex.Lock()
	session.Steps = steps
	session.Status = "thinking"
	session.CurrentStep = 0
	session.Mutex.Unlock()

	log.Printf("开始处理AI请求，共%d个步骤 (Agent类型: %s)", len(steps), session.AgentType)

	// 逐步执行任务
	for i, step := range steps {
		session.Mutex.Lock()
		session.CurrentStep = i
		session.Mutex.Unlock()

		// 更新步骤状态为运行中
		updateStepStatus(session, step.ID, "running", "正在执行...", 0)
		addStepMessage(session, step.ID, "system", fmt.Sprintf("开始执行: %s", step.Title))

		// 执行具体步骤（内部已包含进度更新）
		stepResult := executeStep(session, step, message, i)

		if stepResult.Success {
			updateStepStatus(session, step.ID, "completed", stepResult.Details, 100)
			addStepMessage(session, step.ID, "agent", stepResult.Message)
			log.Printf("步骤 %d 完成: %s", i+1, step.Title)
		} else {
			updateStepStatus(session, step.ID, "failed", stepResult.Error, 0)
			addStepMessage(session, step.ID, "system", fmt.Sprintf("步骤失败: %s", stepResult.Error))

			session.Mutex.Lock()
			session.Status = "error"
			session.Mutex.Unlock()
			return
		}

		// 短暂停顿让用户看到步骤完成
		time.Sleep(200 * time.Millisecond)
	}

	// 所有步骤完成后，调用 Agent 生成最终回复
	log.Printf("开始调用 %s Agent 生成最终回复", session.AgentType)
	session.Mutex.Lock()
	session.Status = "generating"
	session.Mutex.Unlock()

	// 根据Agent类型调用相应的处理方法
	var maxIterations int
	if session.AgentType == "android" {
		maxIterations = 10 // Android任务可能需要更多迭代
	} else {
		maxIterations = 3
	}

	// 使用 Agent 处理请求
	response := session.Agent.Invoke(session.Context, message, maxIterations)

	session.Mutex.Lock()
	defer session.Mutex.Unlock()

	if response.Error != nil {
		session.Status = "error"
		log.Printf("Agent处理失败: %v", response.Error)

		// 添加错误消息
		errorMsg := Message{
			ID:        fmt.Sprintf("msg-%d", time.Now().UnixNano()),
			Type:      "system",
			Content:   fmt.Sprintf("AI处理失败: %v", response.Error),
			Timestamp: time.Now(),
			Status:    "error",
		}
		session.Messages = append(session.Messages, errorMsg)
	} else {
		session.Status = "completed"

		// 添加AI响应消息
		agentMsg := Message{
			ID:        fmt.Sprintf("msg-%d", time.Now().UnixNano()),
			Type:      "agent",
			Content:   response.Content,
			Timestamp: time.Now(),
			Status:    "completed",
		}
		session.Messages = append(session.Messages, agentMsg)

		log.Printf("Agent处理完成: 会话 %s (%s类型)", session.ID, session.AgentType)
	}
}

// healthHandler 健康检查处理器
func healthHandler(w http.ResponseWriter, r *http.Request) {
	enableCORS(w)
	if r.Method == "OPTIONS" {
		return
	}

	w.Header().Set("Content-Type", "application/json")
	response := map[string]interface{}{
		"status":    "ok",
		"timestamp": time.Now().Unix(),
		"service":   "AI Agent Platform Server",
		"version":   "v2.1.0",
		"features": []string{
			"智能任务分析",
			"步骤化处理",
			"实时状态跟踪",
			"多类型任务支持",
			"Android设备控制 (@android)",
			"移动应用自动化",
		},
		"supported_agents": []string{
			"AI智能助手",
			"Android移动设备代理",
		},
	}
	json.NewEncoder(w).Encode(response)
}

// chatHandler 聊天处理器
func chatHandler(w http.ResponseWriter, r *http.Request) {
	enableCORS(w)
	if r.Method == "OPTIONS" {
		return
	}

	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	var req ChatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// 验证请求
	if req.Message == "" {
		http.Error(w, "Message is required", http.StatusBadRequest)
		return
	}

	// 获取或创建会话
	session := getOrCreateSession(req.SessionID)

	// 添加用户消息
	session.Mutex.Lock()
	userMsg := Message{
		ID:        fmt.Sprintf("msg-%d", time.Now().UnixNano()),
		Type:      "user",
		Content:   req.Message,
		Timestamp: time.Now(),
	}
	session.Messages = append(session.Messages, userMsg)
	session.Mutex.Unlock()

	// 分析任务并预设步骤
	steps := analyzeTaskType(req.Message)

	// 构建响应
	response := ChatResponse{
		SessionID: session.ID,
		Message:   "正在分析您的需求，准备执行任务...",
		Status:    "started",
		Timestamp: time.Now(),
		Steps:     steps,
	}

	log.Printf("处理聊天请求: %s (会话: %s, 步骤数: %d)", req.Message, session.ID, len(steps))

	// 异步处理AI请求
	go processAIRequest(session, req.Message)

	json.NewEncoder(w).Encode(response)
}

// streamHandler 流式处理器
func streamHandler(w http.ResponseWriter, r *http.Request) {
	enableCORS(w)
	if r.Method == "OPTIONS" {
		return
	}

	sessionID := r.URL.Path[len("/api/v1/chat/stream/"):]
	if sessionID == "" {
		http.Error(w, "Session ID required", http.StatusBadRequest)
		return
	}

	sessionMutex.RLock()
	session, exists := sessions[sessionID]
	sessionMutex.RUnlock()

	if !exists {
		http.Error(w, "Session not found", http.StatusNotFound)
		return
	}

	log.Printf("启动流式处理: %s", sessionID)

	// 设置 SSE 头部
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming not supported", http.StatusInternalServerError)
		return
	}

	// 定期发送状态更新
	ticker := time.NewTicker(200 * time.Millisecond)
	defer ticker.Stop()

	// 跟踪上次发送的状态，避免重复发送相同内容
	var lastSentStatus string
	var lastSentStepCount int
	var lastSentCurrentStep int
	lastStepProgress := make(map[string]int)

	for {
		select {
		case <-ticker.C:
			session.Mutex.RLock()
			status := session.Status
			currentStep := session.CurrentStep
			steps := make([]ExecutionStep, len(session.Steps))
			copy(steps, session.Steps)

			lastMessage := ""
			if len(session.Messages) > 0 {
				for i := len(session.Messages) - 1; i >= 0; i-- {
					if session.Messages[i].Type == "agent" {
						lastMessage = session.Messages[i].Content
						break
					}
				}
			}
			session.Mutex.RUnlock()

			// 检查是否有实质性变化
			hasChange := false
			if status != lastSentStatus || currentStep != lastSentCurrentStep || len(steps) != lastSentStepCount {
				hasChange = true
			}

			// 检查步骤进度变化
			for _, step := range steps {
				if lastProgress, exists := lastStepProgress[step.ID]; !exists || lastProgress != step.Progress {
					lastStepProgress[step.ID] = step.Progress
					hasChange = true
				}
			}

			// 只有在有变化时才发送更新
			if hasChange {
				// 发送状态更新
				msg := StreamMessage{
					Type:      "status_update",
					SessionID: sessionID,
					Status:    status,
					Content:   lastMessage,
					Steps:     steps,
					Timestamp: time.Now(),
				}

				data, err := json.Marshal(msg)
				if err != nil {
					log.Printf("序列化消息失败: %v", err)
					continue
				}

				// 使用正确的SSE格式
				fmt.Fprintf(w, "data: %s\n\n", data)
				flusher.Flush()

				// 发送步骤进度信息
				if currentStep < len(steps) && currentStep >= 0 {
					currentStepInfo := steps[currentStep]
					stepMsg := StreamMessage{
						Type:      "step_update",
						SessionID: sessionID,
						Status:    fmt.Sprintf("step_%d", currentStep+1),
						Content: fmt.Sprintf("正在执行第 %d/%d 步: %s (进度: %d%%)",
							currentStep+1, len(steps), currentStepInfo.Title, currentStepInfo.Progress),
						Timestamp: time.Now(),
					}

					stepData, err := json.Marshal(stepMsg)
					if err == nil {
						fmt.Fprintf(w, "data: %s\n\n", stepData)
						flusher.Flush()
					}
				}

				// 更新最后发送的状态
				lastSentStatus = status
				lastSentCurrentStep = currentStep
				lastSentStepCount = len(steps)

				log.Printf("发送状态更新: %s (状态: %s, 步骤: %d/%d)", sessionID, status, currentStep+1, len(steps))
			}

			if status == "completed" || status == "error" {
				log.Printf("完成流式处理: %s (状态: %s)", sessionID, status)
				return
			}

		case <-r.Context().Done():
			log.Printf("客户端断开连接: %s", sessionID)
			return
		}
	}
}

// sessionsHandler 会话管理处理器
func sessionsHandler(w http.ResponseWriter, r *http.Request) {
	enableCORS(w)
	if r.Method == "OPTIONS" {
		return
	}

	w.Header().Set("Content-Type", "application/json")

	switch r.Method {
	case "GET":
		sessionMutex.RLock()
		sessionList := make([]map[string]interface{}, 0, len(sessions))
		for _, session := range sessions {
			session.Mutex.RLock()
			sessionInfo := map[string]interface{}{
				"id":           session.ID,
				"created_at":   session.CreatedAt,
				"status":       session.Status,
				"messages":     len(session.Messages),
				"current_step": session.CurrentStep,
				"total_steps":  len(session.Steps),
			}
			session.Mutex.RUnlock()
			sessionList = append(sessionList, sessionInfo)
		}
		sessionMutex.RUnlock()

		json.NewEncoder(w).Encode(map[string]interface{}{
			"sessions": sessionList,
			"total":    len(sessionList),
		})
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func main() {
	// 设置路由
	http.HandleFunc("/health", healthHandler)
	http.HandleFunc("/api/v1/chat", chatHandler)
	http.HandleFunc("/api/v1/chat/stream/", streamHandler)
	http.HandleFunc("/api/v1/sessions", sessionsHandler)

	// 静态文件服务（如果需要）
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))

	port := ":8080"
	log.Printf("🚀 AI Agent Platform Server v2.0 启动在端口 %s", port)
	log.Printf("📊 健康检查: http://localhost%s/health", port)
	log.Printf("💬 聊天API: http://localhost%s/api/v1/chat", port)
	log.Printf("📡 流式API: http://localhost%s/api/v1/chat/stream/{sessionId}", port)
	log.Printf("📋 会话管理: http://localhost%s/api/v1/sessions", port)
	log.Printf("🧠 智能AI助手已就绪，支持：")
	log.Printf("   - 🛒 购物咨询与建议")
	log.Printf("   - 💻 编程协助与调试")
	log.Printf("   - 🔍 信息搜索与分析")
	log.Printf("   - 📊 多步骤任务处理")

	if err := http.ListenAndServe(port, nil); err != nil {
		log.Fatalf("服务器启动失败: %v", err)
	}
}

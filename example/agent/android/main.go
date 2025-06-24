package main

import (
	"context"
	"fmt"
	"os"

	"github.com/joho/godotenv"
	"github.com/showntop/llmack/agent"
	"github.com/showntop/llmack/llm"

	openaic "github.com/showntop/llmack/llm/openai-c"
	"github.com/showntop/llmack/log"
)

func init() {
	godotenv.Load()

	log.SetLogger(&log.WrapLogger{})

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

	// llm.WithSingleConfig(map[string]any{
	// 	"base_url": "https://ark.cn-beijing.volces.com/api/v3",
	// 	"api_key":  os.Getenv("doubao_api_key"),
	// 	// "base_url": "http://v2.open.venus.oa.com/llmproxy",
	// 	// "api_key":  os.Getenv("claude_api_key"),
	// })

	// llm.WithSingleConfig(map[string]any{
	// 	"api_key": os.Getenv("qwen_api_key"),
	// })
	// llm.WithSingleConfig(map[string]any{
	// 	"api_key": os.Getenv("deepseek_api_key"),
	// })
}

// initAndroid 初始化安卓手机
func initAndroid() {
	// adb install droidrun-portable.apk
	// adb install ADBKeyboard.apk
	// adb shell ime enable com.android.adbkeyboard/.AdbIME
	// adb shell ime set com.android.adbkeyboard/.AdbIME
}

var claudeModel = llm.NewInstance(openaic.Name,
	llm.WithDefaultModel("claude-3-7-sonnet-20250219"),
	llm.WithInvokeOptions(&llm.InvokeOptions{
		Metadata: map[string]any{
			"thinking_enabled": "true",
			"thinking_tokens":  2048,
		},
	}), // default invoke options
)

func main() {

	androidAgent := agent.NewMobileAgent("mobile use agent",
		// agent.WithModel(llm.NewInstance(doubao.Name, llm.WithDefaultModel("doubao-1-5-ui-tars-250428"))),
		agent.WithModel(claudeModel),
		// agent.WithModel(llm.NewInstance(qwen.Name, llm.WithDefaultModel("qwen-vl-max-latest"))),
		// agent.WithModel(llm.NewInstance(deepseek.Name, llm.WithDefaultModel("deepseek-chat"))),
	)

	task := `我想购买一辆儿童自行车供3岁小孩使用。
	1. 仔细思考需求，并列出需要考虑的点，综合考虑后再决策。
	2. 你可以浏览器及小红书、淘宝、京东等APP，帮助我决策购物。
	3. 你是一个经验丰富的购物专家，可以帮我分析商品，给出建议。
	4. 你尤其需要关注真实用户的评论。
	`

	response := androidAgent.Invoke(context.Background(), task, agent.WithMaxIterationNum(10))
	// response := androidAgent.Invoke(context.Background(), "在搜索框输入：抖音", agent.WithMaxIterationNum(10))
	if response.Error != nil {
		panic(response.Error)
	}
	fmt.Println(response.Completion())
}

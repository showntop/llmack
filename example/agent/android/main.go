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
	"github.com/showntop/llmack/storage"
	"github.com/showntop/llmack/tool"
)

func init() {
	godotenv.Load()

	log.SetLogger(&mlog{})

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
			// "thinking_enabled": "true",
			// "thinking_tokens": 2048,
		},
	}), // default invoke options
)

func main() {

	androidAgent := agent.NewMobileAgent("mobile use agent", "118.31.173.101:100",
		// agent.WithModel(llm.NewInstance(doubao.Name, llm.WithDefaultModel("doubao-1-5-ui-tars-250428"))),
		agent.WithModel(claudeModel),
		// agent.WithModel(llm.NewInstance(qwen.Name, llm.WithDefaultModel("qwen-vl-max-latest"))),
		// agent.WithModel(llm.NewInstance(deepseek.Name, llm.WithDefaultModel("deepseek-chat"))),
		agent.WithDescription("你是一个擅长使用手机采集数据的专家，你的任务是采集抖音上的商家信息。"),
		agent.WithTools(saveLeadsTool()),
		agent.WithStorage(storage.NewJSONStorage("leads_agent")),
	)
	// 1. 打开美团 2. 导航至抖音首页 3. 进入"团购"频道 4. 点击"丽人美发"按钮 5. 点击地区筛选选择北京市-朝阳区-三里屯商圈 6. 进入一个三里屯商圈商家 7. 截图获取商家信息（名称、地址、主营类目、门店性质、是否开设直播、团购数量、评分、评论数） 8. 点击右上角的"..."按钮 9. 点击"商家资质"按钮 10. 查看并截图商家的营业执照信息 11. 将收集到的信息存储在数据库中 12.返回到古城商圈的商家列表页 13.选择下一家商家继续到第7步的任务采集下一家商家的商家信息和商家资质

	response := androidAgent.Invoke(
		context.Background(),
		task,
		agent.WithMaxIterationNum(10),
		agent.WithSessionID("11009876"),
	)
	// response := androidAgent.Invoke(context.Background(), "在搜索框输入：抖音", agent.WithMaxIterationNum(10))
	if response.Error != nil {
		panic(response.Error)
	}
	fmt.Println(response.Completion())
}

func saveLeadsTool() string {
	// open or create
	ff, err := os.Create("leads.jsonl")
	if err != nil {
		panic(err)
	}
	saveLeads := func(ctx context.Context, args string) (string, error) {
		fmt.Println("lead: ", args)
		_, err := ff.WriteString(args)
		return "", err
	}

	tx := tool.New(
		tool.WithName("save_leads"),
		tool.WithDescription("save leads to database"),
		tool.WithParameters(tool.Parameter{
			Name:          "leads",
			Type:          "string",
			LLMDescrition: "leads",
		}),
		tool.WithFunction(saveLeads),
	)

	tool.Register(tx)
	return tx.Name
}

// 3. 在团购页面中点击搜索框最左侧的按钮，切换城市为北京
var task = `
你的任务流程如下： 
1. 根据给你的截图确认当前是否在[抖音-团购-丽人/美发]页然后在执行 2，如果不在请按序操作以下不走
	- 打开抖音；
	- 点击团购tab；
	- 点击搜索框最左侧的按钮（注意应点击团购页里面的城市切换，而非tab上的城市切换），切换城市为北京；
	- 进入"丽人/美发"频道
2. 筛选三里屯商圈的商家
3. 提取商家信息（名称、地址、主营类目、门店性质、是否开设直播、团购数量、评分、评论数） 
4. 点击右上角的"..."按钮 
5. 点击"商家资质"按钮 
6. 查看并截图商家的营业执照信息 
7. 保存收集到的线索信息

注意：进入团购页后所有操作一定要在团购页操作，不要在其他页面操作，否则会失败
`

type mlog struct{}

func (m mlog) DebugContextf(ctx context.Context, format string, args ...any) {
	fmt.Printf(format, args...)
}

func (m mlog) Debugf(format string, args ...any) {
	fmt.Printf(format, args...)
}

func (m mlog) ErrorContextf(ctx context.Context, format string, args ...any) {
	fmt.Printf(format, args...)
}

func (m mlog) Errorf(format string, args ...any) {
	fmt.Printf(format, args...)
}

func (m mlog) InfoContextf(ctx context.Context, format string, args ...any) {
	fmt.Printf(format, args...)
	fmt.Println()
}

func (m mlog) Infof(format string, args ...any) {
	fmt.Printf(format, args...)
	fmt.Println()
}

func (m mlog) WarnContextf(ctx context.Context, format string, args ...any) {
	fmt.Printf(format, args...)
}

func (m mlog) Warnf(format string, args ...any) {
	fmt.Printf(format, args...)
}

func (m mlog) FatalContextf(ctx context.Context, format string, args ...any) {
	fmt.Printf(format, args...)
}

func (m mlog) Fatalf(format string, args ...any) {
	fmt.Printf(format, args...)
}

func (m mlog) PanicContextf(ctx context.Context, format string, args ...any) {
	fmt.Printf(format, args...)
}

func (m mlog) Panicf(format string, args ...any) {
	fmt.Printf(format, args...)
}

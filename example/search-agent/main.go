package main

import (
	"context"
	"fmt"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"go.uber.org/fx"

	"github.com/showntop/llmack/engine"
	"github.com/showntop/llmack/llm"
	"github.com/showntop/llmack/log"
	"github.com/showntop/llmack/tool"

	"github.com/showntop/llmack/example/search-agent/app"
	"github.com/showntop/llmack/example/search-agent/inter"
	"github.com/showntop/llmack/example/search-agent/workflow"
)

func init() {
	log.SetLogger(&log.WrapLogger{})

	godotenv.Load()
	llm.WithSingleConfig(map[string]any{
		// "api_key":  os.Getenv("deepseek_api_key2"),
		// "base_url": "https://api.lkeap.cloud.tencent.com/v1",
		"api_key": os.Getenv("qwen_api_key"),
	})

	tool.WithConfig(map[string]any{
		"serper": map[string]any{
			"api_key": os.Getenv("serper_api_key"),
		},
	})
}

func main() {
	serve()
	// cmd()
}

func serve() {
	var module = fx.Module("search",
		fx.Provide(app.NewApplication),
		fx.Provide(inter.NewSSEHandler),
	)
	fx.New(
		module,
		fx.Invoke(startGinServer),
	).Run()
}

// startGinServer 创建 gRPC 网关
func startGinServer() error {
	// new gin server with cors
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(gin.Logger())
	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})
	register(r)
	return r.Run(":8000")
}

func register(r *gin.Engine) {
	// 注册SSE
	handler := inter.SSEHandler{}
	r.POST("/api/search", handler.Search)
}

func cmd() {
	settings := engine.DefaultSettings()
	settings.Workflow = workflow.BuildWorkflow()
	eng := engine.NewWorkflowEngine(settings, engine.WithLogger(&log.WrapLogger{}))
	esm := eng.Execute(context.Background(), engine.Input{
		Query: "如何使用 AI 制作绘本",
		// Query: "你是谁？",
	})
	for evt := esm.Next(); evt != nil; evt = esm.Next() {
		if evt.Error != nil {
			panic(evt.Error)
		}
		// fmt.Printf("main event name:%v data: %+v", evt.Source, evt.Data)
		if evt.Source == "answer" { //  llm result
			if cv, ok := evt.Data.(*llm.Chunk); ok {
				_ = cv
				fmt.Print(cv.Delta.Message.Content())
			}
		}

		// if cv, ok := evt.Data.(*llm.Chunk); ok {
		// 	_ = cv
		// 	fmt.Println("main chunk:", cv.Delta.Message)
		// } else {
		// 	// fmt.Println("main event name:", evt.Name, ", data:", evt.Data)
		// }
	}
}

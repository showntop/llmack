package main

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/showntop/llmack/llm"
	llog "github.com/showntop/llmack/log"

	"github.com/bookkeeper-ai/bookkeeper/api"
	"github.com/bookkeeper-ai/bookkeeper/config"
	"github.com/bookkeeper-ai/bookkeeper/database"
	"github.com/bookkeeper-ai/bookkeeper/services"
)

func main() {

	godotenv.Load()
	llog.SetLogger(&llog.WrapLogger{})
	// 注册模型
	llm.WithConfigs(map[string]any{
		"deepseek": map[string]any{
			"api_key": os.Getenv("deepseek_api_key"),
		},
		"qwen": map[string]any{
			"api_key": os.Getenv("qwen_api_key"),
		},
	})

	// 加载配置
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 初始化数据库
	if err := database.InitDB(cfg); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// 初始化服务
	imageService, err := services.NewImageService(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize image service: %v", err)
	}

	analysisService := services.NewAnalysisService()
	authService := services.NewAuthService()
	userService := services.NewUserService()
	chatService, err := services.NewChatService(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize chat service: %v", err)
	}

	// 创建 Gin 路由
	r := gin.Default()

	// 创建处理器
	transactionHandler := api.NewTransactionHandler(imageService)
	analysisHandler := api.NewAnalysisHandler(analysisService)
	authHandler := api.NewAuthHandler(authService)
	userHandler := api.NewUserHandler(userService)
	chatHandler := api.NewChatHandler(chatService)

	// 注册路由
	api := r.Group("/api")
	{
		// 认证相关路由
		api.POST("/auth/register", authHandler.Register)
		api.POST("/auth/login", authHandler.Login)

		// 需要认证的路由
		authorized := api.Group("/")
		// authorized.Use(authHandler.AuthMiddleware())
		{
			// 用户设置相关路由
			authorized.GET("/user/profile", userHandler.GetProfile)
			authorized.PUT("/user/profile", userHandler.UpdateProfile)
			authorized.PUT("/user/password", userHandler.UpdatePassword)
			authorized.POST("/user/avatar", userHandler.UpdateAvatar)

			// chat 相关路由
			authorized.POST("/chat", chatHandler.SendMessage)

			// 交易相关路由
			authorized.POST("/transactions", transactionHandler.CreateTransaction)
			authorized.GET("/transactions", transactionHandler.GetTransactions)
			authorized.GET("/transactions/analysis", transactionHandler.GetTransactionAnalysis)
			authorized.POST("/transactions/upload", transactionHandler.UploadReceipt)

			// 分析相关路由
			authorized.GET("/analysis/monthly", analysisHandler.GetMonthlyAnalysis)
			authorized.GET("/analysis/budget", analysisHandler.GetBudgetRecommendation)
		}
	}

	// 启动服务器
	if err := r.Run(":8080"); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

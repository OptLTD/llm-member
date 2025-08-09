package main

import (
	"log"
	"os"

	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/static"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	"swiflow-auth/internal/config"
	"swiflow-auth/internal/database"
	"swiflow-auth/internal/handlers"
	"swiflow-auth/internal/middleware"
	"swiflow-auth/internal/services"
)

func main() {
	// 加载环境变量
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found")
	}

	// 初始化配置
	cfg := config.Load()

	// 初始化数据库
	db, err := database.Initialize(cfg.DBPath)
	if err != nil {
		log.Fatal("Failed to initialize database:", err)
	}

	// 初始化服务
	userService := services.NewUserService(db)
	authService := services.NewAuthService(userService)
	llmService := services.NewLLMService(cfg)
	logService := services.NewLogService(db)

	// 初始化处理器
	h := handlers.New(llmService, logService, authService, userService, cfg)

	// 设置 Gin 模式
	gin.SetMode(cfg.GinMode)
	r := gin.Default()

	// 配置 CORS
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"*"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	// 静态文件服务
	r.Use(static.Serve("/", static.LocalFile("./web", true)))

	// API 路由
	api := r.Group("/api")
	{
		// 认证路由
		api.POST("/login", h.Login)
		api.POST("/logout", middleware.AuthMiddleware(authService), h.Logout)
		api.GET("/user", middleware.AuthMiddleware(authService), h.GetCurrentUser)

		// 大模型代理路由（使用 API Key 认证）
		api.POST("/chat/completions", middleware.APIKeyMiddleware(authService), h.ChatCompletions)

		// Web管理界面路由（使用Token认证）
		webAuth := api.Group("/", middleware.AuthMiddleware(authService))
		{
			webAuth.GET("/models", h.GetModels)
			webAuth.GET("/logs", h.GetLogs)
			webAuth.DELETE("/logs/:id", h.DeleteLog)
			webAuth.GET("/stats", h.GetStats)
			webAuth.GET("/config", h.GetConfig)
			webAuth.PUT("/config", h.UpdateConfig)
			webAuth.POST("/test/chat", h.TestChat) // 测试用聊天接口
		}

		// 管理功能路由（需要管理员权限）
		admin := api.Group("/admin", middleware.AuthMiddleware(authService), middleware.AdminMiddleware())
		{
			admin.POST("/register", h.Register)
			admin.GET("/users", h.GetUsers)
			admin.PUT("/users/:id", h.UpdateUser)
			admin.POST("/users/:id/regenerate-key", h.RegenerateAPIKey)
			admin.POST("/users/:id/toggle", h.ToggleUserStatus)
			admin.DELETE("/users/:id", h.DeleteUser)
		}
	}

	// 启动服务器
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server starting on port %s", port)
	log.Printf("Admin panel: http://localhost:%s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}

package main

import (
	"log"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	"llm-member/internal/auth"
	"llm-member/internal/config"
	"llm-member/internal/handle"
	"llm-member/internal/service"
)

func main() {
	// 加载环境变量
	if err := godotenv.Load(); err != nil {
		panic("Warning: .env file not found")
	}

	// 初始化配置（包含日志配置）
	cfg := config.Load()
	// 初始化数据库
	if err := config.InitDB(cfg.DBPath); err != nil {
		log.Fatal("Failed to initialize database:", err)
	}

	// 初始化服务
	authService := service.NewAuthService()

	// 设置 Gin 模式
	gin.SetMode(cfg.GinMode)
	// 为 Gin 设置日志输出
	gin.DefaultWriter = cfg.MultiWriter
	gin.DefaultErrorWriter = cfg.MultiWriter

	r := gin.Default()

	// 配置 CORS
	r.Use(cors.New(cors.Config{
		AllowHeaders: []string{"*"},
		AllowOrigins: []string{"*"},
		ExposeHeaders: []string{
			"Content-Length",
		},
		AllowMethods: []string{
			"GET", "POST", "PUT",
			"DELETE", "OPTIONS",
		},
		AllowCredentials: true,
	}))

	// 使用统一的路由和静态文件处理中间件
	r.Use(handle.StaticRouteHandle(cfg))

	// 管理功能路由（需要管理员权限）
	compatibleV1 := r.Group("/compatible-v1", auth.APIKeyMiddleware(authService))
	{
		r := handle.NewRelayHandle()
		compatibleV1.POST("/chat/completions", r.ChatCompletions)
	}

	api := r.Group("/api")
	{
		rootApi := api.Group("/", auth.NoneMiddleware())
		{
			a := handle.NewAuthHandle(authService)
			rootApi.POST("/verify", a.VerifyEmail)
			rootApi.POST("/signin", a.UserSignIn)
			rootApi.POST("/signup", a.UserSignUp)
			rootApi.POST("/signout", a.UserSignOut)
			rootApi.POST("/admin/signin", a.AdminSignIn)
			rootApi.POST("/admin/signout", a.AdminSignOut)
			rootApi.POST("/third/:provider", a.ThirdPartySignin) // 第三方登录

			s := handle.NewPublicHandler()
			rootApi.POST("/pricing-plans", s.GetPricingPlans)
		}

		// 需要认证的路由
		basicApi := api.Group("/", auth.AuthMiddleware(authService))
		{
			b := handle.NewBasicHandle()
			basicApi.PUT("/profile", b.SetUserProfile)
			basicApi.POST("/profile", b.GetUserProfile)
			basicApi.POST("/usage", b.GetUserUsage)
			basicApi.POST("/orders", b.GetUserOrders)
			basicApi.POST("/api-keys", b.GetUserAPIKeys)
			basicApi.POST("/regenerate", b.RegenerateKey)
		}

		// 订单路由
		orderApi := api.Group("/order", auth.AuthMiddleware(authService))
		{
			o := handle.NewOrderHandler()
			orderApi.POST("/create", o.CreatePaymentOrder)
			orderApi.POST("/methods", o.GetPaymentMethods)
			orderApi.POST("/callback", o.DoPaymentCallback)
			orderApi.POST("/query/:id", o.QueryPaymentOrder)
			orderApi.POST("/qrcode/:id", o.ShowPaymentQrcode)
		}

		// 管理员路由
		adminApi := api.Group("/admin", auth.AdminMiddleware(authService))
		{
			h := handle.NewAdminHandle()
			adminApi.GET("/current", h.Current)
			adminApi.PUT("/users/:id", h.UpdateUser)
			adminApi.POST("/users/create", h.CreateUser)
			adminApi.POST("/users/:id/toggle", h.ToggleUserStatus)
			adminApi.POST("/users/:id/generate", h.GenerateAPIKey)

			adminApi.POST("/logs", h.GetLogs)
			adminApi.GET("/stats", h.GetStats)
			adminApi.GET("/models", h.GetModels)
			adminApi.POST("/users", h.GetUsers)
			adminApi.POST("/orders", h.GetOrders)
		}

		// 设置路由
		setupApi := api.Group("/setup", auth.AdminMiddleware(authService))
		{
			s := handle.NewSetupHandler()
			setupApi.GET("/pricing", s.GetPricingPlans)
			setupApi.PUT("/pricing/:plan", s.SetPricingPlan)
		}
	}

	// 启动服务器
	log.Printf("Server starting on port %s", cfg.AppPort)
	log.Printf("Admin panel: http://localhost:%s", cfg.AppPort)
	if err := r.Run(":" + cfg.AppPort); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}

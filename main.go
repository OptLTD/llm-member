package main

import (
	"embed"
	"log"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	"llm-member/internal/auth"
	"llm-member/internal/config"
	"llm-member/internal/handle"
	"llm-member/internal/service"
)

//go:embed webroot
var webroot embed.FS

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found")
	}

	cfg := config.Load()
	if err := config.InitDB(cfg); err != nil {
		log.Fatal("Failed to initialize database:", err)
	}
	if err := config.InitRedis(cfg.Redis); err != nil {
		log.Println("Failed to initialize redis:", err)
	}
	if err := service.HandleInit(); err != nil {
		log.Fatal("Failed to initialize data:", err)
	}

	gin.SetMode(cfg.AppMode)
	r := gin.Default()
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

	r.Use(handle.StaticRouteHandle(cfg, webroot))
	authService := service.NewAuthService()
	v1 := r.Group("/v1")
	{
		relayHandle := handle.NewRelayHandle()
		authHandle := handle.NewAuthHandle(authService)
		authMiddle := auth.AuthMiddleware(authService)
		keyMiddle := auth.APIKeyMiddleware(authService)

		v1.POST("/build-token", authMiddle, authHandle.BuildCallbackToken) // 启用路由
		v1.POST("/verify-token", auth.NoneMiddleware(), authHandle.VerifyCallbackToken)
		v1.POST("/user-profile", keyMiddle, authHandle.GetUserProfile)
		v1.POST("/chat/completions", keyMiddle, relayHandle.ChatCompletions)
	}

	api := r.Group("/api")
	{
		rootApi := api.Group("/", auth.NoneMiddleware())
		{
			a := handle.NewAuthHandle(authService)
			rootApi.POST("/signin", a.UserSignIn)
			rootApi.POST("/signup", a.UserSignUp)
			rootApi.POST("/signout", a.UserSignOut)
			rootApi.POST("/admin/signin", a.AdminSignIn)
			rootApi.POST("/admin/signout", a.AdminSignOut)
			rootApi.POST("/third/:provider", a.ThirdSignin)
			rootApi.POST("/reset-password", a.ResetPassword)
			rootApi.POST("/forgot-password", a.ForgotPassword)
			rootApi.POST("/verify-reset-code", a.VerifyResetCode)
			rootApi.POST("/verify-signup-code", a.VerifySignUpCode)

			s := handle.NewPublicHandler()
			rootApi.POST("/pricing-plans", s.GetPricingPlans)
		}

		basicApi := api.Group("/", auth.AuthMiddleware(authService))
		{
			b := handle.NewBasicHandle()
			basicApi.GET("/profile", b.GetUserProfile)
			basicApi.PUT("/profile", b.SetUserProfile)
			basicApi.POST("/usage", b.GetUserUsage)
			basicApi.POST("/orders", b.GetUserOrders)
			basicApi.POST("/api-keys", b.GetUserAPIKeys)
			basicApi.POST("/regenerate", b.RegenerateKey)
		}

		orderApi := api.Group("/order", auth.AuthMiddleware(authService))
		{
			o := handle.NewOrderHandler()
			orderApi.POST("/create", o.CreatePaymentOrder)
			orderApi.POST("/methods", o.GetPaymentMethods)
			orderApi.POST("/callback", o.DoPaymentCallback)
			orderApi.POST("/query/:id", o.QueryPaymentOrder)
			orderApi.POST("/qrcode/:id", o.ShowPaymentQrcode)
		}

		adminApi := api.Group("/admin", auth.AdminMiddleware(authService))
		{
			h := handle.NewAdminHandle()
			adminApi.GET("/current", h.Current)
			adminApi.PUT("/users/:id", h.UpdateUser)
			adminApi.POST("/users/create", h.CreateUser)
			adminApi.POST("/users/:id/toggle", h.ToggleUserStatus)
			adminApi.POST("/users/:id/generate", h.GenerateAPIKey)

			adminApi.POST("/logs", h.GetLogs)
			adminApi.GET("/usage", h.GetUsage)
			adminApi.GET("/stats", h.GetStats)
			adminApi.GET("/models", h.GetModels)
			adminApi.POST("/users", h.GetUsers)
			adminApi.POST("/orders", h.GetOrders)
		}

		setupApi := api.Group("/setup", auth.AdminMiddleware(authService))
		{
			s := handle.NewSetupHandler()
			setupApi.GET("/pricing", s.GetPricingPlans)
			setupApi.PUT("/pricing/:plan", s.SetPricingPlan)
		}
	}

	log.Printf("Server starting on port %s", cfg.AppPort)
	log.Printf("Admin panel: http://localhost:%s", cfg.AppPort)
	if err := r.Run(":" + cfg.AppPort); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}

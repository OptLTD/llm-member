package routes

import (
	"github.com/OptLTD/member/internal/auth"
	"github.com/OptLTD/member/internal/config"
	"github.com/OptLTD/member/internal/handle"
	"github.com/OptLTD/member/internal/service"

	"github.com/gin-gonic/gin"
)

func Register(r *gin.Engine, authService *service.AuthService) {
	authMiddle := auth.AuthMiddleware(authService)
	adminMiddle := auth.AdminMiddleware(authService)
	keyMiddle := auth.APIKeyMiddleware(authService)
	noneMiddle := auth.NoneMiddleware()

	v1 := r.Group("/v1")
	{
		authHandle := handle.NewAuthHandle(authService)
		v1.POST("/build-token", authMiddle, authHandle.BuildCallbackToken)
		v1.POST("/verify-token", noneMiddle, authHandle.VerifyCallbackToken)

		v1.POST("/user-profile", keyMiddle,
			auth.RequireAnyModule(config.ModuleMember, config.ModuleLLMRelay),
			authHandle.GetUserProfile)

		relayHandle := handle.NewRelayHandle()
		v1.GET("/models", keyMiddle, auth.RequireModule(config.ModuleLLMRelay), relayHandle.ListModels)
		v1.POST("/chat/completions", keyMiddle, auth.RequireModule(config.ModuleLLMRelay), relayHandle.ChatCompletions)
	}

	api := r.Group("/api")
	{
		rootApi := api.Group("/", noneMiddle)
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
			rootApi.POST("/callback/:name", s.DoPaymentCallback)

			ah := handle.NewAnalyticsHandle()
			rootApi.POST("/track", auth.RequireModule(config.ModuleAnalytics), ah.Track)

			rh := handle.NewReferralHandle()
			rootApi.GET("/referral/resolve", auth.RequireModule(config.ModuleReferral), rh.Resolve)
			rootApi.POST("/referral/share-log", auth.RequireModule(config.ModuleReferral), rh.LogShare)

			sh := handle.NewSurveyHandle()
			rootApi.GET("/survey/:id", auth.RequireModule(config.ModuleSurvey), sh.GetPublic)
			rootApi.POST("/survey/:id/submit", auth.RequireModule(config.ModuleSurvey), sh.Submit)

			fh := handle.NewFeedbackHandle()
			rootApi.POST("/feedback", auth.RequireModule(config.ModuleFeedback), fh.Submit)
		}

		basicApi := api.Group("/", authMiddle)
		{
			b := handle.NewBasicHandle()
			basicApi.GET("/profile", auth.RequireModule(config.ModuleMember), b.GetUserProfile)
			basicApi.PUT("/profile", auth.RequireModule(config.ModuleMember), b.SetUserProfile)
			basicApi.POST("/usage", auth.RequireModule(config.ModuleMember), b.GetUserUsage)
			basicApi.POST("/orders", auth.RequireModule(config.ModuleMember), b.GetUserOrders)
			basicApi.POST("/api-keys", auth.RequireModule(config.ModuleMember), b.GetUserAPIKeys)
			basicApi.POST("/regenerate", auth.RequireModule(config.ModuleMember), b.RegenerateKey)

			dh := handle.NewDeviceHandle()
			basicApi.POST("/device/bind", auth.RequireModule(config.ModuleDevice), dh.Bind)
			basicApi.GET("/device/list", auth.RequireModule(config.ModuleDevice), dh.List)
			basicApi.DELETE("/device/:id", auth.RequireModule(config.ModuleDevice), dh.Unbind)

			ch := handle.NewCouponHandle()
			basicApi.POST("/coupon/validate", auth.RequireModule(config.ModuleCoupon), ch.Validate)
			basicApi.POST("/coupon/redeem", auth.RequireModule(config.ModuleCoupon), ch.Redeem)

			rh := handle.NewReferralHandle()
			basicApi.GET("/referral/my-code", auth.RequireModule(config.ModuleReferral), rh.MyCode)
			basicApi.GET("/referral/stats", auth.RequireModule(config.ModuleReferral), rh.Stats)
		}

		orderApi := api.Group("/order", authMiddle, auth.RequireModule(config.ModuleMember))
		{
			o := handle.NewOrderHandler()
			orderApi.POST("/create", o.CreatePaymentOrder)
			orderApi.POST("/methods", o.GetPaymentMethods)
			orderApi.POST("/query/:id", o.QueryPaymentOrder)
			orderApi.POST("/qrcode/:id", o.ShowPaymentQrcode)
		}

		adminApi := api.Group("/admin", adminMiddle)
		{
			h := handle.NewAdminHandle()
			adminApi.GET("/modules", h.GetModules)
			adminApi.GET("/system", h.GetSystemInfo)
			adminApi.PUT("/system/modules", h.UpdateModules)
			adminApi.GET("/current", h.Current)
			adminApi.POST("/me/api-key", h.IssueMyAPIKey)
			adminApi.PUT("/users/:id", h.UpdateUser)
			adminApi.POST("/users/create", h.CreateUser)
			adminApi.POST("/users/:id/toggle", h.ToggleUserStatus)
			adminApi.POST("/users/:id/generate", h.GenerateAPIKey)
			adminApi.POST("/users", h.GetUsers)
			adminApi.POST("/orders", auth.RequireModule(config.ModuleMember), h.GetOrders)
			adminApi.POST("/logs", auth.RequireModule(config.ModuleLLMRelay), h.GetLogs)
			adminApi.GET("/usage", auth.RequireModule(config.ModuleLLMRelay), h.GetUsage)
			adminApi.GET("/models", auth.RequireModule(config.ModuleLLMRelay), h.GetModels)
			adminApi.GET("/stats", h.GetStats)

			dh := handle.NewDeviceHandle()
			adminApi.GET("/devices", auth.RequireModule(config.ModuleDevice), dh.AdminList)
			adminApi.DELETE("/devices/:id", auth.RequireModule(config.ModuleDevice), dh.AdminUnbind)

			ah := handle.NewAnalyticsHandle()
			adminApi.GET("/analytics/sites", auth.RequireModule(config.ModuleAnalytics), ah.ListSites)
			adminApi.POST("/analytics/sites", auth.RequireModule(config.ModuleAnalytics), ah.CreateSite)
			adminApi.PUT("/analytics/sites/:id", auth.RequireModule(config.ModuleAnalytics), ah.UpdateSite)
			adminApi.GET("/analytics/overview", auth.RequireModule(config.ModuleAnalytics), ah.Overview)
			adminApi.GET("/analytics/events", auth.RequireModule(config.ModuleAnalytics), ah.ListEvents)

			ch := handle.NewCouponHandle()
			adminApi.GET("/coupons", auth.RequireModule(config.ModuleCoupon), ch.List)
			adminApi.POST("/coupons", auth.RequireModule(config.ModuleCoupon), ch.Create)
			adminApi.PUT("/coupons/:id", auth.RequireModule(config.ModuleCoupon), ch.Update)
			adminApi.GET("/coupons/:id/usages", auth.RequireModule(config.ModuleCoupon), ch.Usages)

			rh := handle.NewReferralHandle()
			adminApi.GET("/referral/campaigns", auth.RequireModule(config.ModuleReferral), rh.ListCampaigns)
			adminApi.POST("/referral/campaigns", auth.RequireModule(config.ModuleReferral), rh.SaveCampaign)
			adminApi.GET("/referral/list", auth.RequireModule(config.ModuleReferral), rh.ListReferrals)

			sh := handle.NewSurveyHandle()
			adminApi.GET("/surveys", auth.RequireModule(config.ModuleSurvey), sh.List)
			adminApi.POST("/surveys", auth.RequireModule(config.ModuleSurvey), sh.Create)
			adminApi.GET("/surveys/:id", auth.RequireModule(config.ModuleSurvey), sh.Get)
			adminApi.PUT("/surveys/:id", auth.RequireModule(config.ModuleSurvey), sh.Update)
			adminApi.POST("/surveys/questions", auth.RequireModule(config.ModuleSurvey), sh.SaveQuestion)
			adminApi.DELETE("/surveys/questions/:id", auth.RequireModule(config.ModuleSurvey), sh.DeleteQuestion)
			adminApi.GET("/surveys/:id/responses", auth.RequireModule(config.ModuleSurvey), sh.Responses)

			fh := handle.NewFeedbackHandle()
			adminApi.GET("/feedbacks", auth.RequireModule(config.ModuleFeedback), fh.List)
			adminApi.PUT("/feedbacks/:id", auth.RequireModule(config.ModuleFeedback), fh.Update)
			adminApi.GET("/feedbacks/summary", auth.RequireModule(config.ModuleFeedback), fh.Summary)
		}

		setupApi := api.Group("/setup", adminMiddle)
		{
			s := handle.NewSetupHandler()
			setupApi.GET("/pricing", s.GetPricingPlans)
			setupApi.PUT("/pricing/:plan", s.SetPricingPlan)
		}
	}
}

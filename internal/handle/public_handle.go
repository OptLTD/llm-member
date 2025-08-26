package handle

import (
	"html/template"
	"llm-member/internal/config"
	"llm-member/internal/service"
	"net/http"
	"path/filepath"

	"github.com/gin-contrib/static"
	"github.com/gin-gonic/gin"
)

type PublicHandle struct {
	setupService *service.SetupService
}

func NewPublicHandler() *PublicHandle {
	return &PublicHandle{
		setupService: service.NewSetupService(),
	}
}

// GetPricingPlans 获取定价方案
func (h *PublicHandle) GetPricingPlans(c *gin.Context) {
	plans := h.setupService.GetEnablePlans()
	c.JSON(http.StatusOK, gin.H{
		"plans": plans,
	})
}

// StaticRouteHandle 统一的路由和静态文件处理中间件
func StaticRouteHandle(cfg *config.Config) gin.HandlerFunc {
	var i18nService = service.NewI18nService()
	var RenderPage = func(c *gin.Context, templateName string) {
		// 获取语言参数
		language := c.DefaultQuery("lang", "zh")
		translations := i18nService.GetTranslations(language)
		templateData := gin.H{
			"AppName": cfg.AppName, "AppDesc": cfg.AppDesc,
			"Language": language, "T": translations,
		}

		var curr *template.Template
		tmplPath := filepath.Join("./webroot", templateName+".html")
		if tmpl, err := template.ParseFiles(tmplPath); err != nil {
			c.String(404, "页面加载失败")
		} else {
			curr = tmpl
		}

		c.Status(200)
		c.Header("Content-Type", "text/html; charset=utf-8")
		if err := curr.Execute(c.Writer, templateData); err != nil {
			c.String(500, "页面渲染失败")
		}
	}

	return func(c *gin.Context) {
		switch c.Request.URL.Path {
		case "/authorization":
			RenderPage(c, "signin")
		case "/", "/index.html":
			RenderPage(c, "index")
		case "/admin", "/admin.html":
			RenderPage(c, "admin")
		case "/reset", "/reset.html":
			RenderPage(c, "reset")
		case "/verify", "/verify.html":
			RenderPage(c, "verify")
		case "/signin", "/signin.html":
			RenderPage(c, "signin")
		case "/signup", "/signup.html":
			RenderPage(c, "signup")
		case "/forget", "/forget.html":
			RenderPage(c, "forget")
		case "/payment", "/payment.html":
			RenderPage(c, "payment")
		case "/profile", "/profile.html":
			RenderPage(c, "profile")
		case "/pricing", "/pricing.html":
			RenderPage(c, "pricing")
		default:
			fs := static.LocalFile("./webroot", false)
			static.Serve("/", fs)(c)
		}
	}
}

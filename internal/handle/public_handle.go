package handle

import (
	"embed"
	"html/template"
	"llm-member/internal/config"
	"llm-member/internal/model"
	"llm-member/internal/service"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gin-contrib/static"
	"github.com/gin-gonic/gin"
)

type PublicHandle struct {
	setupService *service.SetupService
	orderService *service.OrderService
}

func NewPublicHandler() *PublicHandle {
	return &PublicHandle{
		setupService: service.NewSetupService(),
		orderService: service.NewOrderService(),
	}
}

// GetPricingPlans 获取定价方案
func (h *PublicHandle) GetPricingPlans(c *gin.Context) {
	plans := h.setupService.GetEnablePlans()
	c.JSON(http.StatusOK, gin.H{
		"plans": plans,
	})
}

// DoPaymentCallback 处理支付回调
func (h *PublicHandle) DoPaymentCallback(c *gin.Context) {
	if name := c.Param("name"); name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "支付方式"})
		return
	}

	// 验证支付回调签名
	order, err := h.orderService.VerifyCallback(c.Param("name"), c.Request)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "回调验证失败: " + err.Error()})
		return
	}
	var limit *model.ApiLimit
	if limit, err = h.setupService.GetPlanLimit(order.PayPlan); err == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "该套餐暂不可用"})
		return
	}

	// 标记为支付成功
	err = h.orderService.PaySuccess(order.OrderID, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "支付成功"})
}

// StaticRouteHandle 统一的路由和静态文件处理中间件
func StaticRouteHandle(cfg *config.Config, webroot embed.FS) gin.HandlerFunc {
	var i18nService = service.NewI18nService()
	var RenderPage = func(c *gin.Context, name string) {
		tmplPath := filepath.Join("webroot", name+".html")
		tmplData, err := webroot.ReadFile(tmplPath)
		if err != nil {
			c.String(404, "页面加载失败")
			return
		}
		var curr = template.New(name)

		// 解析主模板
		if _, err := curr.Parse(string(tmplData)); err != nil {
			c.String(500, "模板解析失败")
			return
		}

		pageTmpl := filepath.Join("template", "page."+name+".html")
		if pageData, err := os.ReadFile(pageTmpl); err == nil {
			if _, err := curr.Parse(string(pageData)); err != nil {
				c.String(500, pageTmpl+"模板解析失败: "+err.Error())
				return
			}
		}

		c.Status(200)
		// 获取语言参数
		language := c.DefaultQuery("lang", "zh")
		translations := i18nService.GetTranslations(language)
		templateData := gin.H{
			"AppName": cfg.AppName, "AppDesc": cfg.AppDesc, "Version": cfg.Version,
			"Language": language, "T": translations,
		}
		c.Header("Content-Type", "text/html; charset=utf-8")
		if err := curr.Execute(c.Writer, templateData); err != nil {
			c.String(500, "页面渲染失败："+err.Error())
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
			fs, _ := static.EmbedFolder(
				webroot, "webroot",
			)
			static.Serve("/", fs)(c)
		}
	}
}

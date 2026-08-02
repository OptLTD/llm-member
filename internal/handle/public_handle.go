package handle

import (
	"embed"
	"github.com/OptLTD/member/internal/config"
	"github.com/OptLTD/member/internal/model"
	"github.com/OptLTD/member/internal/service"
	"html/template"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"

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
	plans := h.setupService.GetEnablePlans(service.ProjectID(c))
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	var limit *model.ApiLimit
	if limit, err = h.setupService.GetPlanLimit(order.ProjectID, order.PayPlan); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
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

const uiFrontRoot = "ui-front"

// StaticRouteHandle 统一的路由和静态文件处理中间件
func StaticRouteHandle(cfg *config.Config, uiFront embed.FS) gin.HandlerFunc {
	var i18nService = service.NewI18nService()
	var RenderPage = func(c *gin.Context, name string) {
		tmplPath := filepath.Join(uiFrontRoot, name+".html")
		tmplData, err := uiFront.ReadFile(tmplPath)
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
			serveAdminSPA(c, uiFront)
		case "/feedback", "/feedback.html":
			serveStaticHTML(c, uiFront, uiFrontRoot+"/feedback.html")
		case "/survey", "/survey.html":
			serveStaticHTML(c, uiFront, uiFrontRoot+"/survey.html")
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
		case "/auth-callback", "/auth-callback.html":
			serveStaticHTML(c, uiFront, uiFrontRoot+"/auth-callback.html")
		case "/payment", "/payment.html":
			RenderPage(c, "payment")
		case "/profile", "/profile.html":
			RenderPage(c, "profile")
		case "/pricing", "/pricing.html":
			RenderPage(c, "pricing")
		default:
			if strings.HasPrefix(c.Request.URL.Path, "/survey/") {
				serveStaticHTML(c, uiFront, uiFrontRoot+"/survey.html")
				return
			}
			if strings.HasPrefix(c.Request.URL.Path, "/admin/") {
				serveAdminSPA(c, uiFront)
				return
			}
			fs, _ := static.EmbedFolder(
				uiFront, uiFrontRoot,
			)
			static.Serve("/", fs)(c)
		}
	}
}

func serveStaticHTML(c *gin.Context, uiFront embed.FS, file string) {
	data, err := uiFront.ReadFile(file)
	if err != nil {
		c.String(http.StatusNotFound, "页面不存在")
		return
	}
	c.Data(http.StatusOK, "text/html; charset=utf-8", data)
}

func serveAdminSPA(c *gin.Context, uiFront embed.FS) {
	p := c.Request.URL.Path
	if strings.HasPrefix(p, "/admin/") && p != "/admin/" {
		rel := strings.TrimPrefix(p, "/admin/")
		if rel != "" {
			data, err := uiFront.ReadFile(path.Join(uiFrontRoot, "admin", "dist", rel))
			if err == nil {
				c.Data(http.StatusOK, contentTypeForPath(rel), data)
				return
			}
		}
	}
	data, err := uiFront.ReadFile(uiFrontRoot + "/admin/dist/index.html")
	if err != nil {
		c.String(http.StatusNotFound, "Admin UI not built. Run: cd ui-admin && npm run build")
		return
	}
	c.Data(http.StatusOK, "text/html; charset=utf-8", data)
}

func contentTypeForPath(p string) string {
	switch path.Ext(p) {
	case ".js":
		return "application/javascript"
	case ".css":
		return "text/css"
	case ".svg":
		return "image/svg+xml"
	case ".png":
		return "image/png"
	default:
		return "application/octet-stream"
	}
}

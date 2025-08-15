package handle

import (
	"llm-member/internal/model"
	"llm-member/internal/service"
	"llm-member/internal/support"
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

type AuthHandle struct {
	authService  *service.AuthService
	userService  *service.UserService
	emailService *service.EmailService
}

func NewAuthHandle(authService *service.AuthService) *AuthHandle {
	return &AuthHandle{
		authService:  authService,
		userService:  service.NewUserService(),
		emailService: service.NewEmailService(),
	}
}

// UserSignUp 用户注册
func (h *AuthHandle) UserSignUp(c *gin.Context) {
	var req model.SignUpRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	// 验证至少提供邮箱或手机号之一
	if req.Email == "" && req.Phone == "" {
		c.JSON(400, gin.H{"error": "请提供邮箱或手机号"})
		return
	}

	// 检查是否支持手机号注册
	if support.IsPhoneNumber(req.Username) {
		// 这里可以添加手机号验证逻辑
		req.Email = req.Username + "@temp.phone" // 临时邮箱格式
	}

	// 转换为标准注册请求
	registerReq := &model.SignUpRequest{
		Email:    req.Email,
		Username: req.Username,
		Password: req.Password,
	}

	// 如果没有邮箱但有手机号，使用手机号作为邮箱（临时方案）
	if registerReq.Email == "" && req.Phone != "" {
		registerReq.Email = req.Phone + "@phone.local"
	}

	// 注册用户
	user, err := h.authService.SignUp(registerReq)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 生成注册 token
	reset, err := h.authService.GenerateSignupToken(user.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "生成注册 token 失败",
		})
		return
	}
	// 发送验证邮件
	if err := h.emailService.SendSignupCodeEmail(reset); err != nil {
		// 在生产环境中，您可能希望处理此错误，例如，将任务排队重试
		// 就目前而言，我们只记录它
		log.Printf("发送验证邮件失败: %v", err)
	}

	// 如果提供了手机号，更新用户的手机号字段
	if req.Phone != "" {
		h.userService.UpdateUser(&model.UpdateUserRequest{
			UserId: user.ID, Phone: req.Phone,
		})
	}

	c.JSON(http.StatusCreated, model.RegisterResponse{
		User: user, APIKey: user.APIKey,
		Message: "注册成功，请检查您的邮箱以获取验证链接。",
	})
}

// VerifyEmail handles email verification.
func (h *AuthHandle) VerifyEmail(c *gin.Context) {
	if code := c.Query("code"); code == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "验证码不能为空"})
		return
	}

	err := h.authService.VerifyEmail(c.Query("code"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "邮箱验证成功"})
}

// UserSignIn 用户登录（用户端）
func (h *AuthHandle) UserSignIn(c *gin.Context) {
	var req model.SignInRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 尝试不同的登录方式
	var user *model.UserModel
	var err error

	// 1. 尝试用户名登录
	user, err = h.userService.GetUserByUsername(req.Username)
	if err != nil {
		// 如果用户名登录失败，提示用户不存在
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "用户不存在，是否要创建新账户？",
		})
		return
	}

	// 检查用户是否激活
	if !user.IsActive {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户账户已被禁用"})
		return
	}

	// 验证密码
	if !h.userService.ValidatePassword(user.Password, req.Password) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "密码错误"})
		return
	}

	// 生成 token
	token, _, err := h.authService.SignIn(user.Username, req.Password)
	if err != nil {
		log.Println("SignIn failed:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "生成token失败"})
		return
	}

	// 清除密码字段
	user.Password = ""

	c.JSON(http.StatusOK, model.SignInResponse{
		Token: token, User: user, Message: "登录成功",
	})
}

// UserSignOut 用户登出（用户端）
func (h *AuthHandle) UserSignOut(c *gin.Context) {
	authHeader := c.GetHeader("Authorization")
	if authHeader != "" {
		token := strings.TrimPrefix(authHeader, "Bearer ")
		h.authService.Logout(token)
	}

	c.JSON(http.StatusOK, gin.H{"message": "登出成功"})
}

// AdminSignIn 用户登录（管理员端）
func (h *AuthHandle) AdminSignIn(c *gin.Context) {
	var req model.SignInRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	token, user, err := h.authService.SignIn(req.Username, req.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, model.SignInResponse{
		Token: token, User: user, Message: "登录成功",
	})
}

// AdminSignOut 用户登出（管理员端）
func (h *AuthHandle) AdminSignOut(c *gin.Context) {
	authHeader := c.GetHeader("Authorization")
	if authHeader != "" {
		token := strings.TrimPrefix(authHeader, "Bearer ")
		h.authService.Logout(token)
	}

	c.JSON(http.StatusOK, gin.H{"message": "登出成功"})
}

// ThirdPartySignin 第三方登录（用户端）
func (h *AuthHandle) ThirdPartySignin(c *gin.Context) {
	provider := c.Param("provider") // google, github, etc.
	code := c.Query("code")

	if code == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "缺少授权码"})
		return
	}

	// 这里应该实现第三方OAuth逻辑
	// 1. 使用code换取access_token
	// 2. 获取用户信息
	// 3. 检查用户是否存在，不存在则创建
	// 4. 生成token返回
	c.JSON(http.StatusNotImplemented, gin.H{
		"error":    "第三方登录功能正在开发中",
		"provider": provider,
	})
}

package service

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"sync"
	"time"

	"llm-member/internal/config"
	"llm-member/internal/model"
	"llm-member/internal/support"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

var tokens map[string]*TokenInfo

type AuthService struct {
	db *gorm.DB

	mutex sync.RWMutex

	userService *UserService
}

type TokenInfo struct {
	UserID  uint64
	Expiry  time.Time
	IsAdmin bool
}

func NewAuthService() *AuthService {
	service := &AuthService{
		db: config.GetDB(),

		userService: NewUserService(),
	}
	if tokens == nil {
		tokens = make(map[string]*TokenInfo)
	}

	// 启动定期清理过期 token 的协程
	go service.cleanupExpiredTokens()

	return service
}

// SignIn 用户登录
func (s *AuthService) SignIn(username, password string) (string, *model.UserModel, error) {
	// 获取用户信息
	user, err := s.userService.GetUserByUsername(username)
	if err != nil {
		return "", nil, fmt.Errorf("用户名或密码错误")
	}

	// 验证密码
	if !s.userService.ValidatePassword(user.Password, password) {
		return "", nil, fmt.Errorf("用户名或密码错误")
	}

	// 检查用户是否激活
	if !user.IsActive {
		return "", nil, fmt.Errorf("用户账户已被禁用")
	}

	// 检查用户是否激活
	if !user.Verified {
		return "", nil, fmt.Errorf("未验证用户账户")
	}

	// 生成 token
	token, err := s.generateToken()
	if err != nil {
		return "", nil, err
	}

	// 存储 token 信息
	s.mutex.Lock()
	tokens[token] = &TokenInfo{
		UserID:  user.ID,
		IsAdmin: user.CurrRole == model.RoleAdmin,
		Expiry:  time.Now().Add(24 * time.Hour),
	}
	s.mutex.Unlock()

	// 清除密码字段
	user.Password = ""
	return token, user, nil
}

// VerifyEmail verifies a user's email address.
func (s *AuthService) VerifyEmail(code string) error {
	var reset model.TokenModel
	var query = s.db.Where("token = ?", code)
	if err := query.First(&reset).Error; err != nil {
		return fmt.Errorf("无效的验证码")
	}
	if reset.ExpiredAt.Before(time.Now()) {
		return fmt.Errorf("验证码已过期")
	}

	user, err := s.userService.GetUserByEmail(reset.Email)
	if err != nil || user == nil {
		return fmt.Errorf("用户不存在")
	}
	user.Verified = true
	if err := s.db.Save(&user).Error; err != nil {
		return fmt.Errorf("邮箱验证失败: %v", err)
	}
	if err := s.db.Delete(&reset).Error; err != nil {
		return fmt.Errorf("删除验证码失败: %v", err)
	}
	return nil
}

// SignUp 用户注册
func (s *AuthService) SignUp(req *model.SignUpRequest) (*model.UserModel, error) {
	// 检查用户名是否已存在
	existingUser, err := s.userService.GetUserByUsername(req.Username)
	if err == nil && existingUser != nil {
		return nil, fmt.Errorf("用户名已存在")
	}

	// 检查邮箱是否已存在
	existingUser, err = s.userService.GetUserByEmail(req.Email)
	if err == nil && existingUser != nil {
		return nil, fmt.Errorf("邮箱已存在")
	}

	// 生成密码哈希
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("密码加密失败: %v", err)
	}

	// 生成 API Key
	apiKey, err := support.GenerateAPIKey()
	if err != nil {
		return nil, fmt.Errorf("API Key 生成失败: %v", err)
	}

	// 创建新用户
	user := &model.UserModel{
		DailyLimit: 1000, MonthlyLimit: 10000,
		Email: req.Email, Username: req.Username,
		APIKey: apiKey, Password: string(hashedPassword),
		CurrPlan: model.PlanBasic, CurrRole: model.RoleUser,
	}
	if err := s.userService.CreateUser(user); err != nil {
		return nil, fmt.Errorf("创建用户失败: %v", err)
	}
	return user, nil
}

func (s *AuthService) GenerateSignupToken(email string) (*model.TokenModel, error) {
	reset := &model.TokenModel{Email: email}
	if token, err := s.generateToken(); err != nil {
		return nil, fmt.Errorf("验证码生成失败: %v", err)
	} else {
		reset.Token = token
		reset.ExpiredAt = time.Now().Add(10 * time.Minute)
	}
	return reset, nil
}

// ValidateToken 验证 token
func (s *AuthService) ValidateSigninToken(token string) (*TokenInfo, bool) {
	s.mutex.RLock()
	info, exists := tokens[token]
	s.mutex.RUnlock()

	if !exists {
		return nil, false
	}

	if time.Now().After(info.Expiry) {
		s.mutex.Lock()
		delete(tokens, token)
		s.mutex.Unlock()
		return nil, false
	}

	return info, true
}

// ValidateAPIKey 验证 API Key
func (s *AuthService) ValidateAPIKey(apiKey string) (*model.UserModel, error) {
	user, err := s.userService.GetUserByAPIKey(apiKey)
	if err != nil {
		return nil, fmt.Errorf("无效的 API Key")
	}

	// 检查用户限制
	if err := s.userService.CheckUserLimits(user.ID); err != nil {
		return nil, err
	}

	return user, nil
}

// Logout 用户登出
func (s *AuthService) Logout(token string) {
	s.mutex.Lock()
	delete(tokens, token)
	s.mutex.Unlock()
}

// GetUserFromSigninToken 从 token 获取用户信息
func (s *AuthService) GetUserFromSigninToken(token string) (*model.UserModel, error) {
	tokenInfo, valid := s.ValidateSigninToken(token)
	if !valid {
		return nil, fmt.Errorf("无效的 token")
	}

	user, err := s.userService.GetUserByID(tokenInfo.UserID)
	if err != nil {
		return nil, fmt.Errorf("用户不存在")
	}

	user.Password = ""
	return user, nil
}

// generateToken 生成随机 token
func (s *AuthService) generateToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// cleanupExpiredTokens 定期清理过期的 token
func (s *AuthService) cleanupExpiredTokens() {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	for range ticker.C {
		s.mutex.Lock()
		now := time.Now()
		for token, tokenInfo := range tokens {
			if now.After(tokenInfo.Expiry) {
				delete(tokens, token)
			}
		}
		s.mutex.Unlock()
	}
}

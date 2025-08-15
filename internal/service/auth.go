package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"llm-member/internal/config"
	"llm-member/internal/model"
	"llm-member/internal/support"

	"github.com/redis/go-redis/v9"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// 删除全局变量
// var tokens map[string]*TokenInfo

type AuthService struct {
	conn  *gorm.DB
	redis *redis.Client
	userS *UserService

	// 将tokens移动到结构体内部
	mutex sync.RWMutex
	token map[string]*TokenInfo
}

type TokenInfo struct {
	UserID  uint64
	Expiry  time.Time
	IsAdmin bool
}

func NewAuthService() *AuthService {
	service := &AuthService{
		conn:  config.GetDB(),
		redis: config.GetRedis(),
		userS: NewUserService(),
		token: make(map[string]*TokenInfo), // 初始化tokens
	}

	// 启动定期清理过期 token 的协程
	go service.cleanupExpiredTokens()
	return service
}

// SignIn 用户登录
// SignIn 方法中的token存储部分
func (s *AuthService) SignIn(username, password string) (string, *model.UserModel, error) {
	// 获取用户信息
	user, err := s.userS.GetUserByUsername(username)
	if err != nil {
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

	// 验证密码
	if !s.userS.ValidatePassword(user.Password, password) {
		return "", nil, fmt.Errorf("用户名或密码错误")
	}

	// 生成 token
	token, err := s.generateToken()
	if err != nil {
		return "", nil, err
	}

	// 存储 token 信息
	tokenInfo := &TokenInfo{
		UserID:  user.ID,
		IsAdmin: user.UserRole == model.RoleAdmin,
		Expiry:  time.Now().Add(24 * time.Hour),
	}

	err = s.SetToken(token, tokenInfo)
	if err != nil {
		return "", nil, fmt.Errorf("存储token失败: %v", err)
	}

	// 清除密码字段
	user.Password = ""
	return token, user, nil
}

// SignOut 方法
func (s *AuthService) SignOut(token string) {
	s.DeleteToken(token)
}

// SignUp 用户注册
func (s *AuthService) SignUp(req *model.SignUpRequest) (*model.UserModel, error) {
	// 检查用户名是否已存在
	existingUser, err := s.userS.GetUserByUsername(req.Username)
	if err == nil && existingUser != nil {
		return nil, fmt.Errorf("用户名已存在")
	}

	// 检查邮箱是否已存在
	existingUser, err = s.userS.GetUserByEmail(req.Email)
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
		UserPlan: model.PlanBasic, UserRole: model.RoleUser,
	}
	if err := s.userS.CreateUser(user); err != nil {
		return nil, fmt.Errorf("创建用户失败: %v", err)
	}
	return user, nil
}

func (s *AuthService) DeleteSignup(user *model.UserModel) error {
	return s.userS.DeleteUser(user.ID)
}

// VerifyEmail verifies a user's email address.
func (s *AuthService) VerifySignupCode(code string) error {
	var reset model.VerifyModel
	var query = s.conn.Where("token = ?", code)
	if err := query.First(&reset).Error; err != nil {
		return fmt.Errorf("无效的验证码")
	}
	if reset.ExpiredAt.Before(time.Now()) {
		return fmt.Errorf("验证码已过期")
	}

	user, err := s.userS.GetUserByEmail(reset.Email)
	if err != nil || user == nil {
		return fmt.Errorf("用户不存在")
	}
	user.Verified = true
	if err := s.conn.Save(&user).Error; err != nil {
		return fmt.Errorf("邮箱验证失败: %v", err)
	}
	if err := s.conn.Delete(&reset).Error; err != nil {
		return fmt.Errorf("删除验证码失败: %v", err)
	}
	return nil
}

func (s *AuthService) GenerateSignupCode(email string) (*model.VerifyModel, error) {
	reset := &model.VerifyModel{Email: email}
	if token, err := s.generateToken(); err != nil {
		return nil, fmt.Errorf("验证码生成失败: %v", err)
	} else {
		reset.Token = token
		reset.ExpiredAt = time.Now().Add(10 * time.Minute)
	}
	if res := s.conn.Create(reset); res.Error != nil {
		return nil, fmt.Errorf("验证码生成失败: %v", res.Error)
	}
	return reset, nil
}

// ValidateToken 验证 token
func (s *AuthService) ValidateSigninToken(token string) (*TokenInfo, bool) {
	return s.GetToken(token)
}

// ValidateAPIKey 验证 API Key
func (s *AuthService) ValidateAPIKey(apiKey string) (*model.UserModel, error) {
	user, err := s.userS.GetUserByAPIKey(apiKey)
	if err != nil {
		return nil, fmt.Errorf("无效的 API Key")
	}

	// 检查用户限制
	if err := s.userS.CheckUserLimits(user.ID); err != nil {
		return nil, err
	}

	return user, nil
}

// GetUserFromSigninToken 从 token 获取用户信息
func (s *AuthService) GetUserFromSigninToken(token string) (*model.UserModel, error) {
	tokenInfo, valid := s.ValidateSigninToken(token)
	if !valid {
		return nil, fmt.Errorf("无效的 token")
	}

	user, err := s.userS.GetUserByID(tokenInfo.UserID)
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

// ForgotPassword 忘记密码 - 生成重置token并发送邮件
func (s *AuthService) ForgotPassword(email string) (*model.VerifyModel, error) {
	// 检查用户是否存在
	user, err := s.userS.GetUserByEmail(email)
	if err != nil {
		return nil, fmt.Errorf("该邮箱未注册")
	}

	// 检查用户是否激活
	if !user.IsActive {
		return nil, fmt.Errorf("用户账户已被禁用")
	}

	// 生成重置token
	reset := &model.VerifyModel{Email: email}
	if token, err := s.generateToken(); err != nil {
		return nil, fmt.Errorf("重置码生成失败: %v", err)
	} else {
		reset.Token = token
		reset.ExpiredAt = time.Now().Add(30 * time.Minute) // 30分钟有效期
	}

	// 删除该邮箱之前的重置记录
	s.conn.Where("email = ?", email).Delete(&model.VerifyModel{})

	// 保存新的重置记录
	if res := s.conn.Create(reset); res.Error != nil {
		return nil, fmt.Errorf("重置码生成失败: %v", res.Error)
	}

	return reset, nil
}

// VerifyResetCode 验证重置码
func (s *AuthService) VerifyResetCode(code string) (*model.VerifyModel, error) {
	var reset model.VerifyModel
	if err := s.conn.Where("token = ?", code).First(&reset).Error; err != nil {
		return nil, fmt.Errorf("无效的重置码")
	}

	if reset.ExpiredAt.Before(time.Now()) {
		return nil, fmt.Errorf("重置码已过期")
	}

	return &reset, nil
}

// ResetPassword 重置密码
func (s *AuthService) ResetPassword(code, newPassword string) error {
	// 验证重置码
	reset, err := s.VerifyResetCode(code)
	if err != nil {
		return err
	}

	// 获取用户
	user, err := s.userS.GetUserByEmail(reset.Email)
	if err != nil {
		return fmt.Errorf("用户不存在")
	}

	// 生成新密码哈希
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("密码加密失败: %v", err)
	}

	// 更新用户密码
	user.Password = string(hashedPassword)
	if err := s.conn.Save(user).Error; err != nil {
		return fmt.Errorf("密码更新失败: %v", err)
	}

	// 删除已使用的重置记录
	s.conn.Delete(reset)

	// 清除该用户的所有登录token（强制重新登录）
	s.mutex.Lock()
	for token, tokenInfo := range s.token {
		if tokenInfo.UserID == user.ID {
			delete(s.token, token)
		}
	}
	s.mutex.Unlock()
	return nil
}

// cleanupExpiredTokens 定期清理过期的 token
func (s *AuthService) cleanupExpiredTokens() {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	for range ticker.C {
		s.mutex.Lock()
		now := time.Now()
		for token, tokenInfo := range s.token {
			if now.After(tokenInfo.Expiry) {
				delete(s.token, token)
			}
		}
		s.mutex.Unlock()
	}
}

// SetToken 写入token信息
func (s *AuthService) SetToken(token string, tokenInfo *TokenInfo) error {
	if s.redis == nil {
		// 使用内存存储
		s.mutex.Lock()
		s.token[token] = tokenInfo
		s.mutex.Unlock()
		return nil
	}

	// 使用Redis存储
	ctx := context.Background()
	tokenData, err := json.Marshal(tokenInfo)
	if err != nil {
		return fmt.Errorf("序列化token信息失败: %v", err)
	}

	// 计算过期时间
	expiration := time.Until(tokenInfo.Expiry)
	if expiration <= 0 {
		return fmt.Errorf("token已过期")
	}

	err = s.redis.Set(ctx, "token:"+token, tokenData, expiration).Err()
	if err != nil {
		return fmt.Errorf("Redis存储token失败: %v", err)
	}
	return nil
}

// GetToken 读取token信息
func (s *AuthService) GetToken(token string) (*TokenInfo, bool) {
	if s.redis == nil {

		// 从内存读取
		s.mutex.RLock()
		info, exists := s.token[token]
		s.mutex.RUnlock()

		if !exists {
			return nil, false
		}

		// 检查是否过期
		if time.Now().After(info.Expiry) {
			s.mutex.Lock()
			delete(s.token, token)
			s.mutex.Unlock()
			return nil, false
		}

		return info, true
	}

	// 从Redis读取
	ctx := context.Background()
	tokenData, err := s.redis.Get(ctx, "token:"+token).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, false // token不存在
		}
		// Redis错误，降级到内存存储
		s.mutex.RLock()
		info, exists := s.token[token]
		s.mutex.RUnlock()
		return info, exists
	}

	var tokenInfo TokenInfo
	err = json.Unmarshal([]byte(tokenData), &tokenInfo)
	if err != nil {
		return nil, false
	}

	// 检查是否过期
	if time.Now().After(tokenInfo.Expiry) {
		// 删除过期token
		s.redis.Del(ctx, "token:"+token)
		return nil, false
	}
	return &tokenInfo, true
}

// DeleteToken 删除token
func (s *AuthService) DeleteToken(token string) {
	if s.redis != nil {
		// 从Redis删除
		ctx := context.Background()
		s.redis.Del(ctx, "token:"+token)
	}
	// 同时从内存删除（防止降级情况）
	s.mutex.Lock()
	delete(s.token, token)
	s.mutex.Unlock()
}

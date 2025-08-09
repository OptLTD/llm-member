package services

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"sync"
	"time"

	"swiflow-auth/internal/models"
)

type AuthService struct {
	tokens      map[string]*TokenInfo
	mutex       sync.RWMutex
	userService *UserService
}

type TokenInfo struct {
	UserID  uint
	Expiry  time.Time
	IsAdmin bool
}

func NewAuthService(userService *UserService) *AuthService {
	service := &AuthService{
		tokens:      make(map[string]*TokenInfo),
		userService: userService,
	}
	
	// 启动定期清理过期 token 的协程
	go service.cleanupExpiredTokens()
	
	return service
}

// Login 用户登录
func (s *AuthService) Login(username, password string) (string, *models.User, error) {
	// 获取用户信息
	user, err := s.userService.GetUserByUsername(username)
	if err != nil {
		return "", nil, fmt.Errorf("用户名或密码错误")
	}

	// 检查用户是否激活
	if !user.IsActive {
		return "", nil, fmt.Errorf("用户账户已被禁用")
	}

	// 验证密码
	if !s.userService.ValidatePassword(user.Password, password) {
		return "", nil, fmt.Errorf("用户名或密码错误")
	}

	// 生成 token
	token, err := s.generateToken()
	if err != nil {
		return "", nil, err
	}
	
	// 存储 token 信息
	s.mutex.Lock()
	s.tokens[token] = &TokenInfo{
		UserID:  user.ID,
		Expiry:  time.Now().Add(24 * time.Hour), // token 24小时有效
		IsAdmin: user.IsAdmin,
	}
	s.mutex.Unlock()
	
	// 清除密码字段
	user.Password = ""
	return token, user, nil
}

// ValidateToken 验证 token
func (s *AuthService) ValidateToken(token string) (*TokenInfo, bool) {
	s.mutex.RLock()
	tokenInfo, exists := s.tokens[token]
	s.mutex.RUnlock()
	
	if !exists {
		return nil, false
	}
	
	if time.Now().After(tokenInfo.Expiry) {
		s.mutex.Lock()
		delete(s.tokens, token)
		s.mutex.Unlock()
		return nil, false
	}
	
	return tokenInfo, true
}

// ValidateAPIKey 验证 API Key
func (s *AuthService) ValidateAPIKey(apiKey string) (*models.User, error) {
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
	delete(s.tokens, token)
	s.mutex.Unlock()
}

// GetUserFromToken 从 token 获取用户信息
func (s *AuthService) GetUserFromToken(token string) (*models.User, error) {
	tokenInfo, valid := s.ValidateToken(token)
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
		for token, tokenInfo := range s.tokens {
			if now.After(tokenInfo.Expiry) {
				delete(s.tokens, token)
			}
		}
		s.mutex.Unlock()
	}
}
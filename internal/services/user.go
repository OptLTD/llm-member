package services

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"swiflow-auth/internal/models"
)

type UserService struct {
	db *gorm.DB
}

func NewUserService(db *gorm.DB) *UserService {
	return &UserService{db: db}
}

// generateAPIKey 生成随机 API Key
func (s *UserService) generateAPIKey() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return "sk-" + hex.EncodeToString(bytes), nil
}

// Register 用户注册
func (s *UserService) Register(req *models.RegisterRequest) (*models.User, error) {
	// 检查用户名是否已存在
	var existingUser models.User
	if err := s.db.Where("username = ?", req.Username).First(&existingUser).Error; err == nil {
		return nil, fmt.Errorf("用户名已存在")
	}

	// 检查邮箱是否已存在
	if err := s.db.Where("email = ?", req.Email).First(&existingUser).Error; err == nil {
		return nil, fmt.Errorf("邮箱已存在")
	}

	// 生成密码哈希
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("密码加密失败: %v", err)
	}

	// 生成 API Key
	apiKey, err := s.generateAPIKey()
	if err != nil {
		return nil, fmt.Errorf("API Key 生成失败: %v", err)
	}

	// 创建用户
	user := models.User{
		Username:     req.Username,
		Email:        req.Email,
		Password:     string(hashedPassword),
		APIKey:       apiKey,
		IsActive:     true,
		IsAdmin:      false,
		DailyLimit:   1000,   // 默认每日限制
		MonthlyLimit: 10000,  // 默认每月限制
	}

	if err := s.db.Create(&user).Error; err != nil {
		return nil, fmt.Errorf("用户创建失败: %v", err)
	}

	// 清除密码字段
	user.Password = ""
	return &user, nil
}

// GetUserByUsername 根据用户名获取用户
func (s *UserService) GetUserByUsername(username string) (*models.User, error) {
	var user models.User
	if err := s.db.Where("username = ?", username).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// GetUserByAPIKey 根据 API Key 获取用户
func (s *UserService) GetUserByAPIKey(apiKey string) (*models.User, error) {
	var user models.User
	if err := s.db.Where("api_key = ? AND is_active = ?", apiKey, true).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// ValidatePassword 验证密码
func (s *UserService) ValidatePassword(hashedPassword, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	return err == nil
}

// GetAllUsers 获取所有用户（管理员功能）
func (s *UserService) GetAllUsers(page, pageSize int) ([]models.User, int64, error) {
	var users []models.User
	var total int64

	// 计算总数
	s.db.Model(&models.User{}).Count(&total)

	// 分页查询
	offset := (page - 1) * pageSize
	err := s.db.Offset(offset).Limit(pageSize).Find(&users).Error
	if err != nil {
		return nil, 0, err
	}

	// 清除密码字段
	for i := range users {
		users[i].Password = ""
	}

	return users, total, nil
}

// UpdateUser 更新用户信息
func (s *UserService) UpdateUser(userID uint, req *models.UpdateUserRequest) (*models.User, error) {
	var user models.User
	if err := s.db.First(&user, userID).Error; err != nil {
		return nil, fmt.Errorf("用户不存在")
	}

	// 更新字段
	if req.Email != "" {
		// 检查邮箱是否已被其他用户使用
		var existingUser models.User
		if err := s.db.Where("email = ? AND id != ?", req.Email, userID).First(&existingUser).Error; err == nil {
			return nil, fmt.Errorf("邮箱已被使用")
		}
		user.Email = req.Email
	}

	if req.IsActive != nil {
		user.IsActive = *req.IsActive
	}

	if req.DailyLimit != nil {
		user.DailyLimit = *req.DailyLimit
	}

	if req.MonthlyLimit != nil {
		user.MonthlyLimit = *req.MonthlyLimit
	}

	if err := s.db.Save(&user).Error; err != nil {
		return nil, fmt.Errorf("用户更新失败: %v", err)
	}

	user.Password = ""
	return &user, nil
}

// RegenerateAPIKey 重新生成 API Key
func (s *UserService) RegenerateAPIKey(userID uint) (*models.User, error) {
	var user models.User
	if err := s.db.First(&user, userID).Error; err != nil {
		return nil, fmt.Errorf("用户不存在")
	}

	// 生成新的 API Key
	apiKey, err := s.generateAPIKey()
	if err != nil {
		return nil, fmt.Errorf("API Key 生成失败: %v", err)
	}

	user.APIKey = apiKey
	if err := s.db.Save(&user).Error; err != nil {
		return nil, fmt.Errorf("API Key 更新失败: %v", err)
	}

	user.Password = ""
	return &user, nil
}

// UpdateUserStats 更新用户统计信息
func (s *UserService) UpdateUserStats(userID uint, tokensUsed int) error {
	return s.db.Model(&models.User{}).Where("id = ?", userID).Updates(map[string]interface{}{
		"total_requests": gorm.Expr("total_requests + ?", 1),
		"total_tokens":   gorm.Expr("total_tokens + ?", tokensUsed),
	}).Error
}

// CheckUserLimits 检查用户限制
func (s *UserService) CheckUserLimits(userID uint) error {
	var user models.User
	if err := s.db.First(&user, userID).Error; err != nil {
		return fmt.Errorf("用户不存在")
	}

	// 检查今日请求数
	today := time.Now().Format("2006-01-02")
	var todayRequests int64
	s.db.Model(&models.ChatLog{}).Where("user_id = ? AND DATE(created_at) = ?", userID, today).Count(&todayRequests)

	if todayRequests >= user.DailyLimit {
		return fmt.Errorf("已达到每日请求限制 (%d)", user.DailyLimit)
	}

	// 检查本月请求数
	thisMonth := time.Now().Format("2006-01")
	var monthlyRequests int64
	s.db.Model(&models.ChatLog{}).Where("user_id = ? AND DATE_FORMAT(created_at, '%Y-%m') = ?", userID, thisMonth).Count(&monthlyRequests)

	if monthlyRequests >= user.MonthlyLimit {
		return fmt.Errorf("已达到每月请求限制 (%d)", user.MonthlyLimit)
	}

	return nil
}

// DeleteUser 删除用户（软删除）
func (s *UserService) DeleteUser(userID uint) error {
	return s.db.Delete(&models.User{}, userID).Error
}

// GetUserByID 根据 ID 获取用户
func (s *UserService) GetUserByID(userID uint) (*models.User, error) {
	var user models.User
	if err := s.db.First(&user, userID).Error; err != nil {
		return nil, err
	}
	user.Password = ""
	return &user, nil
}

// ToggleUserStatus 切换用户状态
func (s *UserService) ToggleUserStatus(userID uint, isActive bool) error {
	var user models.User
	if err := s.db.First(&user, userID).Error; err != nil {
		return fmt.Errorf("用户不存在")
	}

	user.IsActive = isActive
	if err := s.db.Save(&user).Error; err != nil {
		return fmt.Errorf("用户状态更新失败: %v", err)
	}

	return nil
}
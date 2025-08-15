package service

import (
	"fmt"
	"time"

	"llm-member/internal/config"
	"llm-member/internal/model"
	"llm-member/internal/support"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type UserService struct {
	db *gorm.DB
}

func NewUserService() *UserService {
	return &UserService{db: config.GetDB()}
}

// createUser 创建用户
func (s *UserService) CreateUser(user *model.UserModel) error {
	return s.db.Create(user).Error
}

// GetUserByAPIKey 根据 API Key 获取用户
func (s *UserService) GetUserByAPIKey(apiKey string) (*model.UserModel, error) {
	var user model.UserModel
	if err := s.db.Where("api_key = ?", apiKey).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// GetUserByUsername 根据用户名获取用户
func (s *UserService) GetUserByUsername(username string) (*model.UserModel, error) {
	var user model.UserModel
	if err := s.db.Where("username = ?", username).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// GetUserByEmail 根据邮箱获取用户
func (s *UserService) GetUserByEmail(email string) (*model.UserModel, error) {
	var user model.UserModel
	if err := s.db.Where("email = ?", email).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// GetUserByPhone 根据手机号获取用户
func (s *UserService) GetUserByPhone(phone string) (*model.UserModel, error) {
	var user model.UserModel
	if err := s.db.Where("phone = ?", phone).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// ValidatePassword 验证密码
func (s *UserService) ValidatePassword(hashedPassword, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	return err == nil
}

// UpdateUser 更新用户信息
func (s *UserService) UpdateUser(req *model.UpdateUserRequest) (*model.UserModel, error) {
	var user model.UserModel
	if err := s.db.First(&user, req.UserId).Error; err != nil {
		return nil, fmt.Errorf("用户不存在")
	}

	// 更新字段
	if req.Email != "" {
		// 检查邮箱是否已被其他用户使用
		var existingUser model.UserModel
		if err := s.db.Where("email = ? AND id != ?", req.Email, req.UserId).First(&existingUser).Error; err == nil {
			return nil, fmt.Errorf("邮箱已被使用")
		}
		user.Email = req.Email
	}

	if req.Phone != "" {
		// 检查手机号是否已被其他用户使用
		var existingUser model.UserModel
		if err := s.db.Where("phone = ? AND id != ?", req.Phone, req.UserId).First(&existingUser).Error; err == nil {
			return nil, fmt.Errorf("手机号已被使用")
		}
		user.Phone = req.Phone
	}

	if req.IsActive != nil {
		user.IsActive = *req.IsActive
	}

	if req.PayPlan != nil {
		user.CurrPlan = *req.PayPlan
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

// RegenerateKey 重新生成 API Key
func (s *UserService) RegenerateKey(userID uint) (*model.UserModel, error) {
	var user model.UserModel
	if err := s.db.First(&user, userID).Error; err != nil {
		return nil, fmt.Errorf("用户不存在")
	}

	// 生成新的 API Key
	apiKey, err := support.GenerateAPIKey()
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
func (s *UserService) UpdateUserStats(userID uint64, tokensUsed int) error {
	var data = map[string]any{
		"total_requests": gorm.Expr("total_requests + ?", 1),
		"total_tokens":   gorm.Expr("total_tokens + ?", tokensUsed),
	}
	return s.db.Model(&model.UserModel{}).Where("id = ?", userID).Updates(data).Error
}

// CheckUserLimits 检查用户限制
func (s *UserService) CheckUserLimits(userID uint64) error {
	var user model.UserModel
	if err := s.db.First(&user, userID).Error; err != nil {
		return fmt.Errorf("用户不存在")
	}

	// 检查今日请求数
	today := time.Now().Format("2006-01-02")
	var todayRequests int64
	s.db.Model(&model.LlmLogModel{}).Where("user_id = ? AND DATE(created_at) = ?", userID, today).Count(&todayRequests)

	if todayRequests >= user.DailyLimit {
		return fmt.Errorf("已达到每日请求限制 (%d)", user.DailyLimit)
	}

	// 检查本月请求数
	thisMonth := time.Now().Format("2006-01")
	var monthlyRequests int64
	s.db.Model(&model.LlmLogModel{}).Where("user_id = ? AND DATE_FORMAT(created_at, '%Y-%m') = ?", userID, thisMonth).Count(&monthlyRequests)

	if monthlyRequests >= user.MonthlyLimit {
		return fmt.Errorf("已达到每月请求限制 (%d)", user.MonthlyLimit)
	}

	return nil
}

// DeleteUser 删除用户（软删除）
func (s *UserService) DeleteUser(userID uint) error {
	return s.db.Delete(&model.UserModel{}, userID).Error
}

// GetUserByID 根据 ID 获取用户
func (s *UserService) GetUserByID(userID uint64) (*model.UserModel, error) {
	var user model.UserModel
	if err := s.db.First(&user, userID).Error; err != nil {
		return nil, err
	}
	user.Password = ""
	return &user, nil
}

// GetUserByID 根据 ID 获取用户
func (s *UserService) GetUserOrders(userID uint64) ([]model.OrderModel, error) {
	var orders []model.OrderModel
	if err := s.db.Where("user_id = ?", userID).
		Order("created_at DESC").Limit(12).
		Find(&orders).Error; err != nil {
		return nil, err
	}
	return orders, nil
}

// ToggleUserStatus 切换用户状态
func (s *UserService) ToggleUserStatus(userID uint, isActive bool) error {
	var user model.UserModel
	if err := s.db.First(&user, userID).Error; err != nil {
		return fmt.Errorf("用户不存在")
	}

	user.IsActive = isActive
	if err := s.db.Save(&user).Error; err != nil {
		return fmt.Errorf("用户状态更新失败: %v", err)
	}

	return nil
}

// GetAllOrders 获取用户订单列表
func (s *UserService) QueryUsers(req *model.PaginateRequest) (*model.PaginateResponse, error) {
	var total int64
	var users []model.UserModel

	query := s.db.Model(&model.UserModel{})
	query.Where(req.Query).Order("created_at DESC")
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}

	// 获取分页数据
	offset := int((req.Page - 1) * req.Size)
	if err := query.Offset(offset).Limit(int(req.Size)).
		Find(&users).Error; err != nil {
		return nil, err
	}

	// 清除密码字段
	for i := range users {
		users[i].Password = ""
	}

	// 构造响应
	response := &model.PaginateResponse{
		Data: users, Page: req.Page, Size: req.Size, Total: total,
		Count: uint((total + int64(req.Size) - 1) / int64(req.Size)),
	}
	return response, nil
}

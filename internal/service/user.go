package service

import (
	"fmt"

	"llm-member/internal/config"
	"llm-member/internal/consts"
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
		return nil, consts.ErrUserNotFound
	}

	// 更新字段
	if req.Email != "" {
		// 检查邮箱是否已被其他用户使用
		var existingUser model.UserModel
		var query = s.db.Where("email = ? AND id != ?", req.Email, req.UserId)
		if err := query.First(&existingUser).Error; err == nil {
			return nil, consts.ErrEmailAlreadyUsed
		}
		user.Email = req.Email
	}

	if req.IsActive != nil {
		user.IsActive = *req.IsActive
	}

	if req.UserPlan != nil {
		user.UserPlan = *req.UserPlan
	}

	if err := s.db.Save(&user).Error; err != nil {
		return nil, fmt.Errorf("%w: %v", consts.ErrUserUpdateFailed, err)
	}

	user.Password = ""
	return &user, nil
}

// RegenerateKey 重新生成 API Key
func (s *UserService) RegenerateKey(userID uint64) (*model.UserModel, error) {
	var user model.UserModel
	if err := s.db.First(&user, userID).Error; err != nil {
		return nil, consts.ErrUserNotFound
	}

	// 生成新的 API Key
	apiKey, err := support.GenerateAPIKey()
	if err != nil {
		return nil, fmt.Errorf("%w: %v", consts.ErrAPIKeyGenerationFailed, err)
	}

	user.APIKey = apiKey
	if err := s.db.Save(&user).Error; err != nil {
		return nil, fmt.Errorf("%w: %v", consts.ErrAPIKeyUpdateFailed, err)
	}

	user.Password = ""
	return &user, nil
}

// DeleteUser 删除用户（软删除）
func (s *UserService) DeleteUser(userID uint64) error {
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
	var query = s.db.Where("user_id = ?", userID).
		Order("created_at DESC").Limit(12)
	if err := query.Find(&orders).Error; err != nil {
		return nil, err
	}
	return orders, nil
}

// ToggleUserStatus 切换用户状态
func (s *UserService) ToggleUserStatus(userID uint, isActive bool) error {
	var user model.UserModel
	if err := s.db.First(&user, userID).Error; err != nil {
		return consts.ErrUserNotFound
	}

	user.IsActive = isActive
	if err := s.db.Save(&user).Error; err != nil {
		return fmt.Errorf("%w: %v", consts.ErrUserStatusUpdateFailed, err)
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

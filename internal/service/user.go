package service

import (
	"errors"
	"fmt"

	"github.com/OptLTD/member/internal/config"
	"github.com/OptLTD/member/internal/consts"
	"github.com/OptLTD/member/internal/model"
	"github.com/OptLTD/member/internal/support"

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

// EnsureUserInProject 校验用户属于指定项目（0 表示不限制，兼容全局管理员）
func (s *UserService) EnsureUserInProject(userID, projectID uint64) error {
	if projectID == 0 {
		return nil
	}
	var u model.UserModel
	if err := s.db.Select("id", "project_id").First(&u, userID).Error; err != nil {
		return err
	}
	if u.ProjectID != 0 && u.ProjectID != projectID {
		return consts.ErrUserNotFound
	}
	return nil
}

// projectIDOf 根据用户 ID 查询所属项目（供内部埋点等使用）
func projectIDOf(userID uint64) uint64 {
	var u model.UserModel
	if err := config.GetDB().Select("project_id").First(&u, userID).Error; err != nil {
		return 0
	}
	return u.ProjectID
}

// CreateQuota 为用户创建配额记录（独立表 llm_user_quota）
func (s *UserService) CreateQuota(userID uint64, usage *model.ApiUsage, limit *model.ApiLimit) error {
	quota := &model.UserQuotaModel{
		UserID:   userID,
		ApiUsage: usage,
		ApiLimit: limit,
	}
	return s.db.Create(quota).Error
}

// GetQuota 获取用户配额记录
func (s *UserService) GetQuota(userID uint64) (*model.UserQuotaModel, error) {
	var quota model.UserQuotaModel
	if err := s.db.Where("user_id = ?", userID).First(&quota).Error; err != nil {
		return nil, err
	}
	return &quota, nil
}

// UpdateQuotaLimit 更新用户配额限制
func (s *UserService) UpdateQuotaLimit(userID uint64, limit *model.ApiLimit) error {
	return s.db.Model(&model.UserQuotaModel{}).
		Where("user_id = ?", userID).
		Update("api_limit", limit).Error
}

// UpdateQuotaUsage 更新用户用量
func (s *UserService) UpdateQuotaUsage(userID uint64, usage *model.ApiUsage) error {
	return s.db.Model(&model.UserQuotaModel{}).
		Where("user_id = ?", userID).
		Update("api_usage", usage).Error
}

// fillQuota 将独立表的配额数据填充到 UserModel（内存字段）
func (s *UserService) fillQuota(user *model.UserModel) error {
	if user == nil || user.ID == 0 {
		return nil
	}
	quota, err := s.GetQuota(user.ID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		}
		return err
	}
	user.ApiUsage = quota.ApiUsage
	user.ApiLimit = quota.ApiLimit
	return nil
}

// GetUserByAPIKey 根据 API Key 获取用户（支持 hash 存储与明文惰性迁移）
func (s *UserService) GetUserByAPIKey(apiKey string) (*model.UserModel, error) {
	var user model.UserModel
	hashed := support.HashAPIKey(apiKey)
	if err := s.db.Where("api_key = ?", hashed).First(&user).Error; err == nil {
		_ = s.fillQuota(&user)
		return &user, nil
	}
	// 兼容旧明文 key，命中后迁移为 hash
	if err := s.db.Where("api_key = ?", apiKey).First(&user).Error; err != nil {
		return nil, err
	}
	if !support.IsHashedAPIKey(user.APIKey) {
		_ = s.db.Model(&user).Update("api_key", hashed).Error
		user.APIKey = hashed
	}
	_ = s.fillQuota(&user)
	return &user, nil
}

// GetUserByUsername 根据用户名获取用户
func (s *UserService) GetUserByUsername(username string) (*model.UserModel, error) {
	var user model.UserModel
	if err := s.db.Where("username = ?", username).First(&user).Error; err != nil {
		return nil, err
	}
	_ = s.fillQuota(&user)
	return &user, nil
}

// GetUserByEmail 根据邮箱获取用户
func (s *UserService) GetUserByEmail(email string) (*model.UserModel, error) {
	var user model.UserModel
	if err := s.db.Where("email = ?", email).First(&user).Error; err != nil {
		return nil, err
	}
	_ = s.fillQuota(&user)
	return &user, nil
}

// GetUserByPhone 根据手机号获取用户
func (s *UserService) GetUserByPhone(phone string) (*model.UserModel, error) {
	var user model.UserModel
	if err := s.db.Where("phone = ?", phone).First(&user).Error; err != nil {
		return nil, err
	}
	_ = s.fillQuota(&user)
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

	// 生成新的 API Key（明文返回，入库存 hash）
	apiKey, err := support.GenerateAPIKey()
	if err != nil {
		return nil, fmt.Errorf("%w: %v", consts.ErrAPIKeyGenerationFailed, err)
	}

	user.APIKey = support.HashAPIKey(apiKey)
	if err := s.db.Save(&user).Error; err != nil {
		return nil, fmt.Errorf("%w: %v", consts.ErrAPIKeyUpdateFailed, err)
	}

	user.Password = ""
	user.APIKey = apiKey
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
	user.APIKey = ""
	_ = s.fillQuota(&user)
	return &user, nil
}

// HasAPIKey 检查用户是否已配置 API Key
func (s *UserService) HasAPIKey(userID uint64) (bool, error) {
	var user model.UserModel
	if err := s.db.Select("api_key").First(&user, userID).Error; err != nil {
		return false, err
	}
	return user.APIKey != "", nil
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

// GetAllOrders 获取用户订单列表（按项目过滤）
func (s *UserService) QueryUsers(req *model.PaginateRequest, projectID uint64) (*model.PaginateResponse, error) {
	var total int64
	var users []model.UserModel

	query := s.db.Model(&model.UserModel{})
	query.Where(req.Query)
	if projectID > 0 {
		query = query.Where("project_id = ?", projectID)
	}
	query = query.Order("created_at DESC")
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}

	// 获取分页数据
	offset := int((req.Page - 1) * req.Size)
	if err := query.Offset(offset).Limit(int(req.Size)).
		Find(&users).Error; err != nil {
		return nil, err
	}

	// 清除密码与 API Key hash，并批量填充配额
	var ids []uint64
	for i := range users {
		users[i].Password = ""
		users[i].APIKey = ""
		ids = append(ids, users[i].ID)
	}
	if len(ids) > 0 {
		var quotas []model.UserQuotaModel
		if err := s.db.Where("user_id IN ?", ids).Find(&quotas).Error; err == nil {
			m := make(map[uint64]*model.UserQuotaModel, len(quotas))
			for i := range quotas {
				m[quotas[i].UserID] = &quotas[i]
			}
			for i := range users {
				if q, ok := m[users[i].ID]; ok {
					users[i].ApiUsage = q.ApiUsage
					users[i].ApiLimit = q.ApiLimit
				}
			}
		}
	}

	// 构造响应
	response := &model.PaginateResponse{
		Data: users, Page: req.Page, Size: req.Size, Total: total,
		Count: uint((total + int64(req.Size) - 1) / int64(req.Size)),
	}
	return response, nil
}

package service

import (
	"gorm.io/gorm"

	"llm-member/internal/config"
	"llm-member/internal/model"
)

type LogService struct {
	db *gorm.DB
}

func NewLogService() *LogService {
	return &LogService{
		db: config.GetDB(),
	}
}

func (s *LogService) CreateLog(log *model.LlmLogModel) error {
	return s.db.Create(log).Error
}

// GetLogs 获取带筛选条件的日志
func (s *LogService) GetLogs(req *model.PaginateRequest) (*model.PaginateResponse, error) {
	var total int64
	var logs []model.LlmLogModel

	// 构建查询条件
	query := s.db.Model(&model.LlmLogModel{})
	if status, ok := req.Query["status"]; ok && status != "" {
		query = query.Where("status = ?", status)
	}

	if model, ok := req.Query["model"]; ok && model != "" {
		query = query.Where("model = ?", model)
	}

	if provider, ok := req.Query["provider"]; ok && provider != "" {
		query = query.Where("provider = ?", provider)
	}

	if ip, ok := req.Query["ip"]; ok && ip != "" {
		query = query.Where("client_ip LIKE ?", "%"+ip.(string)+"%")
	}

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}

	// 分页查询
	offset := int((req.Page - 1) * req.Size)
	if err := query.Order("created_at DESC").
		Offset(offset).Limit(int(req.Size)).
		Find(&logs).Error; err != nil {
		return nil, err
	}

	// 构造响应
	response := &model.PaginateResponse{
		Data: logs, Page: req.Page, Size: req.Size, Total: total,
		Count: uint((total + int64(req.Size) - 1) / int64(req.Size)),
	}
	return response, nil
}

func (s *LogService) DeleteLog(id uint) error {
	return s.db.Delete(&model.LlmLogModel{}, id).Error
}

// QueryUsage 查询消息的使用详情
func (s *LogService) QueryUsage(chat_id string) (*model.LlmLogModel, error) {
	var log model.LlmLogModel
	if err := s.db.Where("chat_id = ?", chat_id).First(&log).Error; err != nil {
		return nil, err
	}
	return &log, nil
}

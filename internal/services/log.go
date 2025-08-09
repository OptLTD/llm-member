package services

import (
	"encoding/json"
	"time"

	"gorm.io/gorm"

	"swiflow-auth/internal/models"
)

type LogService struct {
	db *gorm.DB
}

func NewLogService(db *gorm.DB) *LogService {
	return &LogService{
		db: db,
	}
}

func (s *LogService) CreateLog(log *models.ChatLog) error {
	return s.db.Create(log).Error
}

func (s *LogService) GetLogs(page, pageSize int) ([]models.ChatLog, int64, error) {
	var logs []models.ChatLog
	var total int64

	// 获取总数
	if err := s.db.Model(&models.ChatLog{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (page - 1) * pageSize
	if err := s.db.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&logs).Error; err != nil {
		return nil, 0, err
	}

	return logs, total, nil
}

func (s *LogService) DeleteLog(id uint) error {
	return s.db.Delete(&models.ChatLog{}, id).Error
}

func (s *LogService) GetStats() (*models.StatsResponse, error) {
	var stats models.StatsResponse

	// 总请求数
	s.db.Model(&models.ChatLog{}).Count(&stats.TotalRequests)

	// 总 token 数
	s.db.Model(&models.ChatLog{}).Select("COALESCE(SUM(tokens_used), 0)").Scan(&stats.TotalTokens)

	// 成功率
	var successCount int64
	s.db.Model(&models.ChatLog{}).Where("status = ?", "success").Count(&successCount)
	if stats.TotalRequests > 0 {
		stats.SuccessRate = float64(successCount) / float64(stats.TotalRequests) * 100
	}

	// 平均响应时间
	s.db.Model(&models.ChatLog{}).Select("COALESCE(AVG(duration), 0)").Scan(&stats.AvgDuration)

	// 今日请求数
	today := time.Now().Truncate(24 * time.Hour)
	s.db.Model(&models.ChatLog{}).Where("created_at >= ?", today).Count(&stats.RequestsToday)

	// 本周请求数
	weekStart := today.AddDate(0, 0, -int(today.Weekday()))
	s.db.Model(&models.ChatLog{}).Where("created_at >= ?", weekStart).Count(&stats.RequestsThisWeek)

	// 模型使用统计
	stats.ModelUsage = make(map[string]int64)
	var modelStats []struct {
		Model string
		Count int64
	}
	s.db.Model(&models.ChatLog{}).Select("model, COUNT(*) as count").Group("model").Scan(&modelStats)
	for _, stat := range modelStats {
		stats.ModelUsage[stat.Model] = stat.Count
	}

	// 提供商使用统计
	stats.ProviderUsage = make(map[string]int64)
	var providerStats []struct {
		Provider string
		Count    int64
	}
	s.db.Model(&models.ChatLog{}).Select("provider, COUNT(*) as count").Group("provider").Scan(&providerStats)
	for _, stat := range providerStats {
		stats.ProviderUsage[stat.Provider] = stat.Count
	}

	return &stats, nil
}

func (s *LogService) LogRequest(model, provider string, messages []models.ChatMessage, response *models.ChatResponse, duration time.Duration, status, errorMsg, clientIP, userAgent string) {
	messagesJSON, _ := json.Marshal(messages)
	responseJSON, _ := json.Marshal(response)

	var tokensUsed int
	if response != nil {
		tokensUsed = response.Usage.TotalTokens
	}

	log := &models.ChatLog{
		Model:      model,
		Provider:   provider,
		Messages:   string(messagesJSON),
		Response:   string(responseJSON),
		TokensUsed: tokensUsed,
		Duration:   duration.Milliseconds(),
		Status:     status,
		ErrorMsg:   errorMsg,
		ClientIP:   clientIP,
		UserAgent:  userAgent,
	}

	s.CreateLog(log)
}
package service

import (
	"time"

	"llm-member/internal/config"
	"llm-member/internal/model"

	"gorm.io/gorm"
)

type StatsService struct {
	db *gorm.DB
}

func NewStatsService() *StatsService {
	return &StatsService{
		db: config.GetDB(),
	}
}

// GetStats 获取完整的统计信息
func (s *StatsService) GetStats() (*model.StatsResponse, error) {
	stats := &model.StatsResponse{}

	// API统计
	if err := s.getAPIStats(stats); err != nil {
		return nil, err
	}

	// 会员统计
	if err := s.getMemberStats(stats); err != nil {
		return nil, err
	}

	// 订单统计
	if err := s.getOrderStats(stats); err != nil {
		return nil, err
	}

	return stats, nil
}

// getAPIStats 获取API相关统计
func (s *StatsService) getAPIStats(stats *model.StatsResponse) error {
	// 总请求数
	s.db.Model(&model.LlmLogModel{}).Count(&stats.TotalRequests)

	// 总token数和输入输出token数
	s.db.Model(&model.LlmLogModel{}).Select("COALESCE(SUM(tokens_used), 0)").Scan(&stats.TotalTokens)
	
	// 这里简化处理，实际应该从response中解析具体的输入输出token
	stats.InputTokens = stats.TotalTokens * 60 / 100  // 假设输入token占60%
	stats.OutputTokens = stats.TotalTokens * 40 / 100 // 假设输出token占40%

	// 成功率
	var successCount int64
	s.db.Model(&model.LlmLogModel{}).Where("status = ?", "success").Count(&successCount)
	if stats.TotalRequests > 0 {
		stats.SuccessRate = float64(successCount) / float64(stats.TotalRequests) * 100
	}

	// 今日请求数
	today := time.Now().Truncate(24 * time.Hour)
	s.db.Model(&model.LlmLogModel{}).Where("created_at >= ?", today).Count(&stats.RequestsToday)

	// 本周请求数
	weekStart := today.AddDate(0, 0, -int(today.Weekday()))
	s.db.Model(&model.LlmLogModel{}).Where("created_at >= ?", weekStart).Count(&stats.RequestsThisWeek)

	// 模型使用统计
	stats.ModelUsage = make(map[string]int64)
	var modelStats []struct {
		Model string
		Count int64
	}
	s.db.Model(&model.LlmLogModel{}).Select("the_model as model, COUNT(*) as count").Group("the_model").Scan(&modelStats)
	for _, stat := range modelStats {
		stats.ModelUsage[stat.Model] = stat.Count
	}

	// 提供商使用统计
	stats.ProviderUsage = make(map[string]int64)
	var providerStats []struct {
		Provider string
		Count    int64
	}
	s.db.Model(&model.LlmLogModel{}).Select("provider, COUNT(*) as count").Group("provider").Scan(&providerStats)
	for _, stat := range providerStats {
		stats.ProviderUsage[stat.Provider] = stat.Count
	}

	return nil
}

// getMemberStats 获取会员相关统计
func (s *StatsService) getMemberStats(stats *model.StatsResponse) error {
	// 总会员数
	s.db.Model(&model.UserModel{}).Count(&stats.TotalMembers)

	// 付费会员数（非basic套餐的用户）
	s.db.Model(&model.UserModel{}).Where("pay_plan != ?", "basic").Count(&stats.PaidMembers)

	// 本月新增会员数
	monthStart := time.Date(time.Now().Year(), time.Now().Month(), 1, 0, 0, 0, 0, time.Local)
	s.db.Model(&model.UserModel{}).Where("created_at >= ?", monthStart).Count(&stats.MonthlyNewMembers)

	// 本月新增付费会员数
	s.db.Model(&model.UserModel{}).Where("created_at >= ? AND pay_plan != ?", monthStart, "basic").Count(&stats.MonthlyNewPaidMembers)

	return nil
}

// getOrderStats 获取订单相关统计
func (s *StatsService) getOrderStats(stats *model.StatsResponse) error {
	// 总订单数
	s.db.Model(&model.OrderModel{}).Count(&stats.TotalOrders)

	// 成功订单数
	s.db.Model(&model.OrderModel{}).Where("status = ?", "succeed").Count(&stats.SuccessfulOrders)

	// 总收入
	s.db.Model(&model.OrderModel{}).Select("COALESCE(SUM(amount), 0)").Scan(&stats.TotalRevenue)

	// 成功订单收入
	s.db.Model(&model.OrderModel{}).Where("status = ?", "succeed").Select("COALESCE(SUM(amount), 0)").Scan(&stats.SuccessfulRevenue)

	// 本月收入
	monthStart := time.Date(time.Now().Year(), time.Now().Month(), 1, 0, 0, 0, 0, time.Local)
	s.db.Model(&model.OrderModel{}).Where("created_at >= ?", monthStart).Select("COALESCE(SUM(amount), 0)").Scan(&stats.MonthlyRevenue)

	// 本月成功订单收入
	s.db.Model(&model.OrderModel{}).Where("created_at >= ? AND status = ?", monthStart, "succeed").Select("COALESCE(SUM(amount), 0)").Scan(&stats.MonthlySuccessfulRevenue)

	// 本周收入
	now := time.Now()
	weekStart := now.AddDate(0, 0, -int(now.Weekday()))
	weekStart = time.Date(weekStart.Year(), weekStart.Month(), weekStart.Day(), 0, 0, 0, 0, time.Local)
	s.db.Model(&model.OrderModel{}).Where("created_at >= ?", weekStart).Select("COALESCE(SUM(amount), 0)").Scan(&stats.WeeklyRevenue)

	// 本周成功订单收入
	s.db.Model(&model.OrderModel{}).Where("created_at >= ? AND status = ?", weekStart, "succeed").Select("COALESCE(SUM(amount), 0)").Scan(&stats.WeeklySuccessfulRevenue)

	return nil
}
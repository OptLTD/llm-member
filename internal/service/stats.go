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
	// 获取当前自然月的开始和结束时间
	now := time.Now()
	monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	nextMonth := monthStart.AddDate(0, 1, 0)

	// 总请求数（当前月）
	s.db.Model(&model.LlmLogModel{}).
		Where("req_time >= ? AND req_time < ?", monthStart, nextMonth).
		Count(&stats.TotalRequests)

	// 使用JSON函数直接从all_usage字段提取token统计（当前月）
	// MySQL和SQLite都支持JSON_EXTRACT函数
	s.db.Model(&model.LlmLogModel{}).
		Where("status = ? AND req_time >= ? AND req_time < ?", "success", monthStart, nextMonth).
		Select("COALESCE(SUM(CAST(JSON_EXTRACT(all_usage, '$.total_tokens') AS SIGNED)), 0)").
		Scan(&stats.TotalTokens)

	s.db.Model(&model.LlmLogModel{}).
		Where("status = ? AND req_time >= ? AND req_time < ?", "success", monthStart, nextMonth).
		Select("COALESCE(SUM(CAST(JSON_EXTRACT(all_usage, '$.prompt_tokens') AS SIGNED)), 0)").
		Scan(&stats.InputTokens)

	s.db.Model(&model.LlmLogModel{}).
		Where("status = ? AND req_time >= ? AND req_time < ?", "success", monthStart, nextMonth).
		Select("COALESCE(SUM(CAST(JSON_EXTRACT(all_usage, '$.completion_tokens') AS SIGNED)), 0)").
		Scan(&stats.OutputTokens)

	// 平均响应时间（当前月）
	s.db.Model(&model.LlmLogModel{}).
		Where("status = ? AND req_time >= ? AND req_time < ?", "success", monthStart, nextMonth).
		Select("COALESCE(AVG(duration), 0)").
		Scan(&stats.AvgDuration)

	// 成功率（当前月）
	var successCount int64
	s.db.Model(&model.LlmLogModel{}).
		Where("status = ? AND req_time >= ? AND req_time < ?", "success", monthStart, nextMonth).
		Count(&successCount)
	if stats.TotalRequests > 0 {
		stats.SuccessRate = float64(successCount) / float64(stats.TotalRequests) * 100
	}

	// 今日请求数
	today := time.Now().Truncate(24 * time.Hour)
	s.db.Model(&model.LlmLogModel{}).Where("req_time >= ?", today).Count(&stats.RequestsToday)

	// 本周请求数
	weekStart := today.AddDate(0, 0, -int(today.Weekday()))
	s.db.Model(&model.LlmLogModel{}).Where("req_time >= ?", weekStart).Count(&stats.RequestsThisWeek)

	// 模型使用统计（当前月）
	stats.ModelUsage = make(map[string]int64)
	var modelStats []struct {
		Model string
		Count int64
	}
	s.db.Model(&model.LlmLogModel{}).
		Where("req_time >= ? AND req_time < ?", monthStart, nextMonth).
		Select("model, COUNT(*) as count").Group("model").Scan(&modelStats)
	for _, stat := range modelStats {
		stats.ModelUsage[stat.Model] = stat.Count
	}

	// 提供商使用统计（当前月）
	stats.ProviderUsage = make(map[string]int64)
	var providerStats []struct {
		Provider string
		Count    int64
	}
	s.db.Model(&model.LlmLogModel{}).
		Where("req_time >= ? AND req_time < ?", monthStart, nextMonth).
		Select("provider, COUNT(*) as count").
		Group("provider").
		Scan(&providerStats)
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
	s.db.Model(&model.UserModel{}).Where("user_plan != ?", "basic").Count(&stats.PaidMembers)

	// 本月新增会员数
	monthStart := time.Date(time.Now().Year(), time.Now().Month(), 1, 0, 0, 0, 0, time.Local)
	s.db.Model(&model.UserModel{}).Where("created_at >= ?", monthStart).Count(&stats.MonthlyNewMembers)

	// 本月新增付费会员数
	s.db.Model(&model.UserModel{}).Where("created_at >= ? AND user_plan != ?", monthStart, "basic").Count(&stats.MonthlyNewPaidMembers)

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

// UpdateUserStats 更新用户统计信息
func (s *StatsService) UpdateUserStats(user *model.UserModel) error {
	// 获取当前时间信息
	today := time.Now().Truncate(24 * time.Hour)
	monthStart := time.Date(
		time.Now().Year(), time.Now().Month(),
		1, 0, 0, 0, 0, time.Now().Location(),
	)

	// 初始化 ApiUsage 结构
	var apiUsage = model.ApiUsage{}

	// 统计总的使用量（所有时间）- 一次查询获取三个数据
	var totalStats struct {
		TotalTokens   uint64
		TotalRequests uint64
		TotalProjects uint64
	}

	// 使用子查询一次性获取总统计数据
	allAgg := `
		COUNT(*) as total_requests,
		COALESCE(SUM(CAST(JSON_EXTRACT(all_usage, '$.total_tokens') AS SIGNED)), 0) as total_tokens,
		(
			SELECT COUNT(DISTINCT proj_id) FROM llm_log 
			WHERE status = 'success' AND user_id = ? AND req_time >= ?
		) as total_projects
	`
	result := s.db.Model(&model.LlmLogModel{}).
		Where("status = ? AND user_id = ? AND req_time >= ?", "success", user.ID, monthStart).
		Select(allAgg, user.ID, monthStart).Scan(&totalStats)
	if result.Error != nil {
		return result.Error
	}

	// 统计时间范围内的使用量
	var periodStats struct {
		PeriodRequests uint64
		PeriodTokens   uint64
		PeriodProjects uint64
	}

	// 使用子查询一次性获取时间段统计数据
	periodAgg := `
		COUNT(*) as period_requests,
		COALESCE(SUM(CAST(JSON_EXTRACT(all_usage, '$.total_tokens') AS SIGNED)), 0) as period_tokens,
		(
			SELECT COUNT(DISTINCT proj_id)  FROM llm_log
			WHERE status = 'success' AND user_id = ? AND req_time >= ?
		) as period_projects
	`
	result = s.db.Model(&model.LlmLogModel{}).
		Where("status = 'success' AND user_id = ? AND req_time >= ?", user.ID, today).
		Select(periodAgg, user.ID, today).Scan(&periodStats)

	if result.Error != nil {
		return result.Error
	}

	// 赋值总统计数据
	apiUsage.TotalTokens = totalStats.TotalTokens
	apiUsage.TotalRequests = totalStats.TotalRequests
	apiUsage.TotalProjects = totalStats.TotalProjects

	// 赋值时间段统计数据（今日数据）
	apiUsage.TodayTokens = periodStats.PeriodTokens
	apiUsage.TodayRequests = periodStats.PeriodRequests
	apiUsage.TodayProjects = periodStats.PeriodProjects

	// 构建更新数据
	var query = s.db.Model(&model.UserModel{})
	var data = map[string]any{"api_usage": apiUsage}
	return query.Where("id = ?", user.ID).Updates(data).Error
}

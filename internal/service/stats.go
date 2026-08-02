package service

import (
	"fmt"
	"time"

	"github.com/OptLTD/member/internal/config"
	"github.com/OptLTD/member/internal/model"
	relay "github.com/OptLTD/member/llm-relay"

	"gorm.io/gorm"
)

// jsonIntExpr 按数据库方言从 JSON 文本字段提取整数值
func jsonIntExpr(db *gorm.DB, column, key string) string {
	return config.JSONIntExpr(db, column, key)
}

type StatsService struct {
	db *gorm.DB
}

func NewStatsService() *StatsService {
	return &StatsService{
		db: config.GetDB(),
	}
}

// GetStats 获取完整的统计信息（按项目过滤）
func (s *StatsService) GetStats(projectID uint64) (*model.StatsResponse, error) {
	stats := &model.StatsResponse{}

	// API统计
	if err := s.getAPIStats(stats, projectID); err != nil {
		return nil, err
	}

	// 会员统计
	if err := s.getMemberStats(stats, projectID); err != nil {
		return nil, err
	}

	// 订单统计
	if err := s.getOrderStats(stats, projectID); err != nil {
		return nil, err
	}

	if config.HasModule(config.ModuleAnalytics) {
		if overview, err := NewAnalyticsService().Overview(projectID, 0); err == nil {
			if v, ok := overview["todayPV"].(int64); ok {
				stats.AnalyticsTodayPV = v
			}
			if v, ok := overview["weekEvents"].(int64); ok {
				stats.AnalyticsWeekEvents = v
			}
		}
	}

	return stats, nil
}

// getAPIStats 获取API相关统计
func (s *StatsService) getAPIStats(stats *model.StatsResponse, projectID uint64) error {
	// 获取当前自然月的开始和结束时间
	now := time.Now()
	monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	nextMonth := monthStart.AddDate(0, 1, 0)

	// 项目范围过滤
	scoped := func() *gorm.DB {
		q := s.db.Model(&model.LlmLogModel{})
		if projectID > 0 {
			q = q.Where("project_id = ?", projectID)
		}
		return q
	}

	// 总请求数（当前月）
	scoped().
		Where("req_time >= ? AND req_time < ?", monthStart, nextMonth).
		Count(&stats.TotalRequests)

	// 使用JSON函数直接从all_usage字段提取token统计（当前月）
	scoped().
		Where("status = ? AND req_time >= ? AND req_time < ?", "success", monthStart, nextMonth).
		Select(fmt.Sprintf("COALESCE(SUM(%s), 0)", jsonIntExpr(s.db, "all_usage", "totalTokens"))).
		Scan(&stats.TotalTokens)

	scoped().
		Where("status = ? AND req_time >= ? AND req_time < ?", "success", monthStart, nextMonth).
		Select(fmt.Sprintf("COALESCE(SUM(%s), 0)", jsonIntExpr(s.db, "all_usage", "promptTokens"))).
		Scan(&stats.InputTokens)

	scoped().
		Where("status = ? AND req_time >= ? AND req_time < ?", "success", monthStart, nextMonth).
		Select(fmt.Sprintf("COALESCE(SUM(%s), 0)", jsonIntExpr(s.db, "all_usage", "completionTokens"))).
		Scan(&stats.OutputTokens)

	// 平均响应时间（当前月）
	scoped().
		Where("status = ? AND req_time >= ? AND req_time < ?", "success", monthStart, nextMonth).
		Select("COALESCE(AVG(duration), 0)").
		Scan(&stats.AvgDuration)

	// 成功率（当前月）
	var successCount int64
	scoped().
		Where("status = ? AND req_time >= ? AND req_time < ?", "success", monthStart, nextMonth).
		Count(&successCount)
	if stats.TotalRequests > 0 {
		stats.SuccessRate = float64(successCount) / float64(stats.TotalRequests) * 100
	}

	// 今日请求数
	today := time.Now().Truncate(24 * time.Hour)
	scoped().Where("req_time >= ?", today).Count(&stats.RequestsToday)

	// 本周请求数
	weekStart := today.AddDate(0, 0, -int(today.Weekday()))
	scoped().Where("req_time >= ?", weekStart).Count(&stats.RequestsThisWeek)

	// 模型使用统计（当前月）
	stats.ModelUsage = make(map[string]int64)
	var modelStats []struct {
		Model string
		Count int64
	}
	scoped().
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
	scoped().
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
func (s *StatsService) getMemberStats(stats *model.StatsResponse, projectID uint64) error {
	scoped := func() *gorm.DB {
		q := s.db.Model(&model.UserModel{})
		if projectID > 0 {
			q = q.Where("project_id = ?", projectID)
		}
		return q
	}
	// 总会员数
	scoped().Count(&stats.TotalMembers)

	// 付费会员数（非basic套餐的用户）
	scoped().Where("user_plan != ?", "basic").Count(&stats.PaidMembers)

	// 本月新增会员数
	monthStart := time.Date(time.Now().Year(), time.Now().Month(), 1, 0, 0, 0, 0, time.Local)
	scoped().Where("created_at >= ?", monthStart).Count(&stats.MonthlyNewMembers)

	// 本月新增付费会员数
	scoped().Where("created_at >= ? AND user_plan != ?", monthStart, "basic").Count(&stats.MonthlyNewPaidMembers)

	return nil
}

// getOrderStats 获取订单相关统计
func (s *StatsService) getOrderStats(stats *model.StatsResponse, projectID uint64) error {
	scoped := func() *gorm.DB {
		q := s.db.Model(&model.OrderModel{})
		if projectID > 0 {
			q = q.Where("project_id = ?", projectID)
		}
		return q
	}
	// 总订单数
	scoped().Count(&stats.TotalOrders)

	// 成功订单数
	scoped().Where("status = ?", "succeed").Count(&stats.SuccessfulOrders)

	// 总收入
	scoped().Select("COALESCE(SUM(amount), 0)").Scan(&stats.TotalRevenue)

	// 成功订单收入
	scoped().Where("status = ?", "succeed").Select("COALESCE(SUM(amount), 0)").Scan(&stats.SuccessfulRevenue)

	// 本月收入
	monthStart := time.Date(time.Now().Year(), time.Now().Month(), 1, 0, 0, 0, 0, time.Local)
	scoped().Where("created_at >= ?", monthStart).Select("COALESCE(SUM(amount), 0)").Scan(&stats.MonthlyRevenue)

	// 本月成功订单收入
	scoped().Where("created_at >= ? AND status = ?", monthStart, "succeed").Select("COALESCE(SUM(amount), 0)").Scan(&stats.MonthlySuccessfulRevenue)

	// 本周收入
	now := time.Now()
	weekStart := now.AddDate(0, 0, -int(now.Weekday()))
	weekStart = time.Date(weekStart.Year(), weekStart.Month(), weekStart.Day(), 0, 0, 0, 0, time.Local)
	scoped().Where("created_at >= ?", weekStart).Select("COALESCE(SUM(amount), 0)").Scan(&stats.WeeklyRevenue)

	// 本周成功订单收入
	scoped().Where("created_at >= ? AND status = ?", weekStart, "succeed").Select("COALESCE(SUM(amount), 0)").Scan(&stats.WeeklySuccessfulRevenue)

	return nil
}

// UpdateUserStats 增量刷新已由 QuotaService.Commit 完成；保留接口供管理端手动触发轻量日切
func (s *StatsService) UpdateUserStats(user *model.UserModel) error {
	if user == nil {
		return nil
	}
	user.ApiUsage = relay.EnsureUsageReady(user.ApiUsage)
	return s.db.Model(&model.UserQuotaModel{}).
		Where("user_id = ?", user.ID).
		Update("api_usage", user.ApiUsage).Error
}

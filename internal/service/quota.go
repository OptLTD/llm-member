package service

import (
	"context"
	"fmt"
	"time"

	"github.com/OptLTD/member/internal/config"
	"github.com/OptLTD/member/internal/consts"
	"github.com/OptLTD/member/internal/model"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// QuotaReservation 一次预扣记录，用于 Commit / Rollback
type QuotaReservation struct {
	UserID   uint64
	Method   string
	Amount   uint64
	DayKey   string
	MonthKey string
	UsedRedis bool
}

type QuotaService struct {
	db  *gorm.DB
	rdb *redis.Client
}

func NewQuotaService() *QuotaService {
	return &QuotaService{
		db:  config.GetDB(),
		rdb: config.GetRedis(),
	}
}

// DefaultBasicLimit 在套餐未配置时的保守默认配额
func DefaultBasicLimit() *model.ApiLimit {
	return &model.ApiLimit{
		LimitMethod:     "requests",
		DailyRequests:   100,
		MonthlyRequests: 1000,
		ExpireDays:      1,
	}
}

// EnsureUsageReady 确保 ApiUsage 存在并完成日切
func EnsureUsageReady(user *model.UserModel) {
	today := time.Now().Format("2006-01-02")
	if user.ApiUsage == nil {
		user.ApiUsage = &model.ApiUsage{DayStamp: today}
		return
	}
	if user.ApiUsage.DayStamp != today {
		user.ApiUsage.TodayTokens = 0
		user.ApiUsage.TodayRequests = 0
		user.ApiUsage.TodayProjects = 0
		user.ApiUsage.DayStamp = today
	}
}

func (q *QuotaService) dailyLimit(limit *model.ApiLimit) uint64 {
	switch limit.LimitMethod {
	case "tokens":
		return limit.DailyTokens
	case "projects":
		return limit.DailyProjects
	default:
		return limit.DailyRequests
	}
}

func (q *QuotaService) monthlyLimit(limit *model.ApiLimit) uint64 {
	switch limit.LimitMethod {
	case "tokens":
		return limit.MonthlyTokens
	case "projects":
		return limit.MonthlyProjects
	default:
		return limit.MonthlyRequests
	}
}

func (q *QuotaService) currentDay(usage *model.ApiUsage, method string) uint64 {
	switch method {
	case "tokens":
		return usage.TodayTokens
	case "projects":
		return usage.TodayProjects
	default:
		return usage.TodayRequests
	}
}

func (q *QuotaService) currentMonth(usage *model.ApiUsage, method string) uint64 {
	switch method {
	case "tokens":
		return usage.TotalTokens
	case "projects":
		return usage.TotalProjects
	default:
		return usage.TotalRequests
	}
}

func (q *QuotaService) applyDelta(usage *model.ApiUsage, method string, delta int64) {
	add := func(v *uint64, d int64) {
		if d >= 0 {
			*v += uint64(d)
			return
		}
		sub := uint64(-d)
		if *v > sub {
			*v -= sub
		} else {
			*v = 0
		}
	}
	switch method {
	case "tokens":
		add(&usage.TodayTokens, delta)
		add(&usage.TotalTokens, delta)
	case "projects":
		add(&usage.TodayProjects, delta)
		add(&usage.TotalProjects, delta)
	default:
		add(&usage.TodayRequests, delta)
		add(&usage.TotalRequests, delta)
	}
}

var quotaReserveScript = redis.NewScript(`
if redis.call('EXISTS', KEYS[1]) == 0 then
  redis.call('SET', KEYS[1], ARGV[4], 'EX', ARGV[6])
end
if redis.call('EXISTS', KEYS[2]) == 0 then
  redis.call('SET', KEYS[2], ARGV[5], 'EX', ARGV[7])
end
local day = redis.call('INCRBY', KEYS[1], ARGV[1])
local month = redis.call('INCRBY', KEYS[2], ARGV[1])
local dayLimit = tonumber(ARGV[2])
local monthLimit = tonumber(ARGV[3])
if (dayLimit > 0 and day > dayLimit) or (monthLimit > 0 and month > monthLimit) then
  redis.call('INCRBY', KEYS[1], -ARGV[1])
  redis.call('INCRBY', KEYS[2], -ARGV[1])
  return 0
end
return 1
`)

// Reserve 请求前预扣；amount 为预估增量（requests/projects 通常为 1）
func (q *QuotaService) Reserve(user *model.UserModel, amount uint64) (*QuotaReservation, error) {
	if user.ApiLimit == nil {
		return nil, consts.ErrQuotaNotConfigured
	}
	if amount == 0 {
		amount = 1
	}
	EnsureUsageReady(user)
	method := user.ApiLimit.LimitMethod
	if method == "" {
		method = "requests"
	}

	res := &QuotaReservation{
		UserID: user.ID,
		Method: method,
		Amount: amount,
	}

	dayLimit := q.dailyLimit(user.ApiLimit)
	monthLimit := q.monthlyLimit(user.ApiLimit)

	if q.rdb != nil {
		now := time.Now()
		dayStr := now.Format("2006-01-02")
		monthStr := now.Format("2006-01")
		res.DayKey = fmt.Sprintf("quota:%d:day:%s:%s", user.ID, dayStr, method)
		res.MonthKey = fmt.Sprintf("quota:%d:month:%s:%s", user.ID, monthStr, method)
		res.UsedRedis = true

		dayEnd := time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, now.Location())
		monthEnd := time.Date(now.Year(), now.Month()+1, 1, 0, 0, 0, 0, now.Location())
		dayTTL := int(dayEnd.Sub(now).Seconds())
		monthTTL := int(monthEnd.Sub(now).Seconds())
		if dayTTL < 1 {
			dayTTL = 86400
		}
		if monthTTL < 1 {
			monthTTL = 86400 * 31
		}

		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		ok, err := quotaReserveScript.Run(ctx, q.rdb, []string{res.DayKey, res.MonthKey},
			amount, dayLimit, monthLimit,
			q.currentDay(user.ApiUsage, method),
			q.currentMonth(user.ApiUsage, method),
			dayTTL, monthTTL,
		).Int()
		if err != nil {
			return nil, fmt.Errorf("quota redis reserve: %w", err)
		}
		if ok != 1 {
			return nil, consts.ErrQuotaExceeded
		}
		return res, nil
	}

	// DB 行锁兜底
	err := q.db.Transaction(func(tx *gorm.DB) error {
		var locked model.UserModel
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			First(&locked, user.ID).Error; err != nil {
			return err
		}
		EnsureUsageReady(&locked)
		if locked.ApiLimit == nil {
			locked.ApiLimit = user.ApiLimit
		}
		dayCur := q.currentDay(locked.ApiUsage, method)
		monthCur := q.currentMonth(locked.ApiUsage, method)
		if (dayLimit > 0 && dayCur+amount > dayLimit) ||
			(monthLimit > 0 && monthCur+amount > monthLimit) {
			return consts.ErrQuotaExceeded
		}
		q.applyDelta(locked.ApiUsage, method, int64(amount))
		if err := tx.Model(&locked).Update("api_usage", locked.ApiUsage).Error; err != nil {
			return err
		}
		user.ApiUsage = locked.ApiUsage
		return nil
	})
	if err != nil {
		return nil, err
	}
	return res, nil
}

// Commit 按实际用量结算：tokens 模式下调整预扣差额，并同步 DB
func (q *QuotaService) Commit(user *model.UserModel, res *QuotaReservation, actual uint64) error {
	if res == nil {
		return nil
	}
	if actual == 0 {
		actual = res.Amount
	}
	delta := int64(actual) - int64(res.Amount)

	if res.UsedRedis && q.rdb != nil && delta != 0 {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		q.rdb.IncrBy(ctx, res.DayKey, delta)
		q.rdb.IncrBy(ctx, res.MonthKey, delta)
	}

	return q.db.Transaction(func(tx *gorm.DB) error {
		var locked model.UserModel
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			First(&locked, user.ID).Error; err != nil {
			return err
		}
		EnsureUsageReady(&locked)
		if res.UsedRedis {
			// Redis 路径：预扣未写 DB，此处写入实际用量
			q.applyDelta(locked.ApiUsage, res.Method, int64(actual))
		} else if delta != 0 {
			// DB 路径已预扣 Amount，补差额
			q.applyDelta(locked.ApiUsage, res.Method, delta)
		}
		if err := tx.Model(&locked).Update("api_usage", locked.ApiUsage).Error; err != nil {
			return err
		}
		user.ApiUsage = locked.ApiUsage
		return nil
	})
}

// Rollback 失败时退回预扣
func (q *QuotaService) Rollback(user *model.UserModel, res *QuotaReservation) error {
	if res == nil {
		return nil
	}
	if res.UsedRedis && q.rdb != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		q.rdb.IncrBy(ctx, res.DayKey, -int64(res.Amount))
		q.rdb.IncrBy(ctx, res.MonthKey, -int64(res.Amount))
		return nil
	}
	return q.db.Transaction(func(tx *gorm.DB) error {
		var locked model.UserModel
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			First(&locked, user.ID).Error; err != nil {
			return err
		}
		EnsureUsageReady(&locked)
		q.applyDelta(locked.ApiUsage, res.Method, -int64(res.Amount))
		if err := tx.Model(&locked).Update("api_usage", locked.ApiUsage).Error; err != nil {
			return err
		}
		user.ApiUsage = locked.ApiUsage
		return nil
	})
}

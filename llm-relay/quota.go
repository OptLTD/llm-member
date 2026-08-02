package relay

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// 配额相关错误
var (
	ErrQuotaNotConfigured = errors.New("账户未配置用量限制")
	ErrQuotaExceeded      = errors.New("已超出用量配额")

	ErrDailyTokenLimitReached     = errors.New("已达到每日 Token 限制")
	ErrMonthlyTokenLimitReached   = errors.New("已达到每月 Token 限制")
	ErrDailyRequestLimitReached   = errors.New("已达到每日请求限制")
	ErrMonthlyRequestLimitReached = errors.New("已达到每月请求限制")
	ErrDailyProjectLimitReached   = errors.New("已达到每日项目限制")
	ErrMonthlyProjectLimitReached = errors.New("已达到每月项目限制")
)

// QuotaUser 映射用户配额独立表（llm_user_quota）。
// relay 通过 user_id 读写配额字段，不依赖应用侧用户模型。
type QuotaUser struct {
	ID       uint64    `gorm:"primarykey"`
	UserID   uint64    `gorm:"column:user_id;index;not null"`
	ApiUsage *ApiUsage `gorm:"column:api_usage;serializer:json"`
	ApiLimit *ApiLimit `gorm:"column:api_limit;serializer:json"`
}

func (QuotaUser) TableName() string { return "llm_user_quota" }

// QuotaReservation 一次预扣记录，用于 Commit / Rollback；实现 relay.Reservation
type QuotaReservation struct {
	UserID    uint64
	Method    string
	Amount    uint64
	DayKey    string
	MonthKey  string
	UsedRedis bool
}

// QuotaService 使用配额服务（实现 Quota 接口）。
type QuotaService struct {
	db  *gorm.DB
	rdb *redis.Client
}

// NewQuotaService 创建配额服务；db 用于行锁兜底与持久化，rdb 可选（Redis 原子预扣）。
func NewQuotaService(db *gorm.DB, rdb *redis.Client) *QuotaService {
	return &QuotaService{db: db, rdb: rdb}
}

// DefaultBasicLimit 在套餐未配置时的保守默认配额
func DefaultBasicLimit() *ApiLimit {
	return &ApiLimit{
		LimitMethod:     "requests",
		DailyRequests:   100,
		MonthlyRequests: 1000,
		ExpireDays:      1,
	}
}

// EnsureUsageReady 确保 ApiUsage 已初始化并完成日切；返回（可能新建的）usage。
func EnsureUsageReady(usage *ApiUsage) *ApiUsage {
	today := time.Now().Format("2006-01-02")
	if usage == nil {
		return &ApiUsage{DayStamp: today}
	}
	if usage.DayStamp != today {
		usage.TodayTokens = 0
		usage.TodayRequests = 0
		usage.TodayProjects = 0
		usage.DayStamp = today
	}
	return usage
}

func (q *QuotaService) dailyLimit(limit *ApiLimit) uint64 {
	switch limit.LimitMethod {
	case "tokens":
		return limit.DailyTokens
	case "projects":
		return limit.DailyProjects
	default:
		return limit.DailyRequests
	}
}

func (q *QuotaService) monthlyLimit(limit *ApiLimit) uint64 {
	switch limit.LimitMethod {
	case "tokens":
		return limit.MonthlyTokens
	case "projects":
		return limit.MonthlyProjects
	default:
		return limit.MonthlyRequests
	}
}

func (q *QuotaService) currentDay(usage *ApiUsage, method string) uint64 {
	switch method {
	case "tokens":
		return usage.TodayTokens
	case "projects":
		return usage.TodayProjects
	default:
		return usage.TodayRequests
	}
}

func (q *QuotaService) currentMonth(usage *ApiUsage, method string) uint64 {
	switch method {
	case "tokens":
		return usage.TotalTokens
	case "projects":
		return usage.TotalProjects
	default:
		return usage.TotalRequests
	}
}

func (q *QuotaService) applyDelta(usage *ApiUsage, method string, delta int64) {
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

func (q *QuotaService) loadUser(user User) (*QuotaUser, error) {
	var qu QuotaUser
	if err := q.db.Where("user_id = ?", user.GetID()).First(&qu).Error; err != nil {
		return nil, fmt.Errorf("%w: %v", ErrQuotaNotConfigured, err)
	}
	return &qu, nil
}

// CheckUsage 检查用户使用限制（fail-closed：无 ApiLimit 则拒绝）
func (q *QuotaService) CheckUsage(user User) error {
	qu, err := q.loadUser(user)
	if err != nil {
		return err
	}
	if qu.ApiLimit == nil {
		return ErrQuotaNotConfigured
	}
	qu.ApiUsage = EnsureUsageReady(qu.ApiUsage)
	if qu.ApiUsage == nil {
		return nil
	}

	usage, limit := qu.ApiUsage, qu.ApiLimit
	switch limit.LimitMethod {
	case "tokens":
		if limit.DailyTokens > 0 && usage.TodayTokens >= limit.DailyTokens {
			return fmt.Errorf("%w (%d/%d)", ErrDailyTokenLimitReached, usage.TodayTokens, limit.DailyTokens)
		}
		if limit.MonthlyTokens > 0 && usage.TotalTokens >= limit.MonthlyTokens {
			return fmt.Errorf("%w (%d/%d)", ErrMonthlyTokenLimitReached, usage.TotalTokens, limit.MonthlyTokens)
		}
	case "requests":
		if limit.DailyRequests > 0 && usage.TodayRequests >= limit.DailyRequests {
			return fmt.Errorf("%w (%d/%d)", ErrDailyRequestLimitReached, usage.TodayRequests, limit.DailyRequests)
		}
		if limit.MonthlyRequests > 0 && usage.TotalRequests >= limit.MonthlyRequests {
			return fmt.Errorf("%w (%d/%d)", ErrMonthlyRequestLimitReached, usage.TotalRequests, limit.MonthlyRequests)
		}
	case "projects":
		if limit.DailyProjects > 0 && usage.TodayProjects >= limit.DailyProjects {
			return fmt.Errorf("%w (%d/%d)", ErrDailyProjectLimitReached, usage.TodayProjects, limit.DailyProjects)
		}
		if limit.MonthlyProjects > 0 && usage.TotalProjects >= limit.MonthlyProjects {
			return fmt.Errorf("%w (%d/%d)", ErrMonthlyProjectLimitReached, usage.TotalProjects, limit.MonthlyProjects)
		}
	}
	return nil
}

// Reserve 请求前预扣；amount 为预估增量（requests/projects 通常为 1）
func (q *QuotaService) Reserve(user User, amount uint64) (Reservation, error) {
	qu, err := q.loadUser(user)
	if err != nil {
		return nil, err
	}
	if qu.ApiLimit == nil {
		return nil, ErrQuotaNotConfigured
	}
	if amount == 0 {
		amount = 1
	}
	qu.ApiUsage = EnsureUsageReady(qu.ApiUsage)
	method := qu.ApiLimit.LimitMethod
	if method == "" {
		method = "requests"
	}

	res := &QuotaReservation{
		UserID: user.GetID(),
		Method: method,
		Amount: amount,
	}

	dayLimit := q.dailyLimit(qu.ApiLimit)
	monthLimit := q.monthlyLimit(qu.ApiLimit)

	if q.rdb != nil {
		now := time.Now()
		dayStr := now.Format("2006-01-02")
		monthStr := now.Format("2006-01")
		res.DayKey = fmt.Sprintf("quota:%d:day:%s:%s", user.GetID(), dayStr, method)
		res.MonthKey = fmt.Sprintf("quota:%d:month:%s:%s", user.GetID(), monthStr, method)
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
			q.currentDay(qu.ApiUsage, method),
			q.currentMonth(qu.ApiUsage, method),
			dayTTL, monthTTL,
		).Int()
		if err != nil {
			return nil, fmt.Errorf("quota redis reserve: %w", err)
		}
		if ok != 1 {
			return nil, ErrQuotaExceeded
		}
		return res, nil
	}

	// DB 行锁兜底
	err = q.db.Transaction(func(tx *gorm.DB) error {
		var locked QuotaUser
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("user_id = ?", user.GetID()).First(&locked).Error; err != nil {
			return err
		}
		locked.ApiUsage = EnsureUsageReady(locked.ApiUsage)
		if locked.ApiLimit == nil {
			locked.ApiLimit = qu.ApiLimit
		}
		dayCur := q.currentDay(locked.ApiUsage, method)
		monthCur := q.currentMonth(locked.ApiUsage, method)
		if (dayLimit > 0 && dayCur+amount > dayLimit) ||
			(monthLimit > 0 && monthCur+amount > monthLimit) {
			return ErrQuotaExceeded
		}
		q.applyDelta(locked.ApiUsage, method, int64(amount))
		if err := tx.Model(&QuotaUser{}).
			Where("user_id = ?", user.GetID()).
			Update("api_usage", locked.ApiUsage).Error; err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return res, nil
}

// Commit 按实际用量结算：tokens 模式下调整预扣差额，并同步 DB
func (q *QuotaService) Commit(user User, reservation Reservation, actual uint64) error {
	res, ok := reservation.(*QuotaReservation)
	if !ok || res == nil {
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
		var locked QuotaUser
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("user_id = ?", user.GetID()).First(&locked).Error; err != nil {
			return err
		}
		locked.ApiUsage = EnsureUsageReady(locked.ApiUsage)
		if res.UsedRedis {
			// Redis 路径：预扣未写 DB，此处写入实际用量
			q.applyDelta(locked.ApiUsage, res.Method, int64(actual))
		} else if delta != 0 {
			// DB 路径已预扣 Amount，补差额
			q.applyDelta(locked.ApiUsage, res.Method, delta)
		}
		if err := tx.Model(&QuotaUser{}).
			Where("user_id = ?", user.GetID()).
			Update("api_usage", locked.ApiUsage).Error; err != nil {
			return err
		}
		return nil
	})
}

// Rollback 失败时退回预扣
func (q *QuotaService) Rollback(user User, reservation Reservation) error {
	res, ok := reservation.(*QuotaReservation)
	if !ok || res == nil {
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
		var locked QuotaUser
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("user_id = ?", user.GetID()).First(&locked).Error; err != nil {
			return err
		}
		locked.ApiUsage = EnsureUsageReady(locked.ApiUsage)
		q.applyDelta(locked.ApiUsage, res.Method, -int64(res.Amount))
		if err := tx.Model(&QuotaUser{}).
			Where("user_id = ?", user.GetID()).
			Update("api_usage", locked.ApiUsage).Error; err != nil {
			return err
		}
		return nil
	})
}

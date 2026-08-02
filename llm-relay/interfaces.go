package relay

import "time"

// User 中继所需的最小用户信息；应用侧（如 *model.UserModel）实现该接口。
// 仅暴露中继逻辑需要读取的字段，避免 relay 依赖具体数据模型。
type User interface {
	GetID() uint64
	// GetProjectID 返回用户所属项目 ID
	GetProjectID() uint64
	// GetLimitMethod 返回用户套餐的配额方式（tokens/requests/projects）；未配置时返回空字符串
	GetLimitMethod() string
}

// Reservation 配额预扣的不透明句柄，由 Quota 实现方创建，并在 Commit / Rollback 时原样传回。
type Reservation interface{}

// Quota 使用配额接口，由应用侧实现（如 *service.QuotaService）。
type Quota interface {
	// CheckUsage 请求前检查用户配额是否可用
	CheckUsage(user User) error
	// Reserve 预扣配额，amount 为预估增量（requests/projects 通常为 1）
	Reserve(user User, amount uint64) (Reservation, error)
	// Commit 按实际用量结算
	Commit(user User, reservation Reservation, actual uint64) error
	// Rollback 失败时退回预扣
	Rollback(user User, reservation Reservation) error
}

// LogEntry 中继请求日志条目，由应用侧 Logger 持久化。
type LogEntry struct {
	ProjectID uint64
	UserID    uint64
	ChatID    string
	ProjID    string
	Model     string
	Provider  string
	Messages  any
	Response  any
	AllUsage  any
	Duration  int64
	Status    string
	ErrorMsg  string
	ClientIP  string
	UserAgent string
	ReqTime   time.Time
}

// Logger 日志记录器接口，由应用侧实现（如 *service.LogService），或使用包内默认实现。
type Logger interface {
	CreateLog(entry *LogEntry) error
}

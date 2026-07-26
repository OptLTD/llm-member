package config

import (
	"encoding/json"
	"strings"
	"sync"
)

const (
	ModuleMember    = "member"
	ModuleDevice    = "device"
	ModuleAnalytics = "analytics"
	ModuleCoupon    = "coupon"
	ModuleReferral  = "referral"
	ModuleSurvey    = "survey"
	ModuleFeedback  = "feedback"
	ModuleLLMRelay  = "llm-relay"

	// ConfigKeyModules 数据库中存储已启用模块的配置键
	ConfigKeyModules = "app.modules"
)

var (
	globalCfg *Config
	moduleMu  sync.RWMutex
	moduleSet map[string]struct{}
)

// AppConfig returns loaded application config (may be nil before Load).
func AppConfig() *Config {
	return globalCfg
}

func setGlobalConfig(cfg *Config) {
	globalCfg = cfg
	SetModulesCSV(cfg.Modules)
}

// AllModules 全部可选模块（固定顺序）
func AllModules() []string {
	return []string{
		ModuleMember, ModuleDevice, ModuleAnalytics,
		ModuleCoupon, ModuleReferral, ModuleSurvey, ModuleFeedback, ModuleLLMRelay,
	}
}

func isKnownModule(name string) bool {
	for _, m := range AllModules() {
		if m == name {
			return true
		}
	}
	return false
}

func normalizeModuleName(name string) string {
	switch strings.ToLower(strings.TrimSpace(name)) {
	case "llm", "llm-api":
		return ModuleLLMRelay
	default:
		return strings.ToLower(strings.TrimSpace(name))
	}
}

func parseModules(raw string) map[string]struct{} {
	set := make(map[string]struct{})
	raw = strings.TrimSpace(raw)
	if raw == "" {
		for _, m := range AllModules() {
			set[m] = struct{}{}
		}
		return set
	}
	// 支持 JSON 数组或逗号分隔
	if strings.HasPrefix(raw, "[") {
		var list []string
		if err := json.Unmarshal([]byte(raw), &list); err == nil {
			for _, part := range list {
				part = normalizeModuleName(part)
				if isKnownModule(part) {
					set[part] = struct{}{}
				}
			}
			return set
		}
	}
	for _, part := range strings.Split(raw, ",") {
		part = normalizeModuleName(part)
		if isKnownModule(part) {
			set[part] = struct{}{}
		}
	}
	return set
}

// SetModulesCSV 用逗号分隔字符串更新已启用模块
func SetModulesCSV(raw string) {
	SetModules(parseModules(raw))
}

// SetModulesList 用模块名列表更新已启用模块（未知名忽略）
func SetModulesList(names []string) {
	set := make(map[string]struct{})
	for _, part := range names {
		part = normalizeModuleName(part)
		if isKnownModule(part) {
			set[part] = struct{}{}
		}
	}
	SetModules(set)
}

// SetModules 替换内存中的模块集合
func SetModules(set map[string]struct{}) {
	if set == nil {
		set = make(map[string]struct{})
	}
	moduleMu.Lock()
	moduleSet = set
	moduleMu.Unlock()
	if globalCfg != nil {
		globalCfg.Modules = strings.Join(EnabledModules(), ",")
	}
}

// ModulesJSON 当前启用模块的 JSON 数组字符串（写入 DB）
func ModulesJSON() string {
	b, _ := json.Marshal(EnabledModules())
	return string(b)
}

// HasModule reports whether a module is enabled.
func HasModule(name string) bool {
	moduleMu.RLock()
	defer moduleMu.RUnlock()
	if moduleSet == nil {
		return true
	}
	_, ok := moduleSet[normalizeModuleName(name)]
	return ok
}

// EnabledModules returns sorted module names for admin UI.
func EnabledModules() []string {
	moduleMu.RLock()
	defer moduleMu.RUnlock()
	order := AllModules()
	if moduleSet == nil {
		out := make([]string, len(order))
		copy(out, order)
		return out
	}
	out := make([]string, 0, len(order))
	for _, m := range order {
		if _, ok := moduleSet[m]; ok {
			out = append(out, m)
		}
	}
	return out
}

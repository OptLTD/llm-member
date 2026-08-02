package service

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/OptLTD/member/internal/config"

	"github.com/OptLTD/member/internal/model"
	"github.com/google/uuid"

	"gorm.io/gorm"
)

const defaultSiteName = "default"

type AnalyticsService struct {
	db *gorm.DB
}

func NewAnalyticsService() *AnalyticsService {
	return &AnalyticsService{db: config.GetDB()}
}

// EnsureDefaultSite 单实例会员系统：启动时确保存在默认统计站点。
func EnsureDefaultSite() error {
	if !config.HasModule(config.ModuleAnalytics) {
		return nil
	}
	_, err := NewAnalyticsService().GetDefaultSite(0)
	return err
}

func (s *AnalyticsService) hashIP(ip string) string {
	if ip == "" {
		return ""
	}
	sum := sha256.Sum256([]byte(ip + config.Load().AppHost))
	return hex.EncodeToString(sum[:8])
}

func (s *AnalyticsService) GetDefaultSite(projectID uint64) (*model.SiteModel, error) {
	var site model.SiteModel
	err := s.db.Where("project_id = ? AND name = ? AND status = ?", projectID, defaultSiteName, "active").First(&site).Error
	if err == nil {
		return &site, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	// 兼容旧数据：使用第一个站点
	if err := s.db.Where("project_id = ? AND status = ?", projectID, "active").Order("id asc").First(&site).Error; err == nil {
		return &site, nil
	}
	site = model.SiteModel{
		ProjectID: projectID,
		SiteKey:   strings.ReplaceAll(uuid.New().String(), "-", "")[:16],
		Name:      defaultSiteName,
		Status:    "active",
	}
	if err := s.db.Create(&site).Error; err != nil {
		return nil, err
	}
	return &site, nil
}

func (s *AnalyticsService) resolveSite(projectID uint64, siteKey string) (*model.SiteModel, error) {
	key := strings.TrimSpace(siteKey)
	if key == "" {
		return s.GetDefaultSite(projectID)
	}
	var site model.SiteModel
	err := s.db.Where("project_id = ? AND site_key = ? AND status = ?", projectID, key, "active").First(&site).Error
	return &site, err
}

func (s *AnalyticsService) GetSiteByKey(projectID uint64, siteKey string) (*model.SiteModel, error) {
	return s.resolveSite(projectID, siteKey)
}

func (s *AnalyticsService) CreateSite(site *model.SiteModel) error {
	if site.SiteKey == "" {
		site.SiteKey = strings.ReplaceAll(uuid.New().String(), "-", "")[:16]
	}
	site.Status = "active"
	return s.db.Create(site).Error
}

func (s *AnalyticsService) UpdateSite(id uint64, site *model.SiteModel) error {
	return s.db.Model(&model.SiteModel{}).Where("id = ?", id).Updates(map[string]any{
		"name":    site.Name,
		"domains": site.Domains,
		"status":  site.Status,
	}).Error
}

func (s *AnalyticsService) ListSites(projectID uint64) ([]model.SiteModel, error) {
	var sites []model.SiteModel
	q := s.db.Order("id desc")
	if projectID > 0 {
		q = q.Where("project_id = ?", projectID)
	}
	err := q.Find(&sites).Error
	return sites, err
}

func (s *AnalyticsService) Track(projectID uint64, req *model.TrackRequest, userID uint64, clientIP string) error {
	site, err := s.resolveSite(projectID, req.SiteKey)
	if err != nil {
		return errors.New("invalid site")
	}
	props := model.EventProps(req.Props)
	event := &model.EventModel{
		ProjectID:   projectID,
		SiteID:      site.ID,
		UserID:      userID,
		EventType:   req.EventType,
		PageURL:     req.PageURL,
		Referrer:    req.Referrer,
		UtmSource:   req.UtmSource,
		UtmMedium:   req.UtmMedium,
		UtmCampaign: req.UtmCampaign,
		VisitorID:   req.VisitorID,
		SessionID:   req.SessionID,
		DeviceID:    req.DeviceID,
		Props:       props,
		IPHash:      s.hashIP(clientIP),
		CreatedAt:   time.Now(),
	}
	if err := s.db.Create(event).Error; err != nil {
		return err
	}
	s.upsertDaily(projectID, site.ID, req)
	return nil
}

func (s *AnalyticsService) upsertDaily(projectID, siteID uint64, req *model.TrackRequest) {
	day := time.Now().Format("2006-01-02")
	var daily model.EventDailyModel
	err := s.db.Where("project_id = ? AND site_id = ? AND day = ?", projectID, siteID, day).First(&daily).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		daily = model.EventDailyModel{ProjectID: projectID, SiteID: siteID, Day: day, PV: 0, UV: 0, Events: 0}
	}
	daily.Events++
	if req.EventType == "app.page_view" || req.EventType == "page_view" {
		daily.PV++
	}
	if req.EventType == "app.launch" || req.EventType == "session.start" {
		daily.UV++
	}
	if daily.ID == 0 {
		_ = s.db.Create(&daily).Error
	} else {
		_ = s.db.Save(&daily).Error
	}
}

func (s *AnalyticsService) TrackInternal(projectID, userID uint64, eventType string, props map[string]any) error {
	if !config.HasModule(config.ModuleAnalytics) {
		return nil
	}
	site, err := s.GetDefaultSite(projectID)
	if err != nil {
		return err
	}
	req := &model.TrackRequest{
		SiteKey:   site.SiteKey,
		EventType: eventType,
		Props:     props,
	}
	return s.Track(projectID, req, userID, "")
}

func (s *AnalyticsService) ListEvents(projectID, siteID uint64, page, size int) ([]model.EventModel, int64, error) {
	if page < 1 {
		page = 1
	}
	if size < 1 || size > 100 {
		size = 20
	}
	var total int64
	var events []model.EventModel
	q := s.db.Model(&model.EventModel{})
	if projectID > 0 {
		q = q.Where("project_id = ?", projectID)
	}
	if siteID > 0 {
		q = q.Where("site_id = ?", siteID)
	}
	q.Count(&total)
	err := q.Order("created_at desc").Offset((page - 1) * size).Limit(size).Find(&events).Error
	return events, total, err
}

func (s *AnalyticsService) Overview(projectID, siteID uint64) (map[string]any, error) {
	if siteID == 0 {
		if site, err := s.GetDefaultSite(projectID); err == nil {
			siteID = site.ID
		}
	}
	day := time.Now().Format("2006-01-02")
	var daily model.EventDailyModel
	q := s.db.Where("project_id = ? AND day = ?", projectID, day)
	if siteID > 0 {
		q = q.Where("site_id = ?", siteID)
	}
	_ = q.First(&daily).Error

	var totalEvents int64
	evQ := s.db.Model(&model.EventModel{}).Where("project_id = ? AND created_at >= ?", projectID, time.Now().AddDate(0, 0, -7))
	if siteID > 0 {
		evQ = evQ.Where("site_id = ?", siteID)
	}
	evQ.Count(&totalEvents)

	return map[string]any{
		"todayPV":     daily.PV,
		"todayUV":     daily.UV,
		"todayEvents": daily.Events,
		"weekEvents":  totalEvents,
		"embedCode":   s.EmbedCode(),
	}, nil
}

func (s *AnalyticsService) EmbedCode() string {
	return `<script src="/sdk/analytics.js"></script>`
}

func (s *AnalyticsService) EmbedCodeLegacy(siteKey string) string {
	if siteKey == "" {
		return s.EmbedCode()
	}
	return fmt.Sprintf(`<script src="/sdk/analytics.js" data-site-key="%s"></script>`, siteKey)
}

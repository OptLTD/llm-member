package service

import (
	"errors"
	"fmt"
	"time"

	"github.com/OptLTD/member/internal/config"
	"github.com/OptLTD/member/internal/model"

	"gorm.io/gorm"
)

type DeviceService struct {
	db    *gorm.DB
	setup *SetupService
}

func NewDeviceService() *DeviceService {
	return &DeviceService{
		db:    config.GetDB(),
		setup: NewSetupService(),
	}
}

func (s *DeviceService) maxDevicesForUser(user *model.UserModel) (int, error) {
	if user == nil {
		return 0, errors.New("user required")
	}
	plan := &model.PlanInfo{Plan: string(user.UserPlan)}
	if err := s.setup.GetAsTarget(user.ProjectID, "plan."+plan.Plan, plan); err != nil {
		return 1, nil
	}
	if plan.MaxDevices > 0 {
		return plan.MaxDevices, nil
	}
	switch user.UserPlan {
	case model.PlanBasic:
		return 1, nil
	case model.PlanExtra:
		return 2, nil
	case model.PlanUltra:
		return 3, nil
	case model.PlanSuper:
		return 0, nil // unlimited
	default:
		return 1, nil
	}
}

func (s *DeviceService) Bind(userID uint64, req *model.BindDeviceRequest) (*model.DeviceModel, error) {
	user, err := NewUserService().GetUserByID(userID)
	if err != nil {
		return nil, err
	}
	maxDev, err := s.maxDevicesForUser(user)
	if err != nil {
		return nil, err
	}

	var existing model.DeviceModel
	err = s.db.Where("user_id = ? AND device_id = ?", userID, req.DeviceID).First(&existing).Error
	if err == nil {
		existing.DeviceName = req.DeviceName
		existing.Platform = req.Platform
		existing.AppVersion = req.AppVersion
		existing.LastSeenAt = time.Now()
		if err := s.db.Save(&existing).Error; err != nil {
			return nil, err
		}
		return &existing, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	if maxDev > 0 {
		var count int64
		s.db.Model(&model.DeviceModel{}).Where("user_id = ?", userID).Count(&count)
		if int(count) >= maxDev {
			return nil, fmt.Errorf("设备绑定已达上限（%d 台），请解绑后重试", maxDev)
		}
	}

	device := &model.DeviceModel{
		ProjectID:  user.ProjectID,
		UserID:     userID,
		DeviceID:   req.DeviceID,
		DeviceName: req.DeviceName,
		Platform:   req.Platform,
		AppVersion: req.AppVersion,
		LastSeenAt: time.Now(),
	}
	if err := s.db.Create(device).Error; err != nil {
		return nil, err
	}
	if config.HasModule(config.ModuleAnalytics) {
		_ = NewAnalyticsService().TrackInternal(user.ProjectID, userID, "device.bind", map[string]any{
			"deviceId": req.DeviceID,
		})
	}
	return device, nil
}

func (s *DeviceService) ListByUser(userID uint64) ([]model.DeviceModel, error) {
	var devices []model.DeviceModel
	err := s.db.Where("user_id = ?", userID).Order("last_seen_at desc").Find(&devices).Error
	return devices, err
}

func (s *DeviceService) Unbind(userID, deviceRowID uint64, isAdmin bool) error {
	q := s.db.Where("id = ?", deviceRowID)
	if !isAdmin {
		q = q.Where("user_id = ?", userID)
	}
	res := q.Delete(&model.DeviceModel{})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (s *DeviceService) ListByAdmin(projectID, userID uint64) ([]model.DeviceModel, error) {
	var devices []model.DeviceModel
	q := s.db.Order("last_seen_at desc")
	if projectID > 0 {
		q = q.Where("project_id = ?", projectID)
	}
	if userID > 0 {
		q = q.Where("user_id = ?", userID)
	}
	err := q.Find(&devices).Error
	return devices, err
}

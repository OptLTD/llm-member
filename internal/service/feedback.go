package service

import (
	"github.com/OptLTD/member/internal/config"
	"github.com/OptLTD/member/internal/model"

	"gorm.io/gorm"
)

type FeedbackService struct {
	db *gorm.DB
}

func NewFeedbackService() *FeedbackService {
	return &FeedbackService{db: config.GetDB()}
}

func (s *FeedbackService) Submit(userID uint64, req *model.SubmitFeedbackRequest) (*model.FeedbackModel, error) {
	fb := &model.FeedbackModel{
		UserID:     userID,
		Type:       req.Type,
		Title:      req.Title,
		Content:    req.Content,
		Rating:     req.Rating,
		Contact:    req.Contact,
		AppVersion: req.AppVersion,
		Status:     model.FeedbackStatusOpen,
	}
	if err := s.db.Create(fb).Error; err != nil {
		return nil, err
	}
	if config.HasModule(config.ModuleAnalytics) {
		_ = NewAnalyticsService().TrackInternal(userID, "feedback.submit", map[string]any{
			"type": req.Type,
		})
	}
	return fb, nil
}

func (s *FeedbackService) List(status string, page, size int) ([]model.FeedbackModel, int64, error) {
	if page < 1 {
		page = 1
	}
	if size < 1 {
		size = 20
	}
	var total int64
	var list []model.FeedbackModel
	q := s.db.Model(&model.FeedbackModel{})
	if status != "" {
		q = q.Where("status = ?", status)
	}
	q.Count(&total)
	err := q.Order("id desc").Offset((page - 1) * size).Limit(size).Find(&list).Error
	return list, total, err
}

func (s *FeedbackService) Update(id uint64, req *model.UpdateFeedbackRequest) (*model.FeedbackModel, error) {
	var fb model.FeedbackModel
	if err := s.db.First(&fb, id).Error; err != nil {
		return nil, err
	}
	if req.Status != nil {
		fb.Status = *req.Status
	}
	if req.AdminReply != nil {
		fb.AdminReply = *req.AdminReply
	}
	if err := s.db.Save(&fb).Error; err != nil {
		return nil, err
	}
	return &fb, nil
}

func (s *FeedbackService) Summary() (map[string]any, error) {
	var open, closed int64
	var avgRating float64
	s.db.Model(&model.FeedbackModel{}).Where("status = ?", model.FeedbackStatusOpen).Count(&open)
	s.db.Model(&model.FeedbackModel{}).Where("status = ?", model.FeedbackStatusClosed).Count(&closed)
	s.db.Model(&model.FeedbackModel{}).Where("rating > 0").Select("COALESCE(AVG(rating),0)").Scan(&avgRating)
	return map[string]any{
		"open":      open,
		"closed":    closed,
		"avgRating": avgRating,
	}, nil
}

package service

import (
	"errors"

	"github.com/OptLTD/member/internal/config"
	"github.com/OptLTD/member/internal/model"

	"gorm.io/gorm"
)

type SurveyService struct {
	db *gorm.DB
}

func NewSurveyService() *SurveyService {
	return &SurveyService{db: config.GetDB()}
}

func (s *SurveyService) Create(survey *model.SurveyModel) error {
	return s.db.Create(survey).Error
}

func (s *SurveyService) Update(id uint64, survey *model.SurveyModel) error {
	return s.db.Model(&model.SurveyModel{}).Where("id = ?", id).Updates(map[string]any{
		"title":       survey.Title,
		"description": survey.Description,
		"status":      survey.Status,
		"target_plan": survey.TargetPlan,
		"start_at":    survey.StartAt,
		"end_at":      survey.EndAt,
	}).Error
}

func (s *SurveyService) List() ([]model.SurveyModel, error) {
	var list []model.SurveyModel
	err := s.db.Order("id desc").Find(&list).Error
	return list, err
}

func (s *SurveyService) Get(id uint64) (*model.SurveyModel, []model.SurveyQuestionModel, error) {
	var survey model.SurveyModel
	if err := s.db.First(&survey, id).Error; err != nil {
		return nil, nil, err
	}
	var questions []model.SurveyQuestionModel
	err := s.db.Where("survey_id = ?", id).Order("sort asc, id asc").Find(&questions).Error
	return &survey, questions, err
}

func (s *SurveyService) GetPublished(id uint64) (*model.SurveyModel, []model.SurveyQuestionModel, error) {
	survey, questions, err := s.Get(id)
	if err != nil {
		return nil, nil, err
	}
	if survey.Status != model.SurveyStatusPublished {
		return nil, nil, errors.New("survey not published")
	}
	return survey, questions, nil
}

func (s *SurveyService) SaveQuestion(q *model.SurveyQuestionModel) error {
	if q.ID == 0 {
		return s.db.Create(q).Error
	}
	return s.db.Save(q).Error
}

func (s *SurveyService) DeleteQuestion(id uint64) error {
	return s.db.Delete(&model.SurveyQuestionModel{}, id).Error
}

func (s *SurveyService) Submit(surveyID, userID uint64, answers map[string]any) error {
	resp := &model.SurveyResponseModel{
		SurveyID: surveyID,
		UserID:   userID,
		Answers:  model.AnswerMap(answers),
	}
	if err := s.db.Create(resp).Error; err != nil {
		return err
	}
	if config.HasModule(config.ModuleAnalytics) {
		_ = NewAnalyticsService().TrackInternal(userID, "survey.submit", map[string]any{
			"surveyId": surveyID,
		})
	}
	return nil
}

func (s *SurveyService) ListResponses(surveyID uint64) ([]model.SurveyResponseModel, error) {
	var list []model.SurveyResponseModel
	err := s.db.Where("survey_id = ?", surveyID).Order("id desc").Find(&list).Error
	return list, err
}

func (s *SurveyService) Stats(surveyID uint64) (map[string]any, error) {
	var count int64
	s.db.Model(&model.SurveyResponseModel{}).Where("survey_id = ?", surveyID).Count(&count)
	return map[string]any{"responses": count}, nil
}

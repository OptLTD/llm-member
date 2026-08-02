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

func (s *SurveyService) Create(projectID uint64, survey *model.SurveyModel) error {
	survey.ProjectID = projectID
	return s.db.Create(survey).Error
}

func (s *SurveyService) Update(projectID, id uint64, survey *model.SurveyModel) error {
	return s.db.Model(&model.SurveyModel{}).Where("project_id = ? AND id = ?", projectID, id).Updates(map[string]any{
		"title":       survey.Title,
		"description": survey.Description,
		"status":      survey.Status,
		"target_plan": survey.TargetPlan,
		"start_at":    survey.StartAt,
		"end_at":      survey.EndAt,
	}).Error
}

func (s *SurveyService) List(projectID uint64) ([]model.SurveyModel, error) {
	var list []model.SurveyModel
	q := s.db.Order("id desc")
	if projectID > 0 {
		q = q.Where("project_id = ?", projectID)
	}
	err := q.Find(&list).Error
	return list, err
}

func (s *SurveyService) Get(projectID, id uint64) (*model.SurveyModel, []model.SurveyQuestionModel, error) {
	var survey model.SurveyModel
	if err := s.db.Where("project_id = ? AND id = ?", projectID, id).First(&survey).Error; err != nil {
		return nil, nil, err
	}
	var questions []model.SurveyQuestionModel
	err := s.db.Where("project_id = ? AND survey_id = ?", projectID, id).Order("sort asc, id asc").Find(&questions).Error
	return &survey, questions, err
}

func (s *SurveyService) GetPublished(projectID, id uint64) (*model.SurveyModel, []model.SurveyQuestionModel, error) {
	survey, questions, err := s.Get(projectID, id)
	if err != nil {
		return nil, nil, err
	}
	if survey.Status != model.SurveyStatusPublished {
		return nil, nil, errors.New("survey not published")
	}
	return survey, questions, nil
}

func (s *SurveyService) SaveQuestion(projectID uint64, q *model.SurveyQuestionModel) error {
	if q.ID == 0 {
		q.ProjectID = projectID
		return s.db.Create(q).Error
	}
	return s.db.Save(q).Error
}

func (s *SurveyService) DeleteQuestion(projectID, id uint64) error {
	return s.db.Where("project_id = ? AND id = ?", projectID, id).Delete(&model.SurveyQuestionModel{}).Error
}

func (s *SurveyService) Submit(projectID, surveyID, userID uint64, answers map[string]any) error {
	resp := &model.SurveyResponseModel{
		ProjectID: projectID,
		SurveyID:  surveyID,
		UserID:    userID,
		Answers:   model.AnswerMap(answers),
	}
	if err := s.db.Create(resp).Error; err != nil {
		return err
	}
	if config.HasModule(config.ModuleAnalytics) {
		_ = NewAnalyticsService().TrackInternal(projectID, userID, "survey.submit", map[string]any{
			"surveyId": surveyID,
		})
	}
	return nil
}

func (s *SurveyService) ListResponses(projectID, surveyID uint64) ([]model.SurveyResponseModel, error) {
	var list []model.SurveyResponseModel
	err := s.db.Where("project_id = ? AND survey_id = ?", projectID, surveyID).Order("id desc").Find(&list).Error
	return list, err
}

func (s *SurveyService) Stats(projectID, surveyID uint64) (map[string]any, error) {
	var count int64
	s.db.Model(&model.SurveyResponseModel{}).Where("project_id = ? AND survey_id = ?", projectID, surveyID).Count(&count)
	return map[string]any{"responses": count}, nil
}

package model

import (
	"database/sql/driver"
	"encoding/json"
	"time"

	"gorm.io/gorm"
)

const (
	SurveyStatusDraft     = "draft"
	SurveyStatusPublished = "published"
	SurveyStatusClosed    = "closed"

	QuestionTypeSingle  = "single"
	QuestionTypeMulti   = "multi"
	QuestionTypeText    = "text"
	QuestionTypeRating  = "rating"
)

type SurveyOptions []string

func (o SurveyOptions) Value() (driver.Value, error) {
	if o == nil {
		return "[]", nil
	}
	return json.Marshal(o)
}

func (o *SurveyOptions) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	var data []byte
	switch v := value.(type) {
	case []byte:
		data = v
	case string:
		data = []byte(v)
	default:
		return nil
	}
	return json.Unmarshal(data, o)
}

type AnswerMap map[string]any

func (a AnswerMap) Value() (driver.Value, error) {
	if a == nil {
		return "{}", nil
	}
	return json.Marshal(a)
}

func (a *AnswerMap) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	var data []byte
	switch v := value.(type) {
	case []byte:
		data = v
	case string:
		data = []byte(v)
	default:
		return nil
	}
	return json.Unmarshal(data, a)
}

// SurveyModel 问卷
type SurveyModel struct {
	ID uint64 `json:"id" gorm:"primarykey"`

	Title       string `json:"title" gorm:"column:title;type:varchar(256);not null"`
	Description string `json:"description" gorm:"column:description;type:text"`
	Status     string     `json:"status" gorm:"column:status;type:varchar(20);default:draft"`
	TargetPlan string     `json:"targetPlan" gorm:"column:target_plan;type:varchar(64)"`
	StartAt    *time.Time `json:"startAt" gorm:"column:start_at"`
	EndAt      *time.Time `json:"endAt" gorm:"column:end_at"`

	gorm.Model
}

func (SurveyModel) TableName() string { return "member_survey" }

// SurveyQuestionModel 题目
type SurveyQuestionModel struct {
	ID uint64 `json:"id" gorm:"primarykey"`

	SurveyID uint64        `json:"surveyId" gorm:"column:survey_id;index;not null"`
	Type     string        `json:"type" gorm:"column:type;type:varchar(20);not null"`
	Title    string        `json:"title" gorm:"column:title;type:varchar(512);not null"`
	Options  SurveyOptions `json:"options" gorm:"column:options;serializer:json"`
	Required bool          `json:"required" gorm:"column:required;default:false"`
	Sort     int           `json:"sort" gorm:"column:sort;default:0"`

	gorm.Model
}

func (SurveyQuestionModel) TableName() string { return "member_survey_question" }

// SurveyResponseModel 答卷
type SurveyResponseModel struct {
	ID uint64 `json:"id" gorm:"primarykey"`

	SurveyID uint64    `json:"surveyId" gorm:"column:survey_id;index;not null"`
	UserID   uint64    `json:"userId" gorm:"column:user_id;index"`
	Answers  AnswerMap `json:"answers" gorm:"column:answers;serializer:json"`

	gorm.Model
}

func (SurveyResponseModel) TableName() string { return "member_survey_response" }

type SubmitSurveyRequest struct {
	Answers map[string]any `json:"answers" binding:"required"`
}

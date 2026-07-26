package handle

import (
	"net/http"
	"strconv"

	"github.com/OptLTD/member/internal/model"
	"github.com/OptLTD/member/internal/service"

	"github.com/gin-gonic/gin"
)

type SurveyHandle struct {
	svc *service.SurveyService
}

func NewSurveyHandle() *SurveyHandle {
	return &SurveyHandle{svc: service.NewSurveyService()}
}

func (h *SurveyHandle) GetPublic(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	survey, questions, err := h.svc.GetPublished(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"survey": survey, "questions": questions})
}

func (h *SurveyHandle) Submit(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var req model.SubmitSurveyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	userID := c.GetUint64("UserID")
	if err := h.svc.Submit(id, userID, req.Answers); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "提交成功"})
}

func (h *SurveyHandle) List(c *gin.Context) {
	list, err := h.svc.List()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"surveys": list})
}

func (h *SurveyHandle) Get(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	survey, questions, err := h.svc.Get(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"survey": survey, "questions": questions})
}

func (h *SurveyHandle) Create(c *gin.Context) {
	var survey model.SurveyModel
	if err := c.ShouldBindJSON(&survey); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.svc.Create(&survey); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"survey": survey})
}

func (h *SurveyHandle) Update(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var survey model.SurveyModel
	if err := c.ShouldBindJSON(&survey); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.svc.Update(id, &survey); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "updated"})
}

func (h *SurveyHandle) SaveQuestion(c *gin.Context) {
	var q model.SurveyQuestionModel
	if err := c.ShouldBindJSON(&q); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.svc.SaveQuestion(&q); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"question": q})
}

func (h *SurveyHandle) DeleteQuestion(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if err := h.svc.DeleteQuestion(id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}

func (h *SurveyHandle) Responses(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	list, err := h.svc.ListResponses(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	stats, _ := h.svc.Stats(id)
	c.JSON(http.StatusOK, gin.H{"responses": list, "stats": stats})
}

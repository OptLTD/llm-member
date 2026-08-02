package handle

import (
	"net/http"
	"strconv"

	"github.com/OptLTD/member/internal/model"
	"github.com/OptLTD/member/internal/service"

	"github.com/gin-gonic/gin"
)

type FeedbackHandle struct {
	svc *service.FeedbackService
}

func NewFeedbackHandle() *FeedbackHandle {
	return &FeedbackHandle{svc: service.NewFeedbackService()}
}

func (h *FeedbackHandle) Submit(c *gin.Context) {
	var req model.SubmitFeedbackRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	fb, err := h.svc.Submit(service.ProjectID(c), c.GetUint64("UserID"), &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"feedback": fb})
}

func (h *FeedbackHandle) List(c *gin.Context) {
	status := c.Query("status")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "20"))
	list, total, err := h.svc.List(service.ProjectID(c), status, page, size)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"feedbacks": list, "total": total})
}

func (h *FeedbackHandle) Update(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var req model.UpdateFeedbackRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	fb, err := h.svc.Update(service.ProjectID(c), id, &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"feedback": fb})
}

func (h *FeedbackHandle) Summary(c *gin.Context) {
	data, err := h.svc.Summary(service.ProjectID(c))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, data)
}

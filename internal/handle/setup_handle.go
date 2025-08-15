package handle

import (
	"encoding/json"
	"net/http"

	"llm-member/internal/service"

	"github.com/gin-gonic/gin"
)

type SetupHandler struct {
	setupService *service.SetupService
}

func NewSetupHandler() *SetupHandler {
	return &SetupHandler{
		setupService: service.NewSetupService(),
	}
}

// GetPricingPlans 获取付费方案列表（管理员）
func (h *SetupHandler) GetPricingPlans(c *gin.Context) {
	plans := h.setupService.GetAllPlans()
	c.JSON(http.StatusOK, gin.H{
		"plans": plans,
	})
}

func (h *SetupHandler) SetPricingPlan(c *gin.Context) {
	planKey := c.Param("plan")

	var req map[string]any
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 将整个方案作为JSON对象存储
	planJSON, err := json.Marshal(req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON format"})
		return
	}

	err = h.setupService.SetConfig("plan."+planKey, string(planJSON), "plan")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update plan config"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Plan config updated successfully"})
}

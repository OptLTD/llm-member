package handle

import (
	"encoding/json"
	"net/http"

	"llm-member/internal/model"
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

	var plan model.PlanInfo
	if err := c.ShouldBindJSON(&plan); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if _, err := h.setupService.ParsePlanLimit(&plan); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 将整个方案作为JSON对象存储
	config := &model.ConfigModel{Kind: "plan", Key: "plan." + planKey}
	if data, err := json.Marshal(plan); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON format"})
		return
	} else {
		config.Data = string(data)
	}

	if err := h.setupService.SetConfig(config); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update plan config"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Plan config updated successfully"})
}

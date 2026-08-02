package handle

import (
	"net/http"
	"strconv"

	"github.com/OptLTD/member/internal/model"
	"github.com/OptLTD/member/internal/service"

	"github.com/gin-gonic/gin"
)

type ReferralHandle struct {
	svc *service.ReferralService
}

func NewReferralHandle() *ReferralHandle {
	return &ReferralHandle{svc: service.NewReferralService()}
}

func (h *ReferralHandle) MyCode(c *gin.Context) {
	rc, err := h.svc.GetOrCreateCode(service.ProjectID(c), c.GetUint64("UserID"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": rc.Code})
}

func (h *ReferralHandle) Stats(c *gin.Context) {
	data, err := h.svc.Stats(service.ProjectID(c), c.GetUint64("UserID"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, data)
}

func (h *ReferralHandle) Resolve(c *gin.Context) {
	code := c.Query("code")
	rc, err := h.svc.ResolveCode(service.ProjectID(c), code)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"valid": false})
		return
	}
	c.JSON(http.StatusOK, gin.H{"valid": true, "code": rc.Code})
}

func (h *ReferralHandle) LogShare(c *gin.Context) {
	var req struct {
		Code    string `json:"code"`
		Channel string `json:"channel"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	_ = h.svc.LogShare(service.ProjectID(c), req.Code, req.Channel, c.ClientIP(), 0)
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func (h *ReferralHandle) ListCampaigns(c *gin.Context) {
	list, err := h.svc.ListCampaigns(service.ProjectID(c))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"campaigns": list})
}

func (h *ReferralHandle) SaveCampaign(c *gin.Context) {
	var camp model.CampaignModel
	if err := c.ShouldBindJSON(&camp); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.svc.SaveCampaign(service.ProjectID(c), &camp); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"campaign": camp})
}

func (h *ReferralHandle) ListReferrals(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "20"))
	list, total, err := h.svc.ListReferrals(service.ProjectID(c), page, size)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"referrals": list, "total": total})
}

package handle

import (
	"net/http"
	"strconv"

	"github.com/OptLTD/member/internal/model"
	"github.com/OptLTD/member/internal/service"

	"github.com/gin-gonic/gin"
)

type AnalyticsHandle struct {
	svc *service.AnalyticsService
}

func NewAnalyticsHandle() *AnalyticsHandle {
	return &AnalyticsHandle{svc: service.NewAnalyticsService()}
}

func (h *AnalyticsHandle) Track(c *gin.Context) {
	var req model.TrackRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	userID := c.GetUint64("UserID")
	if err := h.svc.Track(service.ProjectID(c), &req, userID, c.ClientIP()); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func (h *AnalyticsHandle) ListSites(c *gin.Context) {
	sites, err := h.svc.ListSites(service.ProjectID(c))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	for i := range sites {
		sites[i].EmbedCode = h.svc.EmbedCode()
	}
	c.JSON(http.StatusOK, gin.H{"sites": sites})
}

func (h *AnalyticsHandle) CreateSite(c *gin.Context) {
	var site model.SiteModel
	if err := c.ShouldBindJSON(&site); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	site.ProjectID = service.ProjectID(c)
	if err := h.svc.CreateSite(&site); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	site.EmbedCode = h.svc.EmbedCode()
	c.JSON(http.StatusCreated, gin.H{"site": site})
}

func (h *AnalyticsHandle) UpdateSite(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var site model.SiteModel
	if err := c.ShouldBindJSON(&site); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.svc.UpdateSite(id, &site); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "updated"})
}

func (h *AnalyticsHandle) Overview(c *gin.Context) {
	siteID, _ := strconv.ParseUint(c.DefaultQuery("siteId", "0"), 10, 64)
	data, err := h.svc.Overview(service.ProjectID(c), siteID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, data)
}

func (h *AnalyticsHandle) ListEvents(c *gin.Context) {
	siteID, _ := strconv.ParseUint(c.DefaultQuery("siteId", "0"), 10, 64)
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "20"))
	events, total, err := h.svc.ListEvents(service.ProjectID(c), siteID, page, size)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"events": events, "total": total})
}

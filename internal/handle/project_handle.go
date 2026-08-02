package handle

import (
	"net/http"
	"strconv"

	"github.com/OptLTD/member/internal/model"
	"github.com/OptLTD/member/internal/service"

	"github.com/gin-gonic/gin"
)

type ProjectHandle struct {
	svc *service.ProjectService
}

func NewProjectHandle() *ProjectHandle {
	return &ProjectHandle{svc: service.NewProjectService()}
}

// List 项目列表
func (h *ProjectHandle) List(c *gin.Context) {
	list, err := h.svc.List()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	out := make([]gin.H, 0, len(list))
	for i := range list {
		st := list[i].GetSettings()
		out = append(out, gin.H{
			"id":        list[i].ID,
			"name":      list[i].Name,
			"slug":      list[i].Slug,
			"status":    list[i].Status,
			"appName":   st.AppName,
			"settings":  service.ToProjectSettingsResponse(st),
			"createdAt": list[i].CreatedAt,
		})
	}
	c.JSON(http.StatusOK, gin.H{"projects": out})
}

// Current 当前项目
func (h *ProjectHandle) Current(c *gin.Context) {
	p := service.Project(c)
	if p == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "project not found"})
		return
	}
	st := p.GetSettings()
	c.JSON(http.StatusOK, gin.H{
		"id": p.ID, "name": p.Name, "slug": p.Slug, "status": p.Status,
		"settings": service.ToProjectSettingsResponse(st),
	})
}

// Create 创建项目
func (h *ProjectHandle) Create(c *gin.Context) {
	var req struct {
		Name   string `json:"name" binding:"required"`
		Slug   string `json:"slug" binding:"required"`
		Status string `json:"status"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	p := &model.ProjectModel{Name: req.Name, Slug: req.Slug, Status: req.Status}
	if err := h.svc.Create(p); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	// 为新项目初始化默认 plans
	if svc := service.NewSetupService(); svc != nil {
		_ = svc.InitProjectConfigs(p.ID)
	}
	c.JSON(http.StatusCreated, gin.H{"project": p})
}

// Update 更新项目
func (h *ProjectHandle) Update(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var req struct {
		Name   string `json:"name"`
		Slug   string `json:"slug"`
		Status string `json:"status"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	p := &model.ProjectModel{Name: req.Name, Slug: req.Slug, Status: req.Status}
	if err := h.svc.Update(id, p); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "updated"})
}

// Delete 删除项目
func (h *ProjectHandle) Delete(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if err := h.svc.Delete(id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}

// SaveSettings 保存项目设置
func (h *ProjectHandle) SaveSettings(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var req struct {
		AppName      string            `json:"appName"`
		AuthCallback string            `json:"authCallback"`
		LLM          *model.ProjectLLM `json:"llm"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	settings := model.ProjectSettings{
		AppName:      req.AppName,
		AuthCallback: req.AuthCallback,
		LLM:          req.LLM,
	}
	if err := h.svc.SaveSettings(id, settings); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "settings saved"})
}

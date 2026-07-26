package handle

import (
	"net/http"
	"strconv"

	"github.com/OptLTD/member/internal/model"
	"github.com/OptLTD/member/internal/service"

	"github.com/gin-gonic/gin"
)

type DeviceHandle struct {
	svc *service.DeviceService
}

func NewDeviceHandle() *DeviceHandle {
	return &DeviceHandle{svc: service.NewDeviceService()}
}

func (h *DeviceHandle) Bind(c *gin.Context) {
	var req model.BindDeviceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	device, err := h.svc.Bind(c.GetUint64("UserID"), &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"device": device})
}

func (h *DeviceHandle) List(c *gin.Context) {
	devices, err := h.svc.ListByUser(c.GetUint64("UserID"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"devices": devices})
}

func (h *DeviceHandle) Unbind(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if err := h.svc.Unbind(c.GetUint64("UserID"), id, false); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "解绑成功"})
}

func (h *DeviceHandle) AdminList(c *gin.Context) {
	userID, _ := strconv.ParseUint(c.DefaultQuery("userId", "0"), 10, 64)
	devices, err := h.svc.ListByAdmin(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"devices": devices})
}

func (h *DeviceHandle) AdminUnbind(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if err := h.svc.Unbind(0, id, true); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "解绑成功"})
}

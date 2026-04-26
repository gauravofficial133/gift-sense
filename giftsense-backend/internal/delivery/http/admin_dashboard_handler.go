package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/giftsense/backend/internal/usecase"
)

type AdminDashboardHandler struct {
	service *usecase.DashboardService
}

func NewAdminDashboardHandler(service *usecase.DashboardService) *AdminDashboardHandler {
	return &AdminDashboardHandler{service: service}
}

func (h *AdminDashboardHandler) Overview(c *gin.Context) {
	overview, err := h.service.Overview(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "failed to load dashboard"})
		return
	}
	c.JSON(http.StatusOK, overview)
}

func (h *AdminDashboardHandler) Interactions(c *gin.Context) {
	limitStr := c.DefaultQuery("limit", "50")
	limit, _ := strconv.Atoi(limitStr)
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	interactions, err := h.service.InteractionFeed(c.Request.Context(), limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "failed to load interactions"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"interactions": interactions})
}

func (h *AdminDashboardHandler) Families(c *gin.Context) {
	families, err := h.service.ListFamilies(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "failed to load families"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"families": families})
}

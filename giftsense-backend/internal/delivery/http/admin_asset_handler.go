package handler

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/giftsense/backend/internal/port"
	"github.com/giftsense/backend/internal/usecase"
)

type AdminAssetHandler struct {
	assetLib *usecase.AssetLibrary
	planner  *usecase.AssetPromptPlanner
	imageGen port.ImageGenerator
}

func NewAdminAssetHandler(assetLib *usecase.AssetLibrary, planner *usecase.AssetPromptPlanner, imageGen port.ImageGenerator) *AdminAssetHandler {
	return &AdminAssetHandler{assetLib: assetLib, planner: planner, imageGen: imageGen}
}

func (h *AdminAssetHandler) List(c *gin.Context) {
	tags := c.QueryArray("tags")
	style := c.Query("style")

	if h.assetLib == nil {
		c.JSON(http.StatusOK, gin.H{"assets": []interface{}{}})
		return
	}

	assets := h.assetLib.ListAssets(tags, style)
	if assets == nil {
		assets = []usecase.AssetEntry{}
	}
	c.JSON(http.StatusOK, gin.H{"assets": assets})
}

func (h *AdminAssetHandler) PlanPrompt(c *gin.Context) {
	var req usecase.AssetPlanRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if h.planner == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "prompt planner not available"})
		return
	}

	result, err := h.planner.RefinePrompt(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

func (h *AdminAssetHandler) Generate(c *gin.Context) {
	var req struct {
		Prompt string   `json:"prompt" binding:"required"`
		Style  string   `json:"style"`
		Tags   []string `json:"tags"`
		Width  int      `json:"width"`
		Height int      `json:"height"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if h.imageGen == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "image generation not available"})
		return
	}

	if req.Width == 0 {
		req.Width = 512
	}
	if req.Height == 0 {
		req.Height = 512
	}

	result, err := h.imageGen.Generate(c.Request.Context(), port.ImageRequest{
		Prompt: req.Prompt,
		Width:  req.Width,
		Height: req.Height,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "generation failed"})
		return
	}

	if h.assetLib != nil {
		id := fmt.Sprintf("asset_%d", time.Now().UnixMilli())
		_ = h.assetLib.SaveUpload(id, req.Style, req.Tags, result.PNGBase64)
	}

	c.JSON(http.StatusOK, gin.H{"png_base64": result.PNGBase64})
}

func (h *AdminAssetHandler) Upload(c *gin.Context) {
	var req struct {
		PNGBase64 string   `json:"png_base64" binding:"required"`
		Style     string   `json:"style"`
		Tags      []string `json:"tags"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if h.assetLib == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "asset library not available"})
		return
	}

	id := fmt.Sprintf("upload_%d", time.Now().UnixMilli())
	if err := h.assetLib.SaveUpload(id, req.Style, req.Tags, req.PNGBase64); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"id": id})
}

package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/giftsense/backend/internal/domain"
	"github.com/giftsense/backend/internal/usecase"
)

type CardHandler struct {
	cardGen *usecase.CardGenerator
}

func NewCardHandler(cardGen *usecase.CardGenerator) *CardHandler {
	return &CardHandler{cardGen: cardGen}
}

type cardDownloadRequest struct {
	RecipeID    string             `json:"recipe_id" binding:"required"`
	PaletteName string             `json:"palette_name" binding:"required"`
	Content     domain.CardContent `json:"content" binding:"required"`
	MultiPage   bool               `json:"multi_page"`
	DataFields  map[string]string  `json:"data_fields,omitempty"`
	Photos      map[string]string  `json:"photos,omitempty"`
}

func (h *CardHandler) Download(c *gin.Context) {
	var req cardDownloadRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request", "message": err.Error()})
		return
	}

	card := &domain.CardRender{
		RecipeID:    req.RecipeID,
		PaletteName: req.PaletteName,
		Content:     req.Content,
		DataFields:  req.DataFields,
		Photos:      req.Photos,
	}

	var pdfBase64 string
	var err error
	if req.MultiPage {
		pdfBase64, err = h.cardGen.RenderMultiPagePDF(c.Request.Context(), card)
	} else {
		pdfBase64, err = h.cardGen.RenderPDF(c.Request.Context(), card)
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "render_failed", "message": "Failed to render PDF"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"pdf_base64": pdfBase64})
}

type cardReRenderRequest struct {
	RecipeID    string             `json:"recipe_id" binding:"required"`
	PaletteName string             `json:"palette_name" binding:"required"`
	Content     domain.CardContent `json:"content" binding:"required"`
	DataFields  map[string]string  `json:"data_fields,omitempty"`
	Photos      map[string]string  `json:"photos,omitempty"`
}

func (h *CardHandler) ReRender(c *gin.Context) {
	var req cardReRenderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request", "message": err.Error()})
		return
	}

	card := &domain.CardRender{
		RecipeID:    req.RecipeID,
		PaletteName: req.PaletteName,
		Content:     req.Content,
		DataFields:  req.DataFields,
		Photos:      req.Photos,
	}

	previewPNG, err := h.cardGen.RenderPreview(c.Request.Context(), card)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "render_failed", "message": "Failed to re-render card"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"preview_png":  previewPNG,
		"recipe_id":    req.RecipeID,
		"palette_name": req.PaletteName,
		"content":      req.Content,
	})
}

func (h *CardHandler) ListPalettes(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"palettes": usecase.ListPaletteNames()})
}

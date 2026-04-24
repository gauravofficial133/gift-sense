package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/giftsense/backend/internal/adapter/cardrender"
	"github.com/giftsense/backend/internal/domain"
	"github.com/giftsense/backend/internal/usecase"
)

type AdminTemplateHandler struct {
	manager  *usecase.TemplateManager
	compiler *usecase.HTMLCompiler
	renderer *cardrender.Renderer
}

func NewAdminTemplateHandler(manager *usecase.TemplateManager, compiler *usecase.HTMLCompiler, renderer *cardrender.Renderer) *AdminTemplateHandler {
	return &AdminTemplateHandler{manager: manager, compiler: compiler, renderer: renderer}
}

func (h *AdminTemplateHandler) List(c *gin.Context) {
	templates, err := h.manager.List(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if templates == nil {
		templates = []domain.TemplateDefinition{}
	}
	c.JSON(http.StatusOK, gin.H{"templates": templates})
}

func (h *AdminTemplateHandler) Get(c *gin.Context) {
	id := c.Param("id")
	tpl, err := h.manager.Get(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "template not found"})
		return
	}
	c.JSON(http.StatusOK, tpl)
}

func (h *AdminTemplateHandler) Create(c *gin.Context) {
	var tpl domain.TemplateDefinition
	if err := c.ShouldBindJSON(&tpl); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result, err := h.manager.Create(c.Request.Context(), tpl)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, result)
}

func (h *AdminTemplateHandler) Update(c *gin.Context) {
	id := c.Param("id")
	var tpl domain.TemplateDefinition
	if err := c.ShouldBindJSON(&tpl); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result, err := h.manager.Update(c.Request.Context(), id, tpl)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, result)
}

func (h *AdminTemplateHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	if err := h.manager.Delete(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"deleted": true})
}

func (h *AdminTemplateHandler) Duplicate(c *gin.Context) {
	id := c.Param("id")
	result, err := h.manager.Duplicate(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, result)
}

func (h *AdminTemplateHandler) Preview(c *gin.Context) {
	id := c.Param("id")
	tpl, err := h.manager.Get(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "template not found"})
		return
	}

	palette := domain.CardPalette{
		Background:          "#FFF5E6",
		BackgroundSecondary: "#FFEDD5",
		Primary:             "#D4451A",
		Accent:              "#FFB347",
		Ink:                 "#5C3D2E",
		Muted:               "#78716C",
		Overlay:             "rgba(255,180,71,0.1)",
	}

	content := domain.CardContent{
		Headline:  "Happy Birthday, Alex!",
		Body:      "Wishing you a wonderful day filled with joy and laughter. May all your dreams come true.",
		Closing:   "With all my love",
		Signature: "Yours always,",
		Recipient: "Alex",
	}

	html, err := h.compiler.Compile(usecase.CompileInput{
		Template:      *tpl,
		Palette:       palette,
		Content:       content,
		Illustrations: make(map[string]string),
		DataFields:    map[string]string{"message_count": "1247", "top_emoji": "❤️"},
		Photos:        make(map[string]string),
		FontChoices:   make(map[string]string),
		Seed:          time.Now().UnixNano(),
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "compile failed: " + err.Error()})
		return
	}

	if h.renderer == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "renderer not available"})
		return
	}

	png, err := h.renderer.RenderPNG(html, tpl.Canvas.Width, tpl.Canvas.Height)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "render failed"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"preview_png": png})
}

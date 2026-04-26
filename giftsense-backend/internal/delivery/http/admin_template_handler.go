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
	assetLib *usecase.AssetLibrary
}

func NewAdminTemplateHandler(manager *usecase.TemplateManager, compiler *usecase.HTMLCompiler, renderer *cardrender.Renderer, assetLib *usecase.AssetLibrary) *AdminTemplateHandler {
	return &AdminTemplateHandler{manager: manager, compiler: compiler, renderer: renderer, assetLib: assetLib}
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

	var tpl domain.TemplateDefinition
	if err := c.ShouldBindJSON(&tpl); err == nil && len(tpl.Elements) > 0 {
		if tpl.ID == "" {
			tpl.ID = id
		}
	} else {
		saved, err := h.manager.Get(c.Request.Context(), id)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "template not found"})
			return
		}
		tpl = *saved
	}

	seed := time.Now().UnixNano()
	palette := usecase.PickPreviewPaletteVaried(tpl, seed)
	content := usecase.BuildPreviewContentVaried(tpl, seed)
	fontChoices := usecase.PickPreviewFontsVaried(tpl, seed)

	illustrations := make(map[string]string)
	if h.assetLib != nil {
		for _, el := range tpl.Elements {
			if el.IllustrationSlot != nil {
				slotName := el.IllustrationSlot.SlotName
				if el.Decorative != nil && el.Decorative.AssetID != "" {
					if b64 := h.assetLib.FindByID(el.Decorative.AssetID); b64 != nil {
						illustrations[slotName] = *b64
					}
				}
				if _, ok := illustrations[slotName]; !ok {
					all := h.assetLib.ListAssets(nil, el.IllustrationSlot.StyleHint)
					if len(all) > 0 {
						pick := all[seed%int64(len(all))]
						if b64 := h.assetLib.FindByID(pick.ID); b64 != nil {
							illustrations[slotName] = *b64
						}
					}
				}
			}
		}
	}

	html, err := h.compiler.Compile(usecase.CompileInput{
		Template:      tpl,
		Palette:       palette,
		Content:       content,
		Illustrations: illustrations,
		DataFields:    map[string]string{"message_count": "1,247", "top_emoji": "❤️", "top_artist": "Taylor Swift", "hours_listened": "342"},
		Photos:        make(map[string]string),
		FontChoices:   fontChoices,
		Seed:          seed,
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

	h.manager.SavePreview(c.Request.Context(), id, png)

	c.JSON(http.StatusOK, gin.H{"preview_png": png})
}

func (h *AdminTemplateHandler) GetThumbnail(c *gin.Context) {
	id := c.Param("id")
	png, err := h.manager.GetPreview(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "no preview available"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"preview_png": png})
}

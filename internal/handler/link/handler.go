package link

import (
	"context"
	"encoding/json"
	stderrors "errors"
	"net/http"
	"strconv"

	apperrors "project-3/internal/errors"
	"project-3/internal/model"

	"github.com/gin-gonic/gin"
)

type LinkService interface {
	GetByID(ctx context.Context, id int64) (model.Link, error)
	List(ctx context.Context, from int, to int) ([]model.Link, error)
	Create(ctx context.Context, link model.Link) (model.Link, error)
	Update(ctx context.Context, link model.Link) (model.Link, error)
	Delete(ctx context.Context, id int64) error
}

type Handler struct {
	service LinkService
}

type linkRequest struct {
	OriginalUrl string `json:"original_url"`
	ShortName   string `json:"short_name"`
}

func New(service LinkService) *Handler {
	return &Handler{
		service: service,
	}
}

func RegisterRoutes(router gin.IRoutes, handler *Handler) {
	router.GET("/api/links", handler.List)
	router.POST("/api/links", handler.Create)
	router.GET("/api/links/:id", handler.GetByID)
	router.PUT("/api/links/:id", handler.Update)
	router.DELETE("/api/links/:id", handler.Delete)
}

func (h *Handler) GetByID(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}

	link, err := h.service.GetByID(c.Request.Context(), id)
	if err != nil {
		writeError(c, err)
		return
	}

	c.JSON(http.StatusOK, link)
}

func (h *Handler) List(c *gin.Context) {
	from, to, ok := parseRange(c)
	if !ok {
		return
	}

	links, err := h.service.List(c.Request.Context(), from, to)
	if err != nil {
		writeError(c, err)
		return
	}

	c.JSON(http.StatusOK, links)
}

func parseRange(c *gin.Context) (int, int, bool) {
	var values []int
	if err := json.Unmarshal([]byte(c.Query("range")), &values); err != nil {
		writeError(c, apperrors.ErrNotValidQuery)
		return 0, 0, false
	}

	if len(values) != 2 {
		writeError(c, apperrors.ErrNotValidQuery)
		return 0, 0, false
	}

	return values[0], values[1], true
}

func (h *Handler) Create(c *gin.Context) {
	var request linkRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	link, err := h.service.Create(c.Request.Context(), model.Link{
		OriginalUrl: request.OriginalUrl,
		ShortName:   request.ShortName,
	})
	if err != nil {
		writeError(c, err)
		return
	}

	c.JSON(http.StatusCreated, link)
}

func (h *Handler) Update(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}

	var request linkRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	link, err := h.service.Update(c.Request.Context(), model.Link{
		ID:          id,
		OriginalUrl: request.OriginalUrl,
		ShortName:   request.ShortName,
	})
	if err != nil {
		writeError(c, err)
		return
	}

	c.JSON(http.StatusOK, link)
}

func (h *Handler) Delete(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}

	if err := h.service.Delete(c.Request.Context(), id); err != nil {
		writeError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

func parseID(c *gin.Context) (int64, bool) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		writeError(c, apperrors.ErrInvalidLinkID)
		return 0, false
	}

	return id, true
}

func writeError(c *gin.Context, err error) {
	switch {
	case stderrors.Is(err, apperrors.ErrInvalidLinkID):
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	case stderrors.Is(err, apperrors.ErrNotValidQuery):
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	case stderrors.Is(err, apperrors.ErrLinkNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
	case stderrors.Is(err, apperrors.ErrShortNameTaken):
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
}

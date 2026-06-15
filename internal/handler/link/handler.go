package link

import (
	"context"
	"encoding/json"
	stderrors "errors"
	"fmt"
	"net/http"
	"strconv"

	apperrors "project-3/internal/errors"
	"project-3/internal/model"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type LinkService interface {
	GetByID(ctx context.Context, id int64) (model.Link, error)
	GetByShortName(ctx context.Context, shortName string) (model.Link, error)
	List(ctx context.Context, from int, to int) ([]model.Link, int64, error)
	Create(ctx context.Context, link model.Link) (model.Link, error)
	Update(ctx context.Context, link model.Link) (model.Link, error)
	Delete(ctx context.Context, id int64) error
}

type LinkVisitService interface {
	Create(ctx context.Context, visit model.LinkVisit) (model.LinkVisit, error)
}

type Handler struct {
	service      LinkService
	visitService LinkVisitService
}

type createLinkPayload struct {
	OriginalUrl string `json:"original_url" binding:"required,url"`
	ShortName   string `json:"short_name" binding:"omitempty,min=3,max=32"`
}

func New(service LinkService, visitService LinkVisitService) *Handler {
	return &Handler{
		service:      service,
		visitService: visitService,
	}
}

func RegisterRoutes(router gin.IRoutes, handler *Handler) {
	router.GET("/r/:code", handler.Redirect)
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

	links, total, err := h.service.List(c.Request.Context(), from, to)
	if err != nil {
		writeError(c, err)
		return
	}

	c.Header("Access-Control-Expose-Headers", "Content-Range")
	c.Header("Content-Range", fmt.Sprintf("links %d-%d/%d", from, to, total))
	c.JSON(http.StatusOK, links)
}

func (h *Handler) Redirect(c *gin.Context) {
	const status = http.StatusFound

	link, err := h.service.GetByShortName(c.Request.Context(), c.Param("code"))
	if err != nil {
		writeError(c, err)
		return
	}

	if _, err := h.visitService.Create(c.Request.Context(), model.LinkVisit{
		LinkID:    link.ID,
		IP:        c.ClientIP(),
		UserAgent: c.GetHeader("User-Agent"),
		Referer:   c.GetHeader("Referer"),
		Status:    status,
	}); err != nil {
		writeError(c, err)
		return
	}

	c.Redirect(status, link.OriginalUrl)
}

func parseRange(c *gin.Context) (int, int, bool) {
	rawRange := c.Query("range")
	if rawRange == "" {
		rawRange = c.GetHeader("Range")
	}

	var values []int
	if err := json.Unmarshal([]byte(rawRange), &values); err != nil {
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
	request, ok := bindLinkPayload(c)
	if !ok {
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

	request, ok := bindLinkPayload(c)
	if !ok {
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

func bindLinkPayload(c *gin.Context) (createLinkPayload, bool) {
	var request createLinkPayload
	if err := c.ShouldBindJSON(&request); err != nil {
		writeBindError(c, err)
		return createLinkPayload{}, false
	}

	return request, true
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
		writeFieldError(c, "short_name", "short name already in use")
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
}

func writeBindError(c *gin.Context, err error) {
	var validationErrors validator.ValidationErrors
	if stderrors.As(err, &validationErrors) {
		errorsByField := make(map[string]string, len(validationErrors))
		for _, fieldError := range validationErrors {
			field := validationFieldName(fieldError)
			errorsByField[field] = validationErrorMessage(fieldError, field)
		}

		c.JSON(http.StatusUnprocessableEntity, gin.H{"errors": errorsByField})
		return
	}

	c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
}

func writeFieldError(c *gin.Context, field string, message string) {
	c.JSON(http.StatusUnprocessableEntity, gin.H{
		"errors": map[string]string{
			field: message,
		},
	})
}

func validationFieldName(fieldError validator.FieldError) string {
	switch fieldError.StructField() {
	case "OriginalUrl":
		return "original_url"
	case "ShortName":
		return "short_name"
	default:
		return fieldError.Field()
	}
}

func validationErrorMessage(fieldError validator.FieldError, field string) string {
	return fmt.Sprintf(
		"Key: 'createLinkPayload.%s' Error:Field validation for '%s' failed on the '%s' tag",
		field,
		field,
		fieldError.Tag(),
	)
}

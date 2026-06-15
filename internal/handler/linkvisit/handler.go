package linkvisit

import (
	"context"
	"encoding/json"
	stderrors "errors"
	"fmt"
	"net/http"

	apperrors "project-3/internal/errors"
	"project-3/internal/model"

	"github.com/gin-gonic/gin"
)

type Service interface {
	List(ctx context.Context, from int, to int) ([]model.LinkVisit, int64, error)
}

type Handler struct {
	service Service
}

func New(service Service) *Handler {
	return &Handler{service: service}
}

func RegisterRoutes(router gin.IRoutes, handler *Handler) {
	router.GET("/api/link_visits", handler.List)
}

func (h *Handler) List(c *gin.Context) {
	from, to, ok := parseRange(c)
	if !ok {
		return
	}

	visits, total, err := h.service.List(c.Request.Context(), from, to)
	if err != nil {
		writeError(c, err)
		return
	}

	c.Header("Access-Control-Expose-Headers", "Content-Range")
	c.Header("Content-Range", fmt.Sprintf("link_visits %d-%d/%d", from, to, total))
	c.JSON(http.StatusOK, visits)
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

func writeError(c *gin.Context, err error) {
	switch {
	case stderrors.Is(err, apperrors.ErrNotValidQuery):
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
}

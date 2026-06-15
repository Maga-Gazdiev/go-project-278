package app

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestCorsMiddlewareAllowsLocalhostFrontend(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(corsMiddleware())
	router.GET("/ping", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	request := httptest.NewRequest(http.MethodGet, "/ping", nil)
	request.Header.Set("Origin", "http://localhost:5173")

	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, response.Code)
	}
	if got := response.Header().Get("Access-Control-Allow-Origin"); got != "http://localhost:5173" {
		t.Fatalf("expected Access-Control-Allow-Origin %q, got %q", "http://localhost:5173", got)
	}
	if got := response.Header().Get("Access-Control-Expose-Headers"); got != "Content-Range" {
		t.Fatalf("expected Access-Control-Expose-Headers %q, got %q", "Content-Range", got)
	}
}

func TestCorsMiddlewareHandlesPreflight(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(corsMiddleware())
	router.GET("/ping", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	request := httptest.NewRequest(http.MethodOptions, "/ping", nil)
	request.Header.Set("Origin", "http://localhost:5173")
	request.Header.Set("Access-Control-Request-Method", http.MethodPost)

	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)

	if response.Code != http.StatusNoContent {
		t.Fatalf("expected status %d, got %d", http.StatusNoContent, response.Code)
	}
	if got := response.Header().Get("Access-Control-Allow-Origin"); got != "http://localhost:5173" {
		t.Fatalf("expected Access-Control-Allow-Origin %q, got %q", "http://localhost:5173", got)
	}
	if got := response.Header().Get("Access-Control-Allow-Methods"); got != "GET, POST, PUT, PATCH, DELETE, OPTIONS" {
		t.Fatalf("expected Access-Control-Allow-Methods %q, got %q", "GET, POST, PUT, PATCH, DELETE, OPTIONS", got)
	}
	if got := response.Header().Get("Access-Control-Allow-Headers"); got != "Origin, Content-Type, Accept, Authorization, Range, X-Requested-With" {
		t.Fatalf("expected Access-Control-Allow-Headers %q, got %q", "Origin, Content-Type, Accept, Authorization, Range, X-Requested-With", got)
	}
}

package linkvisit

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"project-3/internal/model"

	"github.com/gin-gonic/gin"
)

type fakeService struct {
	from   int
	to     int
	visits []model.LinkVisit
	total  int64
	err    error
}

func (s *fakeService) List(_ context.Context, from int, to int) ([]model.LinkVisit, int64, error) {
	s.from = from
	s.to = to
	return s.visits, s.total, s.err
}

func testRouter(service Service) *gin.Engine {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	RegisterRoutes(router, New(service))

	return router
}

func testRequestWithHeaders(router http.Handler, method, path, body string, headers map[string]string) *httptest.ResponseRecorder {
	request := httptest.NewRequest(method, path, strings.NewReader(body))
	if body != "" {
		request.Header.Set("Content-Type", "application/json")
	}
	for key, value := range headers {
		request.Header.Set(key, value)
	}

	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)

	return response
}

func TestListLinkVisits(t *testing.T) {
	createdAt := time.Date(2025, 10, 31, 13, 1, 43, 0, time.UTC)
	service := &fakeService{
		total: 357,
		visits: []model.LinkVisit{
			{
				ID:        5,
				LinkID:    1,
				CreatedAt: createdAt,
				IP:        "172.18.0.1",
				UserAgent: "curl/8.5.0",
				Referer:   "https://example.com",
				Status:    http.StatusFound,
			},
		},
	}

	response := testRequestWithHeaders(testRouter(service), http.MethodGet, "/api/link_visits", "", map[string]string{
		"Range": "[10,20]",
	})

	if response.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, response.Code)
	}
	if got := response.Header().Get("Content-Range"); got != "link_visits 10-20/357" {
		t.Fatalf("expected Content-Range %q, got %q", "link_visits 10-20/357", got)
	}
	if service.from != 10 || service.to != 20 {
		t.Fatalf("expected range [10, 20], got [%d, %d]", service.from, service.to)
	}

	var visits []model.LinkVisit
	if err := json.NewDecoder(response.Body).Decode(&visits); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(visits) != 1 {
		t.Fatalf("expected 1 visit, got %d", len(visits))
	}
	if visits[0].ID != 5 || visits[0].LinkID != 1 || visits[0].Status != http.StatusFound {
		t.Fatalf("unexpected visit response: %+v", visits[0])
	}
}

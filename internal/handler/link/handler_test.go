package link

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	apperrors "project-3/internal/errors"
	"project-3/internal/model"
	linkservice "project-3/internal/service/link"

	"github.com/gin-gonic/gin"
)

type fakeLinkRepository struct {
	getID       int64
	getLink     model.Link
	getErr      error
	getCode     string
	getCodeLink model.Link
	getCodeErr  error
	listLinks   []model.Link
	listErr     error
	listFrom    int
	listTo      int
	countTotal  int64
	countErr    error
	createInput model.Link
	createLink  model.Link
	createErr   error
	updateInput model.Link
	updateLink  model.Link
	updateErr   error
	deleteID    int64
	deleteErr   error
}

func (r *fakeLinkRepository) GetByID(_ context.Context, id int64) (model.Link, error) {
	r.getID = id
	return r.getLink, r.getErr
}

func (r *fakeLinkRepository) GetByShortName(_ context.Context, shortName string) (model.Link, error) {
	r.getCode = shortName
	return r.getCodeLink, r.getCodeErr
}

func (r *fakeLinkRepository) List(_ context.Context, from int, to int) ([]model.Link, error) {
	r.listFrom = from
	r.listTo = to
	return r.listLinks, r.listErr
}

func (r *fakeLinkRepository) Count(_ context.Context) (int64, error) {
	return r.countTotal, r.countErr
}

func (r *fakeLinkRepository) Create(_ context.Context, link model.Link) (model.Link, error) {
	r.createInput = link
	return r.createLink, r.createErr
}

func (r *fakeLinkRepository) Update(_ context.Context, link model.Link) (model.Link, error) {
	r.updateInput = link
	return r.updateLink, r.updateErr
}

func (r *fakeLinkRepository) Delete(_ context.Context, id int64) error {
	r.deleteID = id
	return r.deleteErr
}

type fakeLinkVisitService struct {
	visitInput model.LinkVisit
	visitErr   error
}

func (s *fakeLinkVisitService) Create(_ context.Context, visit model.LinkVisit) (model.LinkVisit, error) {
	s.visitInput = visit
	return visit, s.visitErr
}

func testRouter(repository linkservice.LinkRepositoryInterface) *gin.Engine {
	return testRouterWithVisitService(repository, &fakeLinkVisitService{})
}

func testRouterWithVisitService(repository linkservice.LinkRepositoryInterface, visitService LinkVisitService) *gin.Engine {
	gin.SetMode(gin.TestMode)

	service := linkservice.NewService(repository, "http://localhost:8888")
	router := gin.New()
	RegisterRoutes(router, New(service, visitService))

	return router
}

func testRequest(router http.Handler, method, path, body string) *httptest.ResponseRecorder {
	request := httptest.NewRequest(method, path, strings.NewReader(body))
	if body != "" {
		request.Header.Set("Content-Type", "application/json")
	}

	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)

	return response
}

func TestListLinks(t *testing.T) {
	repository := &fakeLinkRepository{
		countTotal: 5,
		listLinks: []model.Link{
			{ID: 1, OriginalUrl: "https://example.com/one", ShortName: "one", ShortUrl: "http://localhost:8888/r/one"},
			{ID: 2, OriginalUrl: "https://example.com/two", ShortName: "two", ShortUrl: "http://localhost:8888/r/two"},
		},
	}

	response := testRequest(testRouter(repository), http.MethodGet, "/api/links?range=%5B0,1%5D", "")

	if response.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, response.Code)
	}
	if got := response.Header().Get("Content-Range"); got != "links 0-1/5" {
		t.Fatalf("expected Content-Range %q, got %q", "links 0-1/5", got)
	}
	if got := response.Header().Get("Access-Control-Expose-Headers"); got != "Content-Range" {
		t.Fatalf("expected Access-Control-Expose-Headers %q, got %q", "Content-Range", got)
	}

	var links []model.Link
	if err := json.NewDecoder(response.Body).Decode(&links); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(links) != 2 {
		t.Fatalf("expected 2 links, got %d", len(links))
	}
	if repository.listFrom != 0 || repository.listTo != 1 {
		t.Fatalf("expected range [0, 1], got [%d, %d]", repository.listFrom, repository.listTo)
	}
}

func TestRedirectLink(t *testing.T) {
	repository := &fakeLinkRepository{
		getCodeLink: model.Link{
			ID:          1,
			OriginalUrl: "https://example.com/target",
			ShortName:   "abc",
			ShortUrl:    "http://localhost:8888/r/abc",
		},
	}
	visitService := &fakeLinkVisitService{}

	request := httptest.NewRequest(http.MethodGet, "/r/abc", nil)
	request.RemoteAddr = "172.18.0.1:12345"
	request.Header.Set("User-Agent", "curl/8.5.0")
	request.Header.Set("Referer", "https://example.com/source")
	response := httptest.NewRecorder()

	testRouterWithVisitService(repository, visitService).ServeHTTP(response, request)

	if response.Code != http.StatusFound {
		t.Fatalf("expected status %d, got %d", http.StatusFound, response.Code)
	}
	if got := response.Header().Get("Location"); got != "https://example.com/target" {
		t.Fatalf("expected redirect location %q, got %q", "https://example.com/target", got)
	}
	if repository.getCode != "abc" {
		t.Fatalf("expected short name %q, got %q", "abc", repository.getCode)
	}
	if visitService.visitInput.LinkID != 1 {
		t.Fatalf("expected visit link id 1, got %d", visitService.visitInput.LinkID)
	}
	if visitService.visitInput.IP != "172.18.0.1" {
		t.Fatalf("expected visit ip %q, got %q", "172.18.0.1", visitService.visitInput.IP)
	}
	if visitService.visitInput.UserAgent != "curl/8.5.0" {
		t.Fatalf("expected user agent %q, got %q", "curl/8.5.0", visitService.visitInput.UserAgent)
	}
	if visitService.visitInput.Referer != "https://example.com/source" {
		t.Fatalf("expected referer %q, got %q", "https://example.com/source", visitService.visitInput.Referer)
	}
	if visitService.visitInput.Status != http.StatusFound {
		t.Fatalf("expected visit status %d, got %d", http.StatusFound, visitService.visitInput.Status)
	}
}

func TestRedirectLinkNotFound(t *testing.T) {
	repository := &fakeLinkRepository{getCodeErr: apperrors.ErrLinkNotFound}
	visitService := &fakeLinkVisitService{}

	response := testRequest(testRouterWithVisitService(repository, visitService), http.MethodGet, "/r/missing", "")

	if response.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d", http.StatusNotFound, response.Code)
	}
	if visitService.visitInput != (model.LinkVisit{}) {
		t.Fatalf("expected no visit record, got %+v", visitService.visitInput)
	}
}

func TestCreateLink(t *testing.T) {
	repository := &fakeLinkRepository{
		createLink: model.Link{ID: 1, OriginalUrl: "https://example.com/long-url", ShortName: "exmpl", ShortUrl: "http://localhost:8888/r/exmpl"},
	}

	response := testRequest(testRouter(repository), http.MethodPost, "/api/links", `{"original_url":"https://example.com/long-url","short_name":"exmpl"}`)

	if response.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d", http.StatusCreated, response.Code)
	}
	if repository.createInput.OriginalUrl != "https://example.com/long-url" || repository.createInput.ShortName != "exmpl" {
		t.Fatalf("unexpected create input: %+v", repository.createInput)
	}
	if repository.createInput.ShortUrl != "http://localhost:8888/r/exmpl" {
		t.Fatalf("unexpected short url: %s", repository.createInput.ShortUrl)
	}

	var link model.Link
	if err := json.NewDecoder(response.Body).Decode(&link); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if link.ID != 1 {
		t.Fatalf("expected created link id 1, got %d", link.ID)
	}
}

func TestCreateLinkGeneratesShortName(t *testing.T) {
	repository := &fakeLinkRepository{}

	response := testRequest(testRouter(repository), http.MethodPost, "/api/links", `{"original_url":"https://example.com/long-url"}`)

	if response.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d", http.StatusCreated, response.Code)
	}
	if repository.createInput.OriginalUrl != "https://example.com/long-url" {
		t.Fatalf("unexpected original url: %s", repository.createInput.OriginalUrl)
	}
	if repository.createInput.ShortName == "" {
		t.Fatal("expected generated short name")
	}
	if repository.createInput.ShortUrl != "http://localhost:8888/r/"+repository.createInput.ShortName {
		t.Fatalf("unexpected short url: %s", repository.createInput.ShortUrl)
	}
}

func TestCreateLinkInvalidJSON(t *testing.T) {
	repository := &fakeLinkRepository{}

	response := testRequest(testRouter(repository), http.MethodPost, "/api/links", `{"original_url":`)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, response.Code)
	}

	var body map[string]string
	if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body["error"] != "invalid request" {
		t.Fatalf("expected invalid request error, got %+v", body)
	}
}

func TestCreateLinkInvalidOriginalURL(t *testing.T) {
	repository := &fakeLinkRepository{}

	response := testRequest(testRouter(repository), http.MethodPost, "/api/links", `{"original_url":"not-url","short_name":"valid"}`)

	if response.Code != http.StatusUnprocessableEntity {
		t.Fatalf("expected status %d, got %d", http.StatusUnprocessableEntity, response.Code)
	}

	var body struct {
		Errors map[string]string `json:"errors"`
	}
	if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if !strings.Contains(body.Errors["original_url"], "'url' tag") {
		t.Fatalf("expected original_url validation error, got %+v", body.Errors)
	}
}

func TestCreateLinkInvalidShortName(t *testing.T) {
	repository := &fakeLinkRepository{}

	response := testRequest(testRouter(repository), http.MethodPost, "/api/links", `{"original_url":"https://example.com","short_name":"ab"}`)

	if response.Code != http.StatusUnprocessableEntity {
		t.Fatalf("expected status %d, got %d", http.StatusUnprocessableEntity, response.Code)
	}

	var body struct {
		Errors map[string]string `json:"errors"`
	}
	if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if !strings.Contains(body.Errors["short_name"], "'min' tag") {
		t.Fatalf("expected short_name validation error, got %+v", body.Errors)
	}
}

func TestGetLinkByID(t *testing.T) {
	repository := &fakeLinkRepository{
		getLink: model.Link{ID: 1, OriginalUrl: "https://example.com/long-url", ShortName: "exmpl", ShortUrl: "http://localhost:8888/r/exmpl"},
	}

	response := testRequest(testRouter(repository), http.MethodGet, "/api/links/1", "")

	if response.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, response.Code)
	}
	if repository.getID != 1 {
		t.Fatalf("expected repository id 1, got %d", repository.getID)
	}
}

func TestUpdateLink(t *testing.T) {
	repository := &fakeLinkRepository{
		updateLink: model.Link{ID: 1, OriginalUrl: "https://example.com/new-url", ShortName: "new", ShortUrl: "http://localhost:8888/r/new"},
	}

	response := testRequest(testRouter(repository), http.MethodPut, "/api/links/1", `{"original_url":"https://example.com/new-url","short_name":"new"}`)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, response.Code)
	}
	if repository.updateInput.ID != 1 || repository.updateInput.ShortName != "new" {
		t.Fatalf("unexpected update input: %+v", repository.updateInput)
	}
	if repository.updateInput.ShortUrl != "http://localhost:8888/r/new" {
		t.Fatalf("unexpected short url: %s", repository.updateInput.ShortUrl)
	}
}

func TestUpdateLinkInvalidOriginalURL(t *testing.T) {
	repository := &fakeLinkRepository{}

	response := testRequest(testRouter(repository), http.MethodPut, "/api/links/1", `{"original_url":"not-url","short_name":"valid"}`)

	if response.Code != http.StatusUnprocessableEntity {
		t.Fatalf("expected status %d, got %d", http.StatusUnprocessableEntity, response.Code)
	}

	var body struct {
		Errors map[string]string `json:"errors"`
	}
	if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if !strings.Contains(body.Errors["original_url"], "'url' tag") {
		t.Fatalf("expected original_url validation error, got %+v", body.Errors)
	}
}

func TestDeleteLink(t *testing.T) {
	repository := &fakeLinkRepository{}

	response := testRequest(testRouter(repository), http.MethodDelete, "/api/links/1", "")

	if response.Code != http.StatusNoContent {
		t.Fatalf("expected status %d, got %d", http.StatusNoContent, response.Code)
	}
	if repository.deleteID != 1 {
		t.Fatalf("expected deleted id 1, got %d", repository.deleteID)
	}
	if response.Body.Len() != 0 {
		t.Fatalf("expected empty body, got %q", response.Body.String())
	}
}

func TestGetLinkNotFound(t *testing.T) {
	repository := &fakeLinkRepository{getErr: apperrors.ErrLinkNotFound}

	response := testRequest(testRouter(repository), http.MethodGet, "/api/links/404", "")

	if response.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d", http.StatusNotFound, response.Code)
	}
}

func TestCreateLinkShortNameConflict(t *testing.T) {
	repository := &fakeLinkRepository{createErr: apperrors.ErrShortNameTaken}

	response := testRequest(testRouter(repository), http.MethodPost, "/api/links", `{"original_url":"https://example.com","short_name":"exmpl"}`)

	if response.Code != http.StatusUnprocessableEntity {
		t.Fatalf("expected status %d, got %d", http.StatusUnprocessableEntity, response.Code)
	}

	var body struct {
		Errors map[string]string `json:"errors"`
	}
	if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body.Errors["short_name"] != "short name already in use" {
		t.Fatalf("expected short_name conflict error, got %+v", body.Errors)
	}
}

func TestInvalidLinkID(t *testing.T) {
	repository := &fakeLinkRepository{}

	response := testRequest(testRouter(repository), http.MethodGet, "/api/links/not-number", "")

	if response.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, response.Code)
	}
}

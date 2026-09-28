package pages

import (
	"encoding/json"
	"net/url"
	"testing"

	"github.com/itokun99/blogger-go/gen/schemas"
	"github.com/itokun99/blogger-mcp/internal/tools"
	"github.com/itokun99/blogger-mcp/internal/toolstest"
)

func newHandlers(t *testing.T) (*toolstest.MockAPI, *handlers) {
	t.Helper()

	m := toolstest.NewMockAPI(t)
	return m, &handlers{rt: tools.NewRuntimeWithClient(m.Client())}
}

func decodePageBody(t *testing.T, body []byte) schemas.Page {
	t.Helper()

	var page schemas.Page
	if err := json.Unmarshal(body, &page); err != nil {
		t.Fatalf("unmarshaling page body: %v", err)
	}
	return page
}

func TestListPages(t *testing.T) {
	m, h := newHandlers(t)

	m.Stub("GET", "/v3/blogs/123/pages", 200, `{
		"items": [{"id": "456", "title": "About", "status": "LIVE"}]
	}`)

	_, out, err := h.listPages(t.Context(), nil, listPagesInput{
		BlogID:     "123",
		MaxResults: toolstest.Ptr(int64(5)),
		PageToken:  toolstest.Ptr("cursor-1"),
		Status:     toolstest.Ptr("LIVE"),
		View:       toolstest.Ptr("ADMIN"),
		FetchBody:  toolstest.Ptr(true),
	})
	if err != nil {
		t.Fatalf("listPages unexpected error: %v", err)
	}

	if len(out.Items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(out.Items))
	}
	if out.Items[0].Title != "About" {
		t.Errorf("expected item title 'About', got '%s'", out.Items[0].Title)
	}

	calls := m.Calls()
	if len(calls) != 1 {
		t.Fatalf("expected 1 call, got %d", len(calls))
	}
	if calls[0].Method != "GET" {
		t.Errorf("expected method GET, got %s", calls[0].Method)
	}
	if calls[0].Path != "/v3/blogs/123/pages" {
		t.Errorf("expected path /v3/blogs/123/pages, got %s", calls[0].Path)
	}

	parsed, err := url.ParseQuery(calls[0].RawQuery)
	if err != nil {
		t.Fatalf("failed to parse query: %v", err)
	}

	want := map[string]string{
		"maxResults":  "5",
		"pageToken":   "cursor-1",
		"status":      "LIVE",
		"view":        "ADMIN",
		"fetchBodies": "true",
	}
	for key, value := range want {
		if got := parsed.Get(key); got != value {
			t.Errorf("expected %s=%s, got %s", key, value, got)
		}
	}
}

func TestGetPage(t *testing.T) {
	m, h := newHandlers(t)

	m.Stub("GET", "/v3/blogs/123/pages/456", 200, `{
		"id": "456",
		"title": "About",
		"content": "<p>hello</p>",
		"status": "DRAFT"
	}`)

	_, out, err := h.getPage(t.Context(), nil, getPageInput{BlogID: "123", PageID: "456"})
	if err != nil {
		t.Fatalf("getPage unexpected error: %v", err)
	}

	if out.Id != "456" {
		t.Errorf("expected id '456', got '%s'", out.Id)
	}
	if out.Title != "About" {
		t.Errorf("expected title 'About', got '%s'", out.Title)
	}
	if out.Content != "<p>hello</p>" {
		t.Errorf("expected content '<p>hello</p>', got '%s'", out.Content)
	}

	calls := m.Calls()
	if len(calls) != 1 {
		t.Fatalf("expected 1 call, got %d", len(calls))
	}
	if calls[0].Method != "GET" {
		t.Errorf("expected method GET, got %s", calls[0].Method)
	}
	if calls[0].Path != "/v3/blogs/123/pages/456" {
		t.Errorf("expected path /v3/blogs/123/pages/456, got %s", calls[0].Path)
	}
}

func TestCreatePageWithoutDraftDefaultsToDraft(t *testing.T) {
	m, h := newHandlers(t)

	m.Stub("POST", "/v3/blogs/123/pages", 200, `{
		"id": "789",
		"title": "About",
		"content": "<p>hello</p>",
		"status": "DRAFT"
	}`)

	_, out, err := h.createPage(t.Context(), nil, createPageInput{
		BlogID:  "123",
		Title:   "About",
		Content: "<p>hello</p>",
	})
	if err != nil {
		t.Fatalf("createPage unexpected error: %v", err)
	}

	if out.Id != "789" {
		t.Errorf("expected id '789', got '%s'", out.Id)
	}

	calls := m.Calls()
	if len(calls) != 1 {
		t.Fatalf("expected 1 call, got %d", len(calls))
	}
	if calls[0].Method != "POST" {
		t.Errorf("expected method POST, got %s", calls[0].Method)
	}
	if calls[0].Path != "/v3/blogs/123/pages" {
		t.Errorf("expected path /v3/blogs/123/pages, got %s", calls[0].Path)
	}

	body := decodePageBody(t, calls[0].Body)
	if body.Status != "DRAFT" {
		t.Errorf("expected body status DRAFT, got '%s'", body.Status)
	}
	if body.Title != "About" {
		t.Errorf("expected body title 'About', got '%s'", body.Title)
	}
	if body.Content != "<p>hello</p>" {
		t.Errorf("expected body content '<p>hello</p>', got '%s'", body.Content)
	}
}

func TestCreatePageDraftFalseCreatesLive(t *testing.T) {
	m, h := newHandlers(t)

	m.Stub("POST", "/v3/blogs/123/pages", 200, `{
		"id": "789",
		"title": "About",
		"content": "<p>hello</p>",
		"status": "LIVE"
	}`)

	_, out, err := h.createPage(t.Context(), nil, createPageInput{
		BlogID:  "123",
		Title:   "About",
		Content: "<p>hello</p>",
		Draft:   toolstest.Ptr(false),
	})
	if err != nil {
		t.Fatalf("createPage unexpected error: %v", err)
	}

	if out.Status != "LIVE" {
		t.Errorf("expected returned status LIVE, got '%s'", out.Status)
	}

	calls := m.Calls()
	if len(calls) != 1 {
		t.Fatalf("expected 1 call, got %d", len(calls))
	}

	body := decodePageBody(t, calls[0].Body)
	if body.Status != "LIVE" {
		t.Errorf("expected body status LIVE, got '%s'", body.Status)
	}
}

func TestUpdatePageMergesProvidedFields(t *testing.T) {
	m, h := newHandlers(t)

	m.Stub("GET", "/v3/blogs/123/pages/456", 200, `{
		"id": "456",
		"title": "Old",
		"content": "<p>hello</p>",
		"status": "DRAFT"
	}`)
	m.Stub("PUT", "/v3/blogs/123/pages/456", 200, `{
		"id": "456",
		"title": "New",
		"content": "<p>hello</p>",
		"status": "DRAFT"
	}`)

	_, out, err := h.updatePage(t.Context(), nil, updatePageInput{
		BlogID: "123",
		PageID: "456",
		Title:  toolstest.Ptr("New"),
	})
	if err != nil {
		t.Fatalf("updatePage unexpected error: %v", err)
	}

	if out.Title != "New" {
		t.Errorf("expected returned title 'New', got '%s'", out.Title)
	}

	calls := m.Calls()
	if len(calls) != 2 {
		t.Fatalf("expected 2 calls, got %d", len(calls))
	}
	if calls[0].Method != "GET" || calls[0].Path != "/v3/blogs/123/pages/456" {
		t.Errorf("expected first call GET /v3/blogs/123/pages/456, got %s %s", calls[0].Method, calls[0].Path)
	}
	if calls[1].Method != "PUT" || calls[1].Path != "/v3/blogs/123/pages/456" {
		t.Errorf("expected second call PUT /v3/blogs/123/pages/456, got %s %s", calls[1].Method, calls[1].Path)
	}

	body := decodePageBody(t, calls[1].Body)
	if body.Title != "New" {
		t.Errorf("expected body title 'New', got '%s'", body.Title)
	}
	if body.Content != "<p>hello</p>" {
		t.Errorf("expected body to keep fetched content '<p>hello</p>', got '%s'", body.Content)
	}
	if body.Status != "DRAFT" {
		t.Errorf("expected body to keep fetched status DRAFT, got '%s'", body.Status)
	}
}

func TestPatchPageSendsOnlyProvidedFields(t *testing.T) {
	m, h := newHandlers(t)

	m.Stub("PATCH", "/v3/blogs/123/pages/456", 200, `{
		"id": "456",
		"title": "New title"
	}`)

	_, out, err := h.patchPage(t.Context(), nil, patchPageInput{
		BlogID: "123",
		PageID: "456",
		Title:  toolstest.Ptr("New title"),
	})
	if err != nil {
		t.Fatalf("patchPage unexpected error: %v", err)
	}

	if out.Title != "New title" {
		t.Errorf("expected returned title 'New title', got '%s'", out.Title)
	}

	calls := m.Calls()
	if len(calls) != 1 {
		t.Fatalf("expected 1 call, got %d", len(calls))
	}
	if calls[0].Method != "PATCH" {
		t.Errorf("expected method PATCH, got %s", calls[0].Method)
	}
	if calls[0].Path != "/v3/blogs/123/pages/456" {
		t.Errorf("expected path /v3/blogs/123/pages/456, got %s", calls[0].Path)
	}

	var body map[string]any
	if err := json.Unmarshal(calls[0].Body, &body); err != nil {
		t.Fatalf("unmarshaling request body: %v", err)
	}
	if len(body) != 1 {
		t.Fatalf("expected body with exactly 1 field, got %v", body)
	}
	if body["title"] != "New title" {
		t.Errorf("expected body title 'New title', got %v", body["title"])
	}
}

func TestDeletePage(t *testing.T) {
	m, h := newHandlers(t)

	m.Stub("DELETE", "/v3/blogs/123/pages/456", 200, `{}`)

	_, out, err := h.deletePage(t.Context(), nil, deletePageInput{BlogID: "123", PageID: "456"})
	if err != nil {
		t.Fatalf("deletePage unexpected error: %v", err)
	}

	if !out.Deleted {
		t.Error("expected deleted true")
	}
	if out.BlogID != "123" {
		t.Errorf("expected blog_id '123', got '%s'", out.BlogID)
	}
	if out.PageID != "456" {
		t.Errorf("expected page_id '456', got '%s'", out.PageID)
	}

	calls := m.Calls()
	if len(calls) != 1 {
		t.Fatalf("expected 1 call, got %d", len(calls))
	}
	if calls[0].Method != "DELETE" {
		t.Errorf("expected method DELETE, got %s", calls[0].Method)
	}
	if calls[0].Path != "/v3/blogs/123/pages/456" {
		t.Errorf("expected path /v3/blogs/123/pages/456, got %s", calls[0].Path)
	}
}

func TestPublishPage(t *testing.T) {
	m, h := newHandlers(t)

	m.Stub("POST", "/v3/blogs/123/pages/456/publish", 200, `{
		"id": "456",
		"title": "About",
		"status": "LIVE"
	}`)

	_, out, err := h.publishPage(t.Context(), nil, publishPageInput{BlogID: "123", PageID: "456"})
	if err != nil {
		t.Fatalf("publishPage unexpected error: %v", err)
	}

	if out.Id != "456" {
		t.Errorf("expected id '456', got '%s'", out.Id)
	}
	if out.Status != "LIVE" {
		t.Errorf("expected status LIVE, got '%s'", out.Status)
	}

	calls := m.Calls()
	if len(calls) != 1 {
		t.Fatalf("expected 1 call, got %d", len(calls))
	}
	if calls[0].Method != "POST" {
		t.Errorf("expected method POST, got %s", calls[0].Method)
	}
	if calls[0].Path != "/v3/blogs/123/pages/456/publish" {
		t.Errorf("expected path /v3/blogs/123/pages/456/publish, got %s", calls[0].Path)
	}
}

func TestRevertPage(t *testing.T) {
	m, h := newHandlers(t)

	m.Stub("POST", "/v3/blogs/123/pages/456/revert", 200, `{
		"id": "456",
		"title": "About",
		"status": "DRAFT"
	}`)

	_, out, err := h.revertPage(t.Context(), nil, revertPageInput{BlogID: "123", PageID: "456"})
	if err != nil {
		t.Fatalf("revertPage unexpected error: %v", err)
	}

	if out.Id != "456" {
		t.Errorf("expected id '456', got '%s'", out.Id)
	}
	if out.Status != "DRAFT" {
		t.Errorf("expected status DRAFT, got '%s'", out.Status)
	}

	calls := m.Calls()
	if len(calls) != 1 {
		t.Fatalf("expected 1 call, got %d", len(calls))
	}
	if calls[0].Method != "POST" {
		t.Errorf("expected method POST, got %s", calls[0].Method)
	}
	if calls[0].Path != "/v3/blogs/123/pages/456/revert" {
		t.Errorf("expected path /v3/blogs/123/pages/456/revert, got %s", calls[0].Path)
	}
}

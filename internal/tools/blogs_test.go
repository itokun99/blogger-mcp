package tools

import (
	"net/url"
	"testing"

	"github.com/itokun99/blogger-mcp/internal/toolstest"
)

func TestGetBlog(t *testing.T) {
	m := toolstest.NewMockAPI(t)

	blogJSON := `{
		"id": "123",
		"name": "Test Blog",
		"url": "https://test.blogspot.com"
	}`
	m.Stub("GET", "/v3/blogs/123", 200, blogJSON)

	rt := NewRuntimeWithClient(m.Client())
	h := &blogsHandlers{rt: rt}

	_, out, err := h.getBlog(t.Context(), nil, getBlogInput{
		BlogID:   "123",
		View:     toolstest.Ptr("ADMIN"),
		MaxPosts: toolstest.Ptr(int64(10)),
	})
	if err != nil {
		t.Fatalf("getBlog unexpected error: %v", err)
	}

	if out.Name != "Test Blog" {
		t.Errorf("expected name 'Test Blog', got '%s'", out.Name)
	}

	calls := m.Calls()
	if len(calls) != 1 {
		t.Fatalf("expected 1 call, got %d", len(calls))
	}

	parsed, err := url.ParseQuery(calls[0].RawQuery)
	if err != nil {
		t.Fatalf("failed to parse query: %v", err)
	}

	if parsed.Get("view") != "ADMIN" {
		t.Errorf("expected view=ADMIN, got %s", parsed.Get("view"))
	}

	if parsed.Get("maxPosts") != "10" {
		t.Errorf("expected maxPosts=10, got %s", parsed.Get("maxPosts"))
	}
}

func TestListBlogs(t *testing.T) {
	m := toolstest.NewMockAPI(t)

	blogListJSON := `{
		"items": [
			{"id": "123", "name": "Blog 1"},
			{"id": "456", "name": "Blog 2"}
		]
	}`
	m.Stub("GET", "/v3/users/self/blogs", 200, blogListJSON)

	rt := NewRuntimeWithClient(m.Client())
	h := &blogsHandlers{rt: rt}

	_, out, err := h.listBlogs(t.Context(), nil, listBlogsInput{})
	if err != nil {
		t.Fatalf("listBlogs unexpected error: %v", err)
	}

	if len(out.Items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(out.Items))
	}

	if out.Items[0].Name != "Blog 1" {
		t.Errorf("expected first item name 'Blog 1', got '%s'", out.Items[0].Name)
	}

	calls := m.Calls()
	if len(calls) != 1 {
		t.Fatalf("expected 1 call, got %d", len(calls))
	}

	if calls[0].Path != "/v3/users/self/blogs" {
		t.Errorf("expected path /v3/users/self/blogs, got %s", calls[0].Path)
	}
}

func TestGetBlogByUrl(t *testing.T) {
	m := toolstest.NewMockAPI(t)

	blogJSON := `{
		"id": "789",
		"name": "URL Blog",
		"url": "https://example.blogspot.com"
	}`
	m.Stub("GET", "/v3/blogs/byurl", 200, blogJSON)

	rt := NewRuntimeWithClient(m.Client())
	h := &blogsHandlers{rt: rt}

	_, out, err := h.getBlogByUrl(t.Context(), nil, getBlogByUrlInput{
		URL: "https://example.blogspot.com",
	})
	if err != nil {
		t.Fatalf("getBlogByUrl unexpected error: %v", err)
	}

	if out.Name != "URL Blog" {
		t.Errorf("expected name 'URL Blog', got '%s'", out.Name)
	}

	calls := m.Calls()
	if len(calls) != 1 {
		t.Fatalf("expected 1 call, got %d", len(calls))
	}

	if calls[0].Path != "/v3/blogs/byurl" {
		t.Errorf("expected path /v3/blogs/byurl, got %s", calls[0].Path)
	}

	parsed, err := url.ParseQuery(calls[0].RawQuery)
	if err != nil {
		t.Fatalf("failed to parse query: %v", err)
	}

	if parsed.Get("url") != "https://example.blogspot.com" {
		t.Errorf("expected url=https://example.blogspot.com, got %s", parsed.Get("url"))
	}
}

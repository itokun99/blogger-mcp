package posts

import (
	"encoding/json"
	"net/url"
	"testing"

	"github.com/itokun99/blogger-mcp/internal/tools"
	"github.com/itokun99/blogger-mcp/internal/toolstest"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func parseQuery(t *testing.T, rawQuery string) url.Values {
	t.Helper()

	values, err := url.ParseQuery(rawQuery)
	if err != nil {
		t.Fatalf("parsing query %q: %v", rawQuery, err)
	}

	return values
}

func wantCall(t *testing.T, call toolstest.RecordedCall, method, path string) {
	t.Helper()

	if call.Method != method {
		t.Errorf("expected method %s, got %s", method, call.Method)
	}
	if call.Path != path {
		t.Errorf("expected path %s, got %s", path, call.Path)
	}
}

func TestListPosts(t *testing.T) {
	m := toolstest.NewMockAPI(t)
	m.Stub("GET", "/v3/blogs/111/posts", 200, `{
		"items": [{"id": "222", "title": "First Post", "status": "DRAFT"}],
		"nextPageToken": "TOKEN-2"
	}`)

	h := &handlers{rt: tools.NewRuntimeWithClient(m.Client())}

	_, out, err := h.listPosts(t.Context(), nil, listPostsInput{
		BlogID:     "111",
		MaxResults: toolstest.Ptr(int64(5)),
		Status:     toolstest.Ptr("DRAFT"),
		Labels:     []string{"a", "b"},
		OrderBy:    toolstest.Ptr("PUBLISHED"),
		SortOrder:  toolstest.Ptr("ASCENDING"),
		FetchBody:  toolstest.Ptr(true),
		View:       toolstest.Ptr("ADMIN"),
	})
	if err != nil {
		t.Fatalf("listPosts unexpected error: %v", err)
	}

	if len(out.Items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(out.Items))
	}
	if out.Items[0].Title != "First Post" {
		t.Errorf("expected title 'First Post', got %q", out.Items[0].Title)
	}
	if out.NextPageToken != "TOKEN-2" {
		t.Errorf("expected next page token 'TOKEN-2', got %q", out.NextPageToken)
	}

	calls := m.Calls()
	if len(calls) != 1 {
		t.Fatalf("expected 1 call, got %d", len(calls))
	}
	wantCall(t, calls[0], "GET", "/v3/blogs/111/posts")

	query := parseQuery(t, calls[0].RawQuery)
	want := map[string]string{
		"maxResults": "5",
		"status":     "DRAFT",
		"labels":     "a,b",
		"orderBy":    "PUBLISHED",
		"sortOption": "ASCENDING",
		"fetchBody":  "true",
		"view":       "ADMIN",
	}
	for key, value := range want {
		if got := query.Get(key); got != value {
			t.Errorf("expected %s=%s, got %q", key, value, got)
		}
	}
}

func TestSearchPosts(t *testing.T) {
	m := toolstest.NewMockAPI(t)
	m.Stub("GET", "/v3/blogs/111/posts/search", 200, `{
		"items": [{"id": "222", "title": "Hit"}]
	}`)

	h := &handlers{rt: tools.NewRuntimeWithClient(m.Client())}

	_, out, err := h.searchPosts(t.Context(), nil, searchPostsInput{
		BlogID: "111",
		Query:  "golang",
	})
	if err != nil {
		t.Fatalf("searchPosts unexpected error: %v", err)
	}

	if len(out.Items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(out.Items))
	}
	if out.Items[0].Title != "Hit" {
		t.Errorf("expected title 'Hit', got %q", out.Items[0].Title)
	}

	calls := m.Calls()
	if len(calls) != 1 {
		t.Fatalf("expected 1 call, got %d", len(calls))
	}
	wantCall(t, calls[0], "GET", "/v3/blogs/111/posts/search")

	if got := parseQuery(t, calls[0].RawQuery).Get("q"); got != "golang" {
		t.Errorf("expected q=golang, got %q", got)
	}
}

func TestGetPost(t *testing.T) {
	m := toolstest.NewMockAPI(t)
	m.Stub("GET", "/v3/blogs/111/posts/222", 200, `{
		"id": "222",
		"title": "Hello",
		"content": "<p>world</p>"
	}`)

	h := &handlers{rt: tools.NewRuntimeWithClient(m.Client())}

	_, out, err := h.getPost(t.Context(), nil, getPostInput{BlogID: "111", PostID: "222"})
	if err != nil {
		t.Fatalf("getPost unexpected error: %v", err)
	}

	if out.Id != "222" {
		t.Errorf("expected id '222', got %q", out.Id)
	}
	if out.Title != "Hello" {
		t.Errorf("expected title 'Hello', got %q", out.Title)
	}
	if out.Content != "<p>world</p>" {
		t.Errorf("expected content '<p>world</p>', got %q", out.Content)
	}

	calls := m.Calls()
	if len(calls) != 1 {
		t.Fatalf("expected 1 call, got %d", len(calls))
	}
	wantCall(t, calls[0], "GET", "/v3/blogs/111/posts/222")
}

func TestGetPostByPath(t *testing.T) {
	m := toolstest.NewMockAPI(t)
	m.Stub("GET", "/v3/blogs/111/posts/bypath", 200, `{
		"id": "222",
		"title": "By Path"
	}`)

	h := &handlers{rt: tools.NewRuntimeWithClient(m.Client())}

	_, out, err := h.getPostByPath(t.Context(), nil, getPostByPathInput{
		BlogID: "111",
		Path:   "/2024/01/hello.html",
	})
	if err != nil {
		t.Fatalf("getPostByPath unexpected error: %v", err)
	}

	if out.Title != "By Path" {
		t.Errorf("expected title 'By Path', got %q", out.Title)
	}

	calls := m.Calls()
	if len(calls) != 1 {
		t.Fatalf("expected 1 call, got %d", len(calls))
	}
	wantCall(t, calls[0], "GET", "/v3/blogs/111/posts/bypath")

	if got := parseQuery(t, calls[0].RawQuery).Get("path"); got != "/2024/01/hello.html" {
		t.Errorf("expected path=/2024/01/hello.html, got %q", got)
	}
}

func TestCreatePost(t *testing.T) {
	m := toolstest.NewMockAPI(t)
	m.Stub("POST", "/v3/blogs/111/posts", 200, `{
		"id": "222",
		"title": "New Post"
	}`)

	h := &handlers{rt: tools.NewRuntimeWithClient(m.Client())}

	_, out, err := h.createPost(t.Context(), nil, createPostInput{
		BlogID:  "111",
		Title:   "New Post",
		Content: "<p>body</p>",
	})
	if err != nil {
		t.Fatalf("createPost unexpected error: %v", err)
	}
	if out.Id != "222" {
		t.Errorf("expected id '222', got %q", out.Id)
	}

	if _, _, err := h.createPost(t.Context(), nil, createPostInput{
		BlogID:  "111",
		Title:   "Live Post",
		Content: "<p>live</p>",
		Draft:   toolstest.Ptr(false),
	}); err != nil {
		t.Fatalf("createPost with draft=false unexpected error: %v", err)
	}

	if _, _, err := h.createPost(t.Context(), nil, createPostInput{
		BlogID:  "111",
		Title:   "Explicit Draft",
		Content: "<p>draft</p>",
		Draft:   toolstest.Ptr(true),
	}); err != nil {
		t.Fatalf("createPost with draft=true unexpected error: %v", err)
	}

	calls := m.Calls()
	if len(calls) != 3 {
		t.Fatalf("expected 3 calls, got %d", len(calls))
	}
	for i, call := range calls {
		wantCall(t, call, "POST", "/v3/blogs/111/posts")

		var body struct {
			Title   string `json:"title"`
			Content string `json:"content"`
			Status  string `json:"status"`
		}
		if err := json.Unmarshal(call.Body, &body); err != nil {
			t.Fatalf("call %d: unmarshaling body: %v", i, err)
		}

		wantStatus := "DRAFT"
		if i == 1 {
			wantStatus = "LIVE"
		}
		if body.Status != wantStatus {
			t.Errorf("call %d: expected status %s, got %q", i, wantStatus, body.Status)
		}
	}
}

func TestUpdatePost(t *testing.T) {
	m := toolstest.NewMockAPI(t)
	m.Stub("GET", "/v3/blogs/111/posts/222", 200, `{
		"id": "222",
		"title": "Old",
		"content": "x",
		"status": "LIVE"
	}`)
	m.Stub("PUT", "/v3/blogs/111/posts/222", 200, `{
		"id": "222",
		"title": "New",
		"content": "x",
		"status": "LIVE"
	}`)

	h := &handlers{rt: tools.NewRuntimeWithClient(m.Client())}

	_, out, err := h.updatePost(t.Context(), nil, updatePostInput{
		BlogID: "111",
		PostID: "222",
		Title:  toolstest.Ptr("New"),
	})
	if err != nil {
		t.Fatalf("updatePost unexpected error: %v", err)
	}
	if out.Title != "New" {
		t.Errorf("expected title 'New', got %q", out.Title)
	}

	calls := m.Calls()
	if len(calls) != 2 {
		t.Fatalf("expected 2 calls, got %d", len(calls))
	}
	wantCall(t, calls[0], "GET", "/v3/blogs/111/posts/222")
	wantCall(t, calls[1], "PUT", "/v3/blogs/111/posts/222")

	var body struct {
		Title   string `json:"title"`
		Content string `json:"content"`
		Status  string `json:"status"`
	}
	if err := json.Unmarshal(calls[1].Body, &body); err != nil {
		t.Fatalf("unmarshaling update body: %v", err)
	}
	if body.Title != "New" {
		t.Errorf("expected update body title 'New', got %q", body.Title)
	}
	if body.Content != "x" {
		t.Errorf("expected update body to keep content 'x', got %q", body.Content)
	}
	if body.Status != "LIVE" {
		t.Errorf("expected update body to keep status 'LIVE', got %q", body.Status)
	}
}

func TestPatchPost(t *testing.T) {
	m := toolstest.NewMockAPI(t)
	m.Stub("PATCH", "/v3/blogs/111/posts/222", 200, `{
		"id": "222",
		"title": "Patched"
	}`)

	h := &handlers{rt: tools.NewRuntimeWithClient(m.Client())}

	_, out, err := h.patchPost(t.Context(), nil, patchPostInput{
		BlogID: "111",
		PostID: "222",
		Title:  toolstest.Ptr("Patched"),
	})
	if err != nil {
		t.Fatalf("patchPost unexpected error: %v", err)
	}
	if out.Title != "Patched" {
		t.Errorf("expected title 'Patched', got %q", out.Title)
	}

	calls := m.Calls()
	if len(calls) != 1 {
		t.Fatalf("expected 1 call, got %d", len(calls))
	}
	wantCall(t, calls[0], "PATCH", "/v3/blogs/111/posts/222")

	var body map[string]any
	if err := json.Unmarshal(calls[0].Body, &body); err != nil {
		t.Fatalf("unmarshaling patch body: %v", err)
	}
	if body["title"] != "Patched" {
		t.Errorf("expected patch body title 'Patched', got %v", body["title"])
	}
	for _, key := range []string{"content", "labels", "status", "id"} {
		if _, ok := body[key]; ok {
			t.Errorf("expected patch body to omit %q", key)
		}
	}
}

func TestDeletePost(t *testing.T) {
	m := toolstest.NewMockAPI(t)
	m.Stub("DELETE", "/v3/blogs/111/posts/222", 200, "")

	h := &handlers{rt: tools.NewRuntimeWithClient(m.Client())}

	_, out, err := h.deletePost(t.Context(), nil, deletePostInput{BlogID: "111", PostID: "222"})
	if err != nil {
		t.Fatalf("deletePost unexpected error: %v", err)
	}

	if !out.Deleted {
		t.Error("expected deleted=true")
	}
	if out.BlogID != "111" {
		t.Errorf("expected blog_id '111', got %q", out.BlogID)
	}
	if out.PostID != "222" {
		t.Errorf("expected post_id '222', got %q", out.PostID)
	}

	calls := m.Calls()
	if len(calls) != 1 {
		t.Fatalf("expected 1 call, got %d", len(calls))
	}
	wantCall(t, calls[0], "DELETE", "/v3/blogs/111/posts/222")
}

func TestPublishPost(t *testing.T) {
	m := toolstest.NewMockAPI(t)
	m.Stub("POST", "/v3/blogs/111/posts/222/publish", 200, `{
		"id": "222",
		"title": "Hello",
		"status": "LIVE"
	}`)

	h := &handlers{rt: tools.NewRuntimeWithClient(m.Client())}

	_, out, err := h.publishPost(t.Context(), nil, publishPostInput{BlogID: "111", PostID: "222"})
	if err != nil {
		t.Fatalf("publishPost unexpected error: %v", err)
	}

	if out.Id != "222" {
		t.Errorf("expected id '222', got %q", out.Id)
	}
	if out.Status != "LIVE" {
		t.Errorf("expected status 'LIVE', got %q", out.Status)
	}

	calls := m.Calls()
	if len(calls) != 1 {
		t.Fatalf("expected 1 call, got %d", len(calls))
	}
	wantCall(t, calls[0], "POST", "/v3/blogs/111/posts/222/publish")
}

func TestRevertPost(t *testing.T) {
	m := toolstest.NewMockAPI(t)
	m.Stub("POST", "/v3/blogs/111/posts/222/revert", 200, `{
		"id": "222",
		"title": "Hello",
		"status": "DRAFT"
	}`)

	h := &handlers{rt: tools.NewRuntimeWithClient(m.Client())}

	_, out, err := h.revertPost(t.Context(), nil, revertPostInput{BlogID: "111", PostID: "222"})
	if err != nil {
		t.Fatalf("revertPost unexpected error: %v", err)
	}

	if out.Id != "222" {
		t.Errorf("expected id '222', got %q", out.Id)
	}
	if out.Status != "DRAFT" {
		t.Errorf("expected status 'DRAFT', got %q", out.Status)
	}

	calls := m.Calls()
	if len(calls) != 1 {
		t.Fatalf("expected 1 call, got %d", len(calls))
	}
	wantCall(t, calls[0], "POST", "/v3/blogs/111/posts/222/revert")
}

func TestRegister(t *testing.T) {
	s := mcp.NewServer(&mcp.Implementation{Name: "blogger-mcp-test", Version: "0.0.0"}, nil)
	Register(s, tools.NewRuntimeWithClient(toolstest.NewMockAPI(t).Client()))

	ctx := t.Context()
	clientTransport, serverTransport := mcp.NewInMemoryTransports()

	serverSession, err := s.Connect(ctx, serverTransport, nil)
	if err != nil {
		t.Fatalf("connecting server: %v", err)
	}
	t.Cleanup(func() { _ = serverSession.Close() })

	clientSession, err := mcp.NewClient(&mcp.Implementation{Name: "test-client", Version: "0.0.0"}, nil).Connect(ctx, clientTransport, nil)
	if err != nil {
		t.Fatalf("connecting client: %v", err)
	}
	t.Cleanup(func() { _ = clientSession.Close() })

	result, err := clientSession.ListTools(ctx, nil)
	if err != nil {
		t.Fatalf("listing tools: %v", err)
	}

	got := make(map[string]bool, len(result.Tools))
	for _, tool := range result.Tools {
		got[tool.Name] = true
	}

	want := []string{
		"blogger_list_posts",
		"blogger_search_posts",
		"blogger_get_post",
		"blogger_get_post_by_path",
		"blogger_create_post",
		"blogger_update_post",
		"blogger_patch_post",
		"blogger_delete_post",
		"blogger_publish_post",
		"blogger_revert_post",
	}
	for _, name := range want {
		if !got[name] {
			t.Errorf("tool %q not registered", name)
		}
	}
	if len(result.Tools) != len(want) {
		t.Errorf("expected %d registered tools, got %d", len(want), len(result.Tools))
	}
}

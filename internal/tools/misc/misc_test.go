package misc

import (
	"net/url"
	"testing"

	"github.com/itokun99/blogger-mcp/internal/tools"
	"github.com/itokun99/blogger-mcp/internal/toolstest"
)

func TestGetUserDefaultsToSelf(t *testing.T) {
	m := toolstest.NewMockAPI(t)

	userJSON := `{
		"id": "self-id",
		"displayName": "Test User",
		"url": "https://www.blogger.com/profile/1"
	}`
	m.Stub("GET", "/v3/users/self", 200, userJSON)

	h := &handlers{rt: tools.NewRuntimeWithClient(m.Client())}

	_, out, err := h.getUser(t.Context(), nil, getUserInput{})
	if err != nil {
		t.Fatalf("getUser unexpected error: %v", err)
	}

	if out.DisplayName != "Test User" {
		t.Errorf("expected displayName 'Test User', got '%s'", out.DisplayName)
	}
	if out.Id != "self-id" {
		t.Errorf("expected id 'self-id', got '%s'", out.Id)
	}

	calls := m.Calls()
	if len(calls) != 1 {
		t.Fatalf("expected 1 call, got %d", len(calls))
	}
	if calls[0].Path != "/v3/users/self" {
		t.Errorf("expected path /v3/users/self, got %s", calls[0].Path)
	}
}

func TestGetUserExplicitID(t *testing.T) {
	m := toolstest.NewMockAPI(t)

	userJSON := `{
		"id": "user-42",
		"displayName": "Explicit User"
	}`
	m.Stub("GET", "/v3/users/user-42", 200, userJSON)

	h := &handlers{rt: tools.NewRuntimeWithClient(m.Client())}

	_, out, err := h.getUser(t.Context(), nil, getUserInput{UserID: toolstest.Ptr("user-42")})
	if err != nil {
		t.Fatalf("getUser unexpected error: %v", err)
	}

	if out.DisplayName != "Explicit User" {
		t.Errorf("expected displayName 'Explicit User', got '%s'", out.DisplayName)
	}
	if out.Id != "user-42" {
		t.Errorf("expected id 'user-42', got '%s'", out.Id)
	}

	calls := m.Calls()
	if len(calls) != 1 {
		t.Fatalf("expected 1 call, got %d", len(calls))
	}
	if calls[0].Path != "/v3/users/user-42" {
		t.Errorf("expected path /v3/users/user-42, got %s", calls[0].Path)
	}
}

func TestGetBlogUserInfo(t *testing.T) {
	m := toolstest.NewMockAPI(t)

	infoJSON := `{
		"kind": "blogger#blogUserInfo",
		"blog": {"id": "123", "name": "Test Blog"},
		"blog_user_info": {"blogId": "123", "hasAdminAccess": true, "userId": "self"}
	}`
	m.Stub("GET", "/v3/users/self/blogs/123", 200, infoJSON)

	h := &handlers{rt: tools.NewRuntimeWithClient(m.Client())}

	_, out, err := h.getBlogUserInfo(t.Context(), nil, getBlogUserInfoInput{BlogID: "123"})
	if err != nil {
		t.Fatalf("getBlogUserInfo unexpected error: %v", err)
	}

	if out.Kind != "blogger#blogUserInfo" {
		t.Errorf("expected kind 'blogger#blogUserInfo', got '%s'", out.Kind)
	}
	if out.Blog == nil || out.Blog.Name != "Test Blog" {
		t.Errorf("expected blog name 'Test Blog', got %+v", out.Blog)
	}
	if out.BlogUserInfo == nil || !out.BlogUserInfo.HasAdminAccess {
		t.Errorf("expected hasAdminAccess true, got %+v", out.BlogUserInfo)
	}

	calls := m.Calls()
	if len(calls) != 1 {
		t.Fatalf("expected 1 call, got %d", len(calls))
	}
	if calls[0].Path != "/v3/users/self/blogs/123" {
		t.Errorf("expected path /v3/users/self/blogs/123, got %s", calls[0].Path)
	}
}

func TestGetPageViewsWithRange(t *testing.T) {
	m := toolstest.NewMockAPI(t)

	viewsJSON := `{
		"blogId": "123",
		"kind": "blogger#page_views",
		"counts": [{"count": "42", "timeRange": "ALL_TIME"}]
	}`
	m.Stub("GET", "/v3/blogs/123/pageviews", 200, viewsJSON)

	h := &handlers{rt: tools.NewRuntimeWithClient(m.Client())}

	_, out, err := h.getPageViews(t.Context(), nil, getPageViewsInput{
		BlogID: "123",
		Range:  toolstest.Ptr("ALL_TIME"),
	})
	if err != nil {
		t.Fatalf("getPageViews unexpected error: %v", err)
	}

	if out.BlogId != "123" {
		t.Errorf("expected blogId '123', got '%s'", out.BlogId)
	}
	if len(out.Counts) != 1 {
		t.Fatalf("expected 1 count, got %d", len(out.Counts))
	}
	if out.Counts[0].Count != 42 {
		t.Errorf("expected count 42, got %d", out.Counts[0].Count)
	}
	if out.Counts[0].TimeRange != "ALL_TIME" {
		t.Errorf("expected timeRange 'ALL_TIME', got '%s'", out.Counts[0].TimeRange)
	}

	calls := m.Calls()
	if len(calls) != 1 {
		t.Fatalf("expected 1 call, got %d", len(calls))
	}
	if calls[0].Path != "/v3/blogs/123/pageviews" {
		t.Errorf("expected path /v3/blogs/123/pageviews, got %s", calls[0].Path)
	}

	parsed, err := url.ParseQuery(calls[0].RawQuery)
	if err != nil {
		t.Fatalf("failed to parse query: %v", err)
	}
	if parsed.Get("range") != "ALL_TIME" {
		t.Errorf("expected range=ALL_TIME, got %s", parsed.Get("range"))
	}
}

func TestGetPageViewsWithoutRange(t *testing.T) {
	m := toolstest.NewMockAPI(t)

	m.Stub("GET", "/v3/blogs/123/pageviews", 200, `{"blogId": "123"}`)

	h := &handlers{rt: tools.NewRuntimeWithClient(m.Client())}

	_, _, err := h.getPageViews(t.Context(), nil, getPageViewsInput{BlogID: "123"})
	if err != nil {
		t.Fatalf("getPageViews unexpected error: %v", err)
	}

	calls := m.Calls()
	if len(calls) != 1 {
		t.Fatalf("expected 1 call, got %d", len(calls))
	}
	if calls[0].Path != "/v3/blogs/123/pageviews" {
		t.Errorf("expected path /v3/blogs/123/pageviews, got %s", calls[0].Path)
	}

	parsed, err := url.ParseQuery(calls[0].RawQuery)
	if err != nil {
		t.Fatalf("failed to parse query: %v", err)
	}
	if _, ok := parsed["range"]; ok {
		t.Errorf("expected no range query parameter, got raw query %q", calls[0].RawQuery)
	}
}

func TestListPostUserInfos(t *testing.T) {
	m := toolstest.NewMockAPI(t)

	listJSON := `{
		"kind": "blogger#postList",
		"items": [{
			"kind": "blogger#postUserInfo",
			"post": {"id": "p1", "title": "First Post"},
			"post_user_info": {"blogId": "123", "hasEditAccess": true, "postId": "p1", "userId": "self"}
		}]
	}`
	m.Stub("GET", "/v3/users/self/blogs/123/posts", 200, listJSON)

	h := &handlers{rt: tools.NewRuntimeWithClient(m.Client())}

	_, out, err := h.listPostUserInfos(t.Context(), nil, listPostUserInfosInput{BlogID: "123"})
	if err != nil {
		t.Fatalf("listPostUserInfos unexpected error: %v", err)
	}

	if out.Kind != "blogger#postList" {
		t.Errorf("expected kind 'blogger#postList', got '%s'", out.Kind)
	}
	if len(out.Items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(out.Items))
	}
	if out.Items[0].Post == nil || out.Items[0].Post.Title != "First Post" {
		t.Errorf("expected post title 'First Post', got %+v", out.Items[0].Post)
	}
	if out.Items[0].PostUserInfo == nil || !out.Items[0].PostUserInfo.HasEditAccess {
		t.Errorf("expected hasEditAccess true, got %+v", out.Items[0].PostUserInfo)
	}

	calls := m.Calls()
	if len(calls) != 1 {
		t.Fatalf("expected 1 call, got %d", len(calls))
	}
	if calls[0].Path != "/v3/users/self/blogs/123/posts" {
		t.Errorf("expected path /v3/users/self/blogs/123/posts, got %s", calls[0].Path)
	}
}

func TestGetPostUserInfo(t *testing.T) {
	m := toolstest.NewMockAPI(t)

	infoJSON := `{
		"kind": "blogger#postUserInfo",
		"post": {"id": "p1", "title": "Hello Post"},
		"post_user_info": {"blogId": "123", "hasEditAccess": true, "postId": "p1", "userId": "user-42"}
	}`
	m.Stub("GET", "/v3/users/user-42/blogs/123/posts/p1", 200, infoJSON)

	h := &handlers{rt: tools.NewRuntimeWithClient(m.Client())}

	_, out, err := h.getPostUserInfo(t.Context(), nil, getPostUserInfoInput{
		BlogID: "123",
		PostID: "p1",
		UserID: toolstest.Ptr("user-42"),
	})
	if err != nil {
		t.Fatalf("getPostUserInfo unexpected error: %v", err)
	}

	if out.Kind != "blogger#postUserInfo" {
		t.Errorf("expected kind 'blogger#postUserInfo', got '%s'", out.Kind)
	}
	if out.Post == nil || out.Post.Title != "Hello Post" {
		t.Errorf("expected post title 'Hello Post', got %+v", out.Post)
	}
	if out.PostUserInfo == nil || out.PostUserInfo.UserId != "user-42" {
		t.Errorf("expected userId 'user-42', got %+v", out.PostUserInfo)
	}

	calls := m.Calls()
	if len(calls) != 1 {
		t.Fatalf("expected 1 call, got %d", len(calls))
	}
	if calls[0].Path != "/v3/users/user-42/blogs/123/posts/p1" {
		t.Errorf("expected path /v3/users/user-42/blogs/123/posts/p1, got %s", calls[0].Path)
	}
}

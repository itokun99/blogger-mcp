package comments

import (
	"net/url"
	"testing"

	"github.com/itokun99/blogger-mcp/internal/tools"
	"github.com/itokun99/blogger-mcp/internal/toolstest"
)

// postCommentPath is the single-comment path as the SDK builds it: the
// comments service appends the post scope to the blog comment path, so the
// comment path is /v3/blogs/{blogId}/comments/posts/{postId}/comments/{commentId}.
const postCommentPath = "/v3/blogs/123/comments/posts/456/comments/789"

func TestListCommentsByBlog(t *testing.T) {
	m := toolstest.NewMockAPI(t)

	listJSON := `{
		"kind": "blogger#commentList",
		"items": [
			{"id": "c1", "content": "First!", "status": "LIVE"},
			{"id": "c2", "content": "Buy shoes", "status": "SPAM"}
		],
		"nextPageToken": "page-2"
	}`
	m.Stub("GET", "/v3/blogs/123/comments", 200, listJSON)

	h := &handlers{rt: tools.NewRuntimeWithClient(m.Client())}

	_, out, err := h.listComments(t.Context(), nil, listCommentsInput{
		BlogID:      "123",
		MaxResults:  toolstest.Ptr(int64(25)),
		PageToken:   toolstest.Ptr("page-1"),
		Status:      toolstest.Ptr("LIVE"),
		View:        toolstest.Ptr("ADMIN"),
		StartDate:   toolstest.Ptr("2024-01-01T00:00:00Z"),
		EndDate:     toolstest.Ptr("2024-12-31T23:59:59Z"),
		FetchBodies: toolstest.Ptr(true),
	})
	if err != nil {
		t.Fatalf("listComments unexpected error: %v", err)
	}

	if len(out.Items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(out.Items))
	}
	if out.Items[0].Content != "First!" {
		t.Errorf("expected first item content 'First!', got %q", out.Items[0].Content)
	}
	if out.NextPageToken != "page-2" {
		t.Errorf("expected next page token 'page-2', got %q", out.NextPageToken)
	}

	calls := m.Calls()
	if len(calls) != 1 {
		t.Fatalf("expected 1 call, got %d", len(calls))
	}
	if calls[0].Method != "GET" {
		t.Errorf("expected method GET, got %s", calls[0].Method)
	}
	if calls[0].Path != "/v3/blogs/123/comments" {
		t.Errorf("expected path /v3/blogs/123/comments, got %s", calls[0].Path)
	}

	parsed, err := url.ParseQuery(calls[0].RawQuery)
	if err != nil {
		t.Fatalf("failed to parse query: %v", err)
	}

	wantQuery := map[string]string{
		"maxResults":  "25",
		"pageToken":   "page-1",
		"status":      "LIVE",
		"view":        "ADMIN",
		"startDate":   "2024-01-01T00:00:00Z",
		"endDate":     "2024-12-31T23:59:59Z",
		"fetchBodies": "true",
	}
	for key, want := range wantQuery {
		if got := parsed.Get(key); got != want {
			t.Errorf("expected %s=%s, got %s", key, want, got)
		}
	}
}

func TestListCommentsByPost(t *testing.T) {
	m := toolstest.NewMockAPI(t)

	listJSON := `{
		"items": [
			{"id": "c1", "content": "On the post", "status": "PENDING"}
		]
	}`
	m.Stub("GET", "/v3/blogs/123/comments/posts/456/comments", 200, listJSON)

	h := &handlers{rt: tools.NewRuntimeWithClient(m.Client())}

	_, out, err := h.listComments(t.Context(), nil, listCommentsInput{
		BlogID:     "123",
		PostID:     toolstest.Ptr("456"),
		MaxResults: toolstest.Ptr(int64(10)),
		Status:     toolstest.Ptr("PENDING"),
	})
	if err != nil {
		t.Fatalf("listComments unexpected error: %v", err)
	}

	if len(out.Items) != 1 || out.Items[0].Status != "PENDING" {
		t.Fatalf("expected 1 PENDING item, got %+v", out.Items)
	}

	calls := m.Calls()
	if len(calls) != 1 {
		t.Fatalf("expected 1 call, got %d", len(calls))
	}
	if calls[0].Path != "/v3/blogs/123/comments/posts/456/comments" {
		t.Errorf("expected path /v3/blogs/123/comments/posts/456/comments, got %s", calls[0].Path)
	}

	parsed, err := url.ParseQuery(calls[0].RawQuery)
	if err != nil {
		t.Fatalf("failed to parse query: %v", err)
	}
	if got := parsed.Get("maxResults"); got != "10" {
		t.Errorf("expected maxResults=10, got %s", got)
	}
	if got := parsed.Get("status"); got != "PENDING" {
		t.Errorf("expected status=PENDING, got %s", got)
	}
}

func TestGetComment(t *testing.T) {
	m := toolstest.NewMockAPI(t)

	commentJSON := `{
		"id": "789",
		"content": "Hello there",
		"status": "LIVE",
		"author": {"displayName": "Alice"}
	}`
	m.Stub("GET", postCommentPath, 200, commentJSON)

	h := &handlers{rt: tools.NewRuntimeWithClient(m.Client())}

	_, out, err := h.getComment(t.Context(), nil, getCommentInput{
		BlogID:    "123",
		PostID:    "456",
		CommentID: "789",
	})
	if err != nil {
		t.Fatalf("getComment unexpected error: %v", err)
	}

	if out.Id != "789" {
		t.Errorf("expected id '789', got %q", out.Id)
	}
	if out.Content != "Hello there" {
		t.Errorf("expected content 'Hello there', got %q", out.Content)
	}
	if out.Status != "LIVE" {
		t.Errorf("expected status 'LIVE', got %q", out.Status)
	}
	if out.Author == nil || out.Author.DisplayName != "Alice" {
		t.Errorf("expected author display name 'Alice', got %+v", out.Author)
	}

	calls := m.Calls()
	if len(calls) != 1 {
		t.Fatalf("expected 1 call, got %d", len(calls))
	}
	if calls[0].Method != "GET" {
		t.Errorf("expected method GET, got %s", calls[0].Method)
	}
	if calls[0].Path != postCommentPath {
		t.Errorf("expected path %s, got %s", postCommentPath, calls[0].Path)
	}
}

func TestApproveComment(t *testing.T) {
	m := toolstest.NewMockAPI(t)

	m.Stub("POST", postCommentPath+"/approve", 200, `{"id": "789", "status": "LIVE"}`)

	h := &handlers{rt: tools.NewRuntimeWithClient(m.Client())}

	_, out, err := h.approveComment(t.Context(), nil, approveCommentInput{
		BlogID:    "123",
		PostID:    "456",
		CommentID: "789",
	})
	if err != nil {
		t.Fatalf("approveComment unexpected error: %v", err)
	}

	if out.Id != "789" || out.Status != "LIVE" {
		t.Errorf("expected approved comment 789 with status LIVE, got id=%q status=%q", out.Id, out.Status)
	}

	calls := m.Calls()
	if len(calls) != 1 {
		t.Fatalf("expected 1 call, got %d", len(calls))
	}
	if calls[0].Method != "POST" {
		t.Errorf("expected method POST, got %s", calls[0].Method)
	}
	if calls[0].Path != postCommentPath+"/approve" {
		t.Errorf("expected path %s, got %s", postCommentPath+"/approve", calls[0].Path)
	}
}

func TestDeleteComment(t *testing.T) {
	m := toolstest.NewMockAPI(t)

	m.Stub("DELETE", postCommentPath, 200, `{}`)

	h := &handlers{rt: tools.NewRuntimeWithClient(m.Client())}

	_, out, err := h.deleteComment(t.Context(), nil, deleteCommentInput{
		BlogID:    "123",
		PostID:    "456",
		CommentID: "789",
	})
	if err != nil {
		t.Fatalf("deleteComment unexpected error: %v", err)
	}

	if !out.Deleted {
		t.Error("expected deleted to be true")
	}
	if out.BlogID != "123" || out.PostID != "456" || out.CommentID != "789" {
		t.Errorf("expected ids 123/456/789, got %s/%s/%s", out.BlogID, out.PostID, out.CommentID)
	}

	calls := m.Calls()
	if len(calls) != 1 {
		t.Fatalf("expected 1 call, got %d", len(calls))
	}
	if calls[0].Method != "DELETE" {
		t.Errorf("expected method DELETE, got %s", calls[0].Method)
	}
	if calls[0].Path != postCommentPath {
		t.Errorf("expected path %s, got %s", postCommentPath, calls[0].Path)
	}
}

func TestMarkCommentSpam(t *testing.T) {
	m := toolstest.NewMockAPI(t)

	m.Stub("POST", postCommentPath+"/spam", 200, `{"id": "789", "status": "SPAM"}`)

	h := &handlers{rt: tools.NewRuntimeWithClient(m.Client())}

	_, out, err := h.markCommentSpam(t.Context(), nil, markCommentSpamInput{
		BlogID:    "123",
		PostID:    "456",
		CommentID: "789",
	})
	if err != nil {
		t.Fatalf("markCommentSpam unexpected error: %v", err)
	}

	if out.Status != "SPAM" {
		t.Errorf("expected status 'SPAM', got %q", out.Status)
	}

	calls := m.Calls()
	if len(calls) != 1 {
		t.Fatalf("expected 1 call, got %d", len(calls))
	}
	if calls[0].Method != "POST" {
		t.Errorf("expected method POST, got %s", calls[0].Method)
	}
	if calls[0].Path != postCommentPath+"/spam" {
		t.Errorf("expected path %s, got %s", postCommentPath+"/spam", calls[0].Path)
	}
}

func TestRemoveCommentContent(t *testing.T) {
	m := toolstest.NewMockAPI(t)

	m.Stub("POST", postCommentPath+"/removecontent", 200, `{"id": "789", "content": "", "status": "EMPTIED"}`)

	h := &handlers{rt: tools.NewRuntimeWithClient(m.Client())}

	_, out, err := h.removeCommentContent(t.Context(), nil, removeCommentContentInput{
		BlogID:    "123",
		PostID:    "456",
		CommentID: "789",
	})
	if err != nil {
		t.Fatalf("removeCommentContent unexpected error: %v", err)
	}

	if out.Content != "" {
		t.Errorf("expected empty content, got %q", out.Content)
	}
	if out.Status != "EMPTIED" {
		t.Errorf("expected status 'EMPTIED', got %q", out.Status)
	}

	calls := m.Calls()
	if len(calls) != 1 {
		t.Fatalf("expected 1 call, got %d", len(calls))
	}
	if calls[0].Method != "POST" {
		t.Errorf("expected method POST, got %s", calls[0].Method)
	}
	if calls[0].Path != postCommentPath+"/removecontent" {
		t.Errorf("expected path %s, got %s", postCommentPath+"/removecontent", calls[0].Path)
	}
}

// Package posts implements the Blogger posts MCP tools.
package posts

import (
	"context"
	"fmt"

	"github.com/itokun99/blogger-go/gen/schemas"
	"github.com/itokun99/blogger-go/gen/services"
	"github.com/itokun99/blogger-mcp/internal/tools"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// Post status values used when creating posts. blogger-go exports status
// constants for pages (services.PageStatusLive/PageStatusDraft) but not for
// posts, so they are defined locally.
const (
	statusDraft = "DRAFT"
	statusLive  = "LIVE"
)

type handlers struct {
	rt *tools.Runtime
}

type deletePostOutput = struct {
	Deleted bool   `json:"deleted"`
	BlogID  string `json:"blog_id"`
	PostID  string `json:"post_id"`
}

type listPostsInput struct {
	BlogID      string   `json:"blog_id" jsonschema:"Required. The ID of the blog whose posts to list."`
	MaxResults  *int64   `json:"max_results,omitempty" jsonschema:"Maximum number of posts to return in a single page (1-100)."`
	PageToken   *string  `json:"page_token,omitempty" jsonschema:"Pagination token for the page of results to fetch, taken from the next_page_token of a previous response."`
	Status      *string  `json:"status,omitempty" jsonschema:"Restrict to posts in this status: LIVE, DRAFT, SCHEDULED, or SOFT_TRASHED."`
	Labels      []string `json:"labels,omitempty" jsonschema:"Restrict to posts carrying these labels."`
	OrderBy     *string  `json:"order_by,omitempty" jsonschema:"Order results by PUBLISHED or UPDATED."`
	SortOrder   *string  `json:"sort_order,omitempty" jsonschema:"Sort direction: DESCENDING or ASCENDING."`
	StartDate   *string  `json:"start_date,omitempty" jsonschema:"Restrict to posts published no earlier than this RFC 3339 date-time."`
	EndDate     *string  `json:"end_date,omitempty" jsonschema:"Restrict to posts published no later than this RFC 3339 date-time."`
	FetchBody   *bool    `json:"fetch_body,omitempty" jsonschema:"Include the body content of each post."`
	FetchImages *bool    `json:"fetch_images,omitempty" jsonschema:"Include the images of each post."`
	View        *string  `json:"view,omitempty" jsonschema:"Access level for the returned posts: READER, AUTHOR, or ADMIN."`
}

type searchPostsInput struct {
	BlogID      string   `json:"blog_id" jsonschema:"Required. The ID of the blog whose posts to search."`
	Query       string   `json:"query" jsonschema:"Required. The search query string."`
	MaxResults  *int64   `json:"max_results,omitempty" jsonschema:"Maximum number of posts to return in a single page (1-100)."`
	PageToken   *string  `json:"page_token,omitempty" jsonschema:"Pagination token for the page of results to fetch, taken from the next_page_token of a previous response."`
	Status      *string  `json:"status,omitempty" jsonschema:"Restrict to posts in this status: LIVE, DRAFT, SCHEDULED, or SOFT_TRASHED."`
	Labels      []string `json:"labels,omitempty" jsonschema:"Restrict to posts carrying these labels."`
	OrderBy     *string  `json:"order_by,omitempty" jsonschema:"Order results by PUBLISHED or UPDATED."`
	SortOrder   *string  `json:"sort_order,omitempty" jsonschema:"Sort direction: DESCENDING or ASCENDING."`
	StartDate   *string  `json:"start_date,omitempty" jsonschema:"Restrict to posts published no earlier than this RFC 3339 date-time."`
	EndDate     *string  `json:"end_date,omitempty" jsonschema:"Restrict to posts published no later than this RFC 3339 date-time."`
	FetchBody   *bool    `json:"fetch_body,omitempty" jsonschema:"Include the body content of each post."`
	FetchImages *bool    `json:"fetch_images,omitempty" jsonschema:"Include the images of each post."`
	View        *string  `json:"view,omitempty" jsonschema:"Access level for the returned posts: READER, AUTHOR, or ADMIN."`
}

type getPostInput struct {
	BlogID string `json:"blog_id" jsonschema:"Required. The ID of the blog containing the post."`
	PostID string `json:"post_id" jsonschema:"Required. The ID of the post to retrieve."`
}

type getPostByPathInput struct {
	BlogID string `json:"blog_id" jsonschema:"Required. The ID of the blog containing the post."`
	Path   string `json:"path" jsonschema:"Required. The blog path of the post, for example /2024/01/hello.html."`
}

type createPostInput struct {
	BlogID  string   `json:"blog_id" jsonschema:"Required. The ID of the blog to create the post on."`
	Title   string   `json:"title" jsonschema:"Required. The title of the post."`
	Content string   `json:"content" jsonschema:"Required. The content of the post; may contain HTML markup."`
	Labels  []string `json:"labels,omitempty" jsonschema:"Labels to tag the post with."`
	Draft   *bool    `json:"draft,omitempty" jsonschema:"Create the post as a draft. Defaults to true; set to false to publish it immediately."`
}

type updatePostInput struct {
	BlogID  string   `json:"blog_id" jsonschema:"Required. The ID of the blog containing the post."`
	PostID  string   `json:"post_id" jsonschema:"Required. The ID of the post to update."`
	Title   *string  `json:"title,omitempty" jsonschema:"New title. Omit to keep the current title."`
	Content *string  `json:"content,omitempty" jsonschema:"New content. Omit to keep the current content."`
	Labels  []string `json:"labels,omitempty" jsonschema:"New labels, replacing the current labels. Omit to keep the current labels."`
	Status  *string  `json:"status,omitempty" jsonschema:"New status: LIVE, DRAFT, SCHEDULED, or SOFT_TRASHED. Omit to keep the current status."`
}

type patchPostInput struct {
	BlogID  string   `json:"blog_id" jsonschema:"Required. The ID of the blog containing the post."`
	PostID  string   `json:"post_id" jsonschema:"Required. The ID of the post to patch."`
	Title   *string  `json:"title,omitempty" jsonschema:"New title. Omit to leave the title unchanged."`
	Content *string  `json:"content,omitempty" jsonschema:"New content. Omit to leave the content unchanged."`
	Labels  []string `json:"labels,omitempty" jsonschema:"New labels, replacing the current labels. Omit to leave the labels unchanged."`
}

type deletePostInput struct {
	BlogID string `json:"blog_id" jsonschema:"Required. The ID of the blog containing the post."`
	PostID string `json:"post_id" jsonschema:"Required. The ID of the post to delete."`
}

type publishPostInput struct {
	BlogID string `json:"blog_id" jsonschema:"Required. The ID of the blog containing the post."`
	PostID string `json:"post_id" jsonschema:"Required. The ID of the post to publish."`
}

type revertPostInput struct {
	BlogID string `json:"blog_id" jsonschema:"Required. The ID of the blog containing the post."`
	PostID string `json:"post_id" jsonschema:"Required. The ID of the post to revert."`
}

// Register adds the Blogger posts tools to the server.
func Register(s *mcp.Server, rt *tools.Runtime) {
	h := &handlers{rt: rt}

	mcp.AddTool(s, &mcp.Tool{
		Name:        "blogger_list_posts",
		Description: "List the posts of a blog, with optional filters, ordering, and pagination.",
	}, h.listPosts)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "blogger_search_posts",
		Description: "Search the posts of a blog by a query string, with optional filters.",
	}, h.searchPosts)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "blogger_get_post",
		Description: "Get a post by its ID.",
	}, h.getPost)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "blogger_get_post_by_path",
		Description: "Get a post by its blog path, for example /2024/01/hello.html.",
	}, h.getPostByPath)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "blogger_create_post",
		Description: "Create a post on a blog. The post is created as a DRAFT unless draft is false.",
	}, h.createPost)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "blogger_update_post",
		Description: "Update a post by reading it and applying only the provided fields.",
	}, h.updatePost)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "blogger_patch_post",
		Description: "Patch a post, updating only the provided fields without reading it first.",
	}, h.patchPost)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "blogger_delete_post",
		Description: "Delete a post.",
	}, h.deletePost)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "blogger_publish_post",
		Description: "Publish a post, making it publicly visible.",
	}, h.publishPost)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "blogger_revert_post",
		Description: "Revert a post, discarding its unpublished changes.",
	}, h.revertPost)
}

func (h *handlers) listPosts(ctx context.Context, req *mcp.CallToolRequest, in listPostsInput) (*mcp.CallToolResult, *schemas.PostList, error) {
	c, err := h.rt.Client()
	if err != nil {
		return nil, nil, fmt.Errorf("blogger_list_posts: getting client: %w", err)
	}

	list, err := c.Posts().List(ctx, in.BlogID, listPostOptions(in)...)
	if err != nil {
		return nil, nil, fmt.Errorf("blogger_list_posts: listing posts: %w", err)
	}

	return nil, list, nil
}

func (h *handlers) searchPosts(ctx context.Context, req *mcp.CallToolRequest, in searchPostsInput) (*mcp.CallToolResult, *schemas.PostList, error) {
	c, err := h.rt.Client()
	if err != nil {
		return nil, nil, fmt.Errorf("blogger_search_posts: getting client: %w", err)
	}

	list, err := c.Posts().Search(ctx, in.BlogID, in.Query, searchPostOptions(in)...)
	if err != nil {
		return nil, nil, fmt.Errorf("blogger_search_posts: searching posts: %w", err)
	}

	return nil, list, nil
}

func (h *handlers) getPost(ctx context.Context, req *mcp.CallToolRequest, in getPostInput) (*mcp.CallToolResult, *schemas.Post, error) {
	c, err := h.rt.Client()
	if err != nil {
		return nil, nil, fmt.Errorf("blogger_get_post: getting client: %w", err)
	}

	post, err := c.Posts().Get(ctx, in.BlogID, in.PostID)
	if err != nil {
		return nil, nil, fmt.Errorf("blogger_get_post: getting post: %w", err)
	}

	return nil, post, nil
}

func (h *handlers) getPostByPath(ctx context.Context, req *mcp.CallToolRequest, in getPostByPathInput) (*mcp.CallToolResult, *schemas.Post, error) {
	c, err := h.rt.Client()
	if err != nil {
		return nil, nil, fmt.Errorf("blogger_get_post_by_path: getting client: %w", err)
	}

	post, err := c.Posts().GetByPath(ctx, in.BlogID, in.Path)
	if err != nil {
		return nil, nil, fmt.Errorf("blogger_get_post_by_path: getting post: %w", err)
	}

	return nil, post, nil
}

func (h *handlers) createPost(ctx context.Context, req *mcp.CallToolRequest, in createPostInput) (*mcp.CallToolResult, *schemas.Post, error) {
	c, err := h.rt.Client()
	if err != nil {
		return nil, nil, fmt.Errorf("blogger_create_post: getting client: %w", err)
	}

	status := statusDraft
	if in.Draft != nil && !*in.Draft {
		status = statusLive
	}

	post, err := c.Posts().Insert(ctx, in.BlogID, &schemas.Post{
		Title:   in.Title,
		Content: in.Content,
		Labels:  in.Labels,
		Status:  status,
	})
	if err != nil {
		return nil, nil, fmt.Errorf("blogger_create_post: creating post: %w", err)
	}

	return nil, post, nil
}

func (h *handlers) updatePost(ctx context.Context, req *mcp.CallToolRequest, in updatePostInput) (*mcp.CallToolResult, *schemas.Post, error) {
	c, err := h.rt.Client()
	if err != nil {
		return nil, nil, fmt.Errorf("blogger_update_post: getting client: %w", err)
	}

	current, err := c.Posts().Get(ctx, in.BlogID, in.PostID)
	if err != nil {
		return nil, nil, fmt.Errorf("blogger_update_post: getting post: %w", err)
	}

	if in.Title != nil {
		current.Title = *in.Title
	}
	if in.Content != nil {
		current.Content = *in.Content
	}
	if in.Labels != nil {
		current.Labels = in.Labels
	}
	if in.Status != nil {
		current.Status = *in.Status
	}

	updated, err := c.Posts().Update(ctx, in.BlogID, in.PostID, current)
	if err != nil {
		return nil, nil, fmt.Errorf("blogger_update_post: updating post: %w", err)
	}

	return nil, updated, nil
}

func (h *handlers) patchPost(ctx context.Context, req *mcp.CallToolRequest, in patchPostInput) (*mcp.CallToolResult, *schemas.Post, error) {
	c, err := h.rt.Client()
	if err != nil {
		return nil, nil, fmt.Errorf("blogger_patch_post: getting client: %w", err)
	}

	body := &schemas.Post{}
	if in.Title != nil {
		body.Title = *in.Title
	}
	if in.Content != nil {
		body.Content = *in.Content
	}
	if in.Labels != nil {
		body.Labels = in.Labels
	}

	patched, err := c.Posts().Patch(ctx, in.BlogID, in.PostID, body)
	if err != nil {
		return nil, nil, fmt.Errorf("blogger_patch_post: patching post: %w", err)
	}

	return nil, patched, nil
}

func (h *handlers) deletePost(ctx context.Context, req *mcp.CallToolRequest, in deletePostInput) (*mcp.CallToolResult, deletePostOutput, error) {
	c, err := h.rt.Client()
	if err != nil {
		return nil, deletePostOutput{}, fmt.Errorf("blogger_delete_post: getting client: %w", err)
	}

	if err := c.Posts().Delete(ctx, in.BlogID, in.PostID); err != nil {
		return nil, deletePostOutput{}, fmt.Errorf("blogger_delete_post: deleting post: %w", err)
	}

	return nil, deletePostOutput{
		Deleted: true,
		BlogID:  in.BlogID,
		PostID:  in.PostID,
	}, nil
}

func (h *handlers) publishPost(ctx context.Context, req *mcp.CallToolRequest, in publishPostInput) (*mcp.CallToolResult, *schemas.Post, error) {
	c, err := h.rt.Client()
	if err != nil {
		return nil, nil, fmt.Errorf("blogger_publish_post: getting client: %w", err)
	}

	post, err := c.Posts().Publish(ctx, in.BlogID, in.PostID)
	if err != nil {
		return nil, nil, fmt.Errorf("blogger_publish_post: publishing post: %w", err)
	}

	return nil, post, nil
}

func (h *handlers) revertPost(ctx context.Context, req *mcp.CallToolRequest, in revertPostInput) (*mcp.CallToolResult, *schemas.Post, error) {
	c, err := h.rt.Client()
	if err != nil {
		return nil, nil, fmt.Errorf("blogger_revert_post: getting client: %w", err)
	}

	post, err := c.Posts().Revert(ctx, in.BlogID, in.PostID)
	if err != nil {
		return nil, nil, fmt.Errorf("blogger_revert_post: reverting post: %w", err)
	}

	return nil, post, nil
}

func listPostOptions(in listPostsInput) []services.PostOption {
	var opts []services.PostOption

	if in.MaxResults != nil {
		opts = append(opts, services.WithMaxResults(*in.MaxResults))
	}
	if in.PageToken != nil {
		opts = append(opts, services.WithPageToken(*in.PageToken))
	}
	if in.Status != nil {
		opts = append(opts, services.WithStatus(*in.Status))
	}
	if in.Labels != nil {
		opts = append(opts, services.WithLabels(in.Labels...))
	}
	if in.OrderBy != nil {
		opts = append(opts, services.WithOrderBy(*in.OrderBy))
	}
	if in.SortOrder != nil {
		opts = append(opts, services.WithSortOrder(*in.SortOrder))
	}
	if in.StartDate != nil {
		opts = append(opts, services.WithStartDate(*in.StartDate))
	}
	if in.EndDate != nil {
		opts = append(opts, services.WithEndDate(*in.EndDate))
	}
	if in.FetchBody != nil {
		opts = append(opts, services.WithFetchBody(*in.FetchBody))
	}
	if in.FetchImages != nil {
		opts = append(opts, services.WithFetchImages(*in.FetchImages))
	}
	if in.View != nil {
		opts = append(opts, services.WithView(*in.View))
	}

	return opts
}

func searchPostOptions(in searchPostsInput) []services.PostOption {
	var opts []services.PostOption

	if in.MaxResults != nil {
		opts = append(opts, services.WithMaxResults(*in.MaxResults))
	}
	if in.PageToken != nil {
		opts = append(opts, services.WithPageToken(*in.PageToken))
	}
	if in.Status != nil {
		opts = append(opts, services.WithStatus(*in.Status))
	}
	if in.Labels != nil {
		opts = append(opts, services.WithLabels(in.Labels...))
	}
	if in.OrderBy != nil {
		opts = append(opts, services.WithOrderBy(*in.OrderBy))
	}
	if in.SortOrder != nil {
		opts = append(opts, services.WithSortOrder(*in.SortOrder))
	}
	if in.StartDate != nil {
		opts = append(opts, services.WithStartDate(*in.StartDate))
	}
	if in.EndDate != nil {
		opts = append(opts, services.WithEndDate(*in.EndDate))
	}
	if in.FetchBody != nil {
		opts = append(opts, services.WithFetchBody(*in.FetchBody))
	}
	if in.FetchImages != nil {
		opts = append(opts, services.WithFetchImages(*in.FetchImages))
	}
	if in.View != nil {
		opts = append(opts, services.WithView(*in.View))
	}

	return opts
}

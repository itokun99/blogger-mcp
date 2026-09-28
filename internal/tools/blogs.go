package tools

import (
	"context"
	"fmt"

	"github.com/itokun99/blogger-go/gen/schemas"
	"github.com/itokun99/blogger-go/gen/services"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type blogsHandlers struct {
	rt *Runtime
}

type listBlogsInput struct {
	UserID *string `json:"user_id,omitempty" jsonschema:"User ID to list blogs for. Defaults to 'self' for the authenticated user."`
}

type getBlogInput struct {
	BlogID   string  `json:"blog_id" jsonschema:"Required. The ID of the blog to retrieve."`
	View     *string `json:"view,omitempty" jsonschema:"Access level for the returned resource. Valid values are READER, AUTHOR, or ADMIN."`
	MaxPosts *int64  `json:"max_posts,omitempty" jsonschema:"Maximum number of posts to include in the response."`
}

type getBlogByUrlInput struct {
	URL  string  `json:"url" jsonschema:"Required. The URL of the blog to retrieve."`
	View *string `json:"view,omitempty" jsonschema:"Access level for the returned resource. Valid values are READER, AUTHOR, or ADMIN."`
}

func RegisterBlogs(s *mcp.Server, rt *Runtime) {
	h := &blogsHandlers{rt: rt}

	mcp.AddTool(s, &mcp.Tool{
		Name:        "blogger_list_blogs",
		Description: "List the blogs the user can access. Defaults to the authenticated user (user_id self).",
	}, h.listBlogs)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "blogger_get_blog",
		Description: "Get a blog by its ID.",
	}, h.getBlog)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "blogger_get_blog_by_url",
		Description: "Get a blog by its URL.",
	}, h.getBlogByUrl)
}

func (h *blogsHandlers) listBlogs(ctx context.Context, req *mcp.CallToolRequest, in listBlogsInput) (*mcp.CallToolResult, *schemas.BlogList, error) {
	userID := "self"
	if in.UserID != nil && *in.UserID != "" {
		userID = *in.UserID
	}

	c, err := h.rt.Client()
	if err != nil {
		return nil, nil, fmt.Errorf("blogger_list_blogs: getting client: %w", err)
	}

	list, err := c.Blogs().ListByUser(ctx, userID)
	if err != nil {
		return nil, nil, fmt.Errorf("blogger_list_blogs: listing blogs: %w", err)
	}

	return nil, list, nil
}

func (h *blogsHandlers) getBlog(ctx context.Context, req *mcp.CallToolRequest, in getBlogInput) (*mcp.CallToolResult, *schemas.Blog, error) {
	var opts []services.BlogsOption

	if in.View != nil {
		opts = append(opts, services.WithBlogView(*in.View))
	}
	if in.MaxPosts != nil {
		opts = append(opts, services.WithBlogMaxPosts(*in.MaxPosts))
	}

	c, err := h.rt.Client()
	if err != nil {
		return nil, nil, fmt.Errorf("blogger_get_blog: getting client: %w", err)
	}

	blog, err := c.Blogs().Get(ctx, in.BlogID, opts...)
	if err != nil {
		return nil, nil, fmt.Errorf("blogger_get_blog: getting blog: %w", err)
	}

	return nil, blog, nil
}

func (h *blogsHandlers) getBlogByUrl(ctx context.Context, req *mcp.CallToolRequest, in getBlogByUrlInput) (*mcp.CallToolResult, *schemas.Blog, error) {
	var opts []services.BlogsOption

	if in.View != nil {
		opts = append(opts, services.WithBlogView(*in.View))
	}

	c, err := h.rt.Client()
	if err != nil {
		return nil, nil, fmt.Errorf("blogger_get_blog_by_url: getting client: %w", err)
	}

	blog, err := c.Blogs().GetByUrl(ctx, in.URL, opts...)
	if err != nil {
		return nil, nil, fmt.Errorf("blogger_get_blog_by_url: getting blog: %w", err)
	}

	return nil, blog, nil
}

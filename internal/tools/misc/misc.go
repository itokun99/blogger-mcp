package misc

import (
	"context"
	"fmt"

	"github.com/itokun99/blogger-go/gen/schemas"
	"github.com/itokun99/blogger-go/gen/services"
	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/itokun99/blogger-mcp/internal/tools"
)

type handlers struct {
	rt *tools.Runtime
}

type getUserInput struct {
	UserID *string `json:"user_id,omitempty" jsonschema:"User ID to retrieve. Defaults to 'self' for the authenticated user."`
}

type getBlogUserInfoInput struct {
	BlogID string  `json:"blog_id" jsonschema:"Required. The ID of the blog to retrieve user info for."`
	UserID *string `json:"user_id,omitempty" jsonschema:"User ID to retrieve info for. Defaults to 'self' for the authenticated user."`
}

type getPageViewsInput struct {
	BlogID string  `json:"blog_id" jsonschema:"Required. The ID of the blog to retrieve page views for."`
	Range  *string `json:"range,omitempty" jsonschema:"Time range of the view counts, for example ALL_TIME or 7DAYS. Passed through unchanged. Omitted to use the API default."`
}

type listPostUserInfosInput struct {
	BlogID string  `json:"blog_id" jsonschema:"Required. The ID of the blog to list post user infos for."`
	UserID *string `json:"user_id,omitempty" jsonschema:"User ID to list post user infos for. Defaults to 'self' for the authenticated user."`
}

type getPostUserInfoInput struct {
	BlogID string  `json:"blog_id" jsonschema:"Required. The ID of the blog containing the post."`
	PostID string  `json:"post_id" jsonschema:"Required. The ID of the post to retrieve user info for."`
	UserID *string `json:"user_id,omitempty" jsonschema:"User ID to retrieve info for. Defaults to 'self' for the authenticated user."`
}

// Register adds the users, blog user info, page views, and post user info
// tools to s.
func Register(s *mcp.Server, rt *tools.Runtime) {
	h := &handlers{rt: rt}

	mcp.AddTool(s, &mcp.Tool{
		Name:        "blogger_get_user",
		Description: "Get a Blogger user profile. Defaults to the authenticated user (user_id self).",
	}, h.getUser)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "blogger_get_blog_user_info",
		Description: "Get the blog and per-user info pair for a blog and user.",
	}, h.getBlogUserInfo)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "blogger_get_page_views",
		Description: "Get the page view counts of a blog, one count per available time range.",
	}, h.getPageViews)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "blogger_list_post_user_infos",
		Description: "List posts with per-user info for a blog and user.",
	}, h.listPostUserInfos)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "blogger_get_post_user_info",
		Description: "Get the post and per-user info pair for a blog, post, and user.",
	}, h.getPostUserInfo)
}

func resolveUserID(in *string) string {
	if in == nil || *in == "" {
		return "self"
	}
	return *in
}

func (h *handlers) getUser(ctx context.Context, req *mcp.CallToolRequest, in getUserInput) (*mcp.CallToolResult, *schemas.User, error) {
	c, err := h.rt.Client()
	if err != nil {
		return nil, nil, fmt.Errorf("blogger_get_user: getting client: %w", err)
	}

	user, err := c.Users().Get(ctx, resolveUserID(in.UserID))
	if err != nil {
		return nil, nil, fmt.Errorf("blogger_get_user: getting user: %w", err)
	}

	return nil, user, nil
}

func (h *handlers) getBlogUserInfo(ctx context.Context, req *mcp.CallToolRequest, in getBlogUserInfoInput) (*mcp.CallToolResult, *schemas.BlogUserInfo, error) {
	c, err := h.rt.Client()
	if err != nil {
		return nil, nil, fmt.Errorf("blogger_get_blog_user_info: getting client: %w", err)
	}

	info, err := c.BlogUserInfos().Get(ctx, resolveUserID(in.UserID), in.BlogID)
	if err != nil {
		return nil, nil, fmt.Errorf("blogger_get_blog_user_info: getting blog user info: %w", err)
	}

	return nil, info, nil
}

func (h *handlers) getPageViews(ctx context.Context, req *mcp.CallToolRequest, in getPageViewsInput) (*mcp.CallToolResult, *schemas.Pageviews, error) {
	var opts []services.PageViewsOption

	if in.Range != nil {
		opts = append(opts, services.WithPageViewRange(*in.Range))
	}

	c, err := h.rt.Client()
	if err != nil {
		return nil, nil, fmt.Errorf("blogger_get_page_views: getting client: %w", err)
	}

	views, err := c.PageViews().Get(ctx, in.BlogID, opts...)
	if err != nil {
		return nil, nil, fmt.Errorf("blogger_get_page_views: getting page views: %w", err)
	}

	return nil, views, nil
}

func (h *handlers) listPostUserInfos(ctx context.Context, req *mcp.CallToolRequest, in listPostUserInfosInput) (*mcp.CallToolResult, *schemas.PostUserInfosList, error) {
	c, err := h.rt.Client()
	if err != nil {
		return nil, nil, fmt.Errorf("blogger_list_post_user_infos: getting client: %w", err)
	}

	list, err := c.PostUserInfos().List(ctx, resolveUserID(in.UserID), in.BlogID)
	if err != nil {
		return nil, nil, fmt.Errorf("blogger_list_post_user_infos: listing post user infos: %w", err)
	}

	return nil, list, nil
}

func (h *handlers) getPostUserInfo(ctx context.Context, req *mcp.CallToolRequest, in getPostUserInfoInput) (*mcp.CallToolResult, *schemas.PostUserInfo, error) {
	c, err := h.rt.Client()
	if err != nil {
		return nil, nil, fmt.Errorf("blogger_get_post_user_info: getting client: %w", err)
	}

	info, err := c.PostUserInfos().Get(ctx, resolveUserID(in.UserID), in.BlogID, in.PostID)
	if err != nil {
		return nil, nil, fmt.Errorf("blogger_get_post_user_info: getting post user info: %w", err)
	}

	return nil, info, nil
}

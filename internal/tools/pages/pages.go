package pages

import (
	"context"
	"fmt"

	"github.com/itokun99/blogger-go/gen/schemas"
	"github.com/itokun99/blogger-go/gen/services"
	"github.com/itokun99/blogger-mcp/internal/tools"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type handlers struct {
	rt *tools.Runtime
}

// deletePageOutput must stay an alias to an anonymous struct: it is the
// output contract of blogger_delete_page, not a named schema.
type deletePageOutput = struct {
	Deleted bool   `json:"deleted"`
	BlogID  string `json:"blog_id"`
	PageID  string `json:"page_id"`
}

type listPagesInput struct {
	BlogID     string  `json:"blog_id" jsonschema:"Required. The ID of the blog whose pages to list."`
	MaxResults *int64  `json:"max_results,omitempty" jsonschema:"Maximum number of pages to include in the result."`
	PageToken  *string `json:"page_token,omitempty" jsonschema:"Pagination token to continue a previous listing."`
	Status     *string `json:"status,omitempty" jsonschema:"Filter by page status. Valid values are LIVE, DRAFT, or SOFT_TRASHED."`
	View       *string `json:"view,omitempty" jsonschema:"Access level for the returned resource. Valid values are READER, AUTHOR, or ADMIN."`
	FetchBody  *bool   `json:"fetch_body,omitempty" jsonschema:"Whether to include each page's HTML body content."`
}

type getPageInput struct {
	BlogID string `json:"blog_id" jsonschema:"Required. The ID of the blog containing the page."`
	PageID string `json:"page_id" jsonschema:"Required. The ID of the page to retrieve."`
}

type createPageInput struct {
	BlogID  string `json:"blog_id" jsonschema:"Required. The ID of the blog to create the page on."`
	Title   string `json:"title" jsonschema:"Required. The title of the page."`
	Content string `json:"content" jsonschema:"Required. The HTML body content of the page."`
	Draft   *bool  `json:"draft,omitempty" jsonschema:"Whether to create the page as a draft. Defaults to true: the page is created as a DRAFT unless draft=false, which creates it as LIVE."`
}

type updatePageInput struct {
	BlogID  string  `json:"blog_id" jsonschema:"Required. The ID of the blog containing the page."`
	PageID  string  `json:"page_id" jsonschema:"Required. The ID of the page to update."`
	Title   *string `json:"title,omitempty" jsonschema:"New title for the page. Fields left unset keep their current value."`
	Content *string `json:"content,omitempty" jsonschema:"New HTML body content for the page. Fields left unset keep their current value."`
	Status  *string `json:"status,omitempty" jsonschema:"New page status, LIVE or DRAFT. Fields left unset keep their current value."`
}

type patchPageInput struct {
	BlogID  string  `json:"blog_id" jsonschema:"Required. The ID of the blog containing the page."`
	PageID  string  `json:"page_id" jsonschema:"Required. The ID of the page to patch."`
	Title   *string `json:"title,omitempty" jsonschema:"New title for the page."`
	Content *string `json:"content,omitempty" jsonschema:"New HTML body content for the page."`
}

type deletePageInput struct {
	BlogID string `json:"blog_id" jsonschema:"Required. The ID of the blog containing the page."`
	PageID string `json:"page_id" jsonschema:"Required. The ID of the page to delete."`
}

type publishPageInput struct {
	BlogID string `json:"blog_id" jsonschema:"Required. The ID of the blog containing the page."`
	PageID string `json:"page_id" jsonschema:"Required. The ID of the page to publish."`
}

type revertPageInput struct {
	BlogID string `json:"blog_id" jsonschema:"Required. The ID of the blog containing the page."`
	PageID string `json:"page_id" jsonschema:"Required. The ID of the page to revert."`
}

func Register(s *mcp.Server, rt *tools.Runtime) {
	h := &handlers{rt: rt}

	mcp.AddTool(s, &mcp.Tool{
		Name:        "blogger_list_pages",
		Description: "List the pages of a blog, optionally filtered by status, view, and body inclusion.",
	}, h.listPages)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "blogger_get_page",
		Description: "Get a page by its ID.",
	}, h.getPage)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "blogger_create_page",
		Description: "Create a page on a blog. Creates a DRAFT by default; set draft=false to create it as LIVE immediately.",
	}, h.createPage)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "blogger_update_page",
		Description: "Update a page: fetches the current page, applies the provided fields, then saves the merged page.",
	}, h.updatePage)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "blogger_patch_page",
		Description: "Patch a page by sending only the provided fields.",
	}, h.patchPage)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "blogger_delete_page",
		Description: "Delete a page.",
	}, h.deletePage)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "blogger_publish_page",
		Description: "Publish a page, making it publicly visible.",
	}, h.publishPage)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "blogger_revert_page",
		Description: "Revert a page, discarding its unpublished changes.",
	}, h.revertPage)
}

func (h *handlers) listPages(ctx context.Context, req *mcp.CallToolRequest, in listPagesInput) (*mcp.CallToolResult, *schemas.PageList, error) {
	var opts []services.PageListOption

	if in.MaxResults != nil {
		opts = append(opts, services.WithPageListMaxResults(*in.MaxResults))
	}
	if in.PageToken != nil {
		opts = append(opts, services.WithPageListToken(*in.PageToken))
	}
	if in.Status != nil {
		opts = append(opts, services.WithPageListStatus(*in.Status))
	}
	if in.View != nil {
		opts = append(opts, services.WithPageView(*in.View))
	}
	if in.FetchBody != nil {
		opts = append(opts, services.WithPageFetchBodies(*in.FetchBody))
	}

	c, err := h.rt.Client()
	if err != nil {
		return nil, nil, fmt.Errorf("blogger_list_pages: getting client: %w", err)
	}

	list, err := c.Pages().List(ctx, in.BlogID, opts...)
	if err != nil {
		return nil, nil, fmt.Errorf("blogger_list_pages: listing pages: %w", err)
	}

	return nil, list, nil
}

func (h *handlers) getPage(ctx context.Context, req *mcp.CallToolRequest, in getPageInput) (*mcp.CallToolResult, *schemas.Page, error) {
	c, err := h.rt.Client()
	if err != nil {
		return nil, nil, fmt.Errorf("blogger_get_page: getting client: %w", err)
	}

	page, err := c.Pages().Get(ctx, in.BlogID, in.PageID)
	if err != nil {
		return nil, nil, fmt.Errorf("blogger_get_page: getting page: %w", err)
	}

	return nil, page, nil
}

func (h *handlers) createPage(ctx context.Context, req *mcp.CallToolRequest, in createPageInput) (*mcp.CallToolResult, *schemas.Page, error) {
	status := services.PageStatusDraft
	if in.Draft != nil && !*in.Draft {
		status = services.PageStatusLive
	}

	page := &schemas.Page{
		Title:   in.Title,
		Content: in.Content,
		Status:  status,
	}

	c, err := h.rt.Client()
	if err != nil {
		return nil, nil, fmt.Errorf("blogger_create_page: getting client: %w", err)
	}

	created, err := c.Pages().Insert(ctx, in.BlogID, page)
	if err != nil {
		return nil, nil, fmt.Errorf("blogger_create_page: creating page: %w", err)
	}

	return nil, created, nil
}

func (h *handlers) updatePage(ctx context.Context, req *mcp.CallToolRequest, in updatePageInput) (*mcp.CallToolResult, *schemas.Page, error) {
	c, err := h.rt.Client()
	if err != nil {
		return nil, nil, fmt.Errorf("blogger_update_page: getting client: %w", err)
	}

	page, err := c.Pages().Get(ctx, in.BlogID, in.PageID)
	if err != nil {
		return nil, nil, fmt.Errorf("blogger_update_page: getting current page: %w", err)
	}

	if in.Title != nil {
		page.Title = *in.Title
	}
	if in.Content != nil {
		page.Content = *in.Content
	}
	if in.Status != nil {
		page.Status = *in.Status
	}

	updated, err := c.Pages().Update(ctx, in.BlogID, in.PageID, page)
	if err != nil {
		return nil, nil, fmt.Errorf("blogger_update_page: updating page: %w", err)
	}

	return nil, updated, nil
}

func (h *handlers) patchPage(ctx context.Context, req *mcp.CallToolRequest, in patchPageInput) (*mcp.CallToolResult, *schemas.Page, error) {
	patch := &schemas.Page{}
	if in.Title != nil {
		patch.Title = *in.Title
	}
	if in.Content != nil {
		patch.Content = *in.Content
	}

	c, err := h.rt.Client()
	if err != nil {
		return nil, nil, fmt.Errorf("blogger_patch_page: getting client: %w", err)
	}

	patched, err := c.Pages().Patch(ctx, in.BlogID, in.PageID, patch)
	if err != nil {
		return nil, nil, fmt.Errorf("blogger_patch_page: patching page: %w", err)
	}

	return nil, patched, nil
}

func (h *handlers) deletePage(ctx context.Context, req *mcp.CallToolRequest, in deletePageInput) (*mcp.CallToolResult, deletePageOutput, error) {
	c, err := h.rt.Client()
	if err != nil {
		return nil, deletePageOutput{}, fmt.Errorf("blogger_delete_page: getting client: %w", err)
	}

	if err := c.Pages().Delete(ctx, in.BlogID, in.PageID); err != nil {
		return nil, deletePageOutput{}, fmt.Errorf("blogger_delete_page: deleting page: %w", err)
	}

	return nil, deletePageOutput{Deleted: true, BlogID: in.BlogID, PageID: in.PageID}, nil
}

func (h *handlers) publishPage(ctx context.Context, req *mcp.CallToolRequest, in publishPageInput) (*mcp.CallToolResult, *schemas.Page, error) {
	c, err := h.rt.Client()
	if err != nil {
		return nil, nil, fmt.Errorf("blogger_publish_page: getting client: %w", err)
	}

	page, err := c.Pages().Publish(ctx, in.BlogID, in.PageID)
	if err != nil {
		return nil, nil, fmt.Errorf("blogger_publish_page: publishing page: %w", err)
	}

	return nil, page, nil
}

func (h *handlers) revertPage(ctx context.Context, req *mcp.CallToolRequest, in revertPageInput) (*mcp.CallToolResult, *schemas.Page, error) {
	c, err := h.rt.Client()
	if err != nil {
		return nil, nil, fmt.Errorf("blogger_revert_page: getting client: %w", err)
	}

	page, err := c.Pages().Revert(ctx, in.BlogID, in.PageID)
	if err != nil {
		return nil, nil, fmt.Errorf("blogger_revert_page: reverting page: %w", err)
	}

	return nil, page, nil
}

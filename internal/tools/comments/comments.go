// Package comments registers the Blogger comment moderation MCP tools.
package comments

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

type listCommentsInput struct {
	BlogID      string  `json:"blog_id" jsonschema:"Required. The ID of the blog containing the comments."`
	PostID      *string `json:"post_id,omitempty" jsonschema:"Restrict the listing to the comments of this post."`
	MaxResults  *int64  `json:"max_results,omitempty" jsonschema:"Maximum number of comments to include in the response."`
	PageToken   *string `json:"page_token,omitempty" jsonschema:"Pagination token to continue a previous listing."`
	Status      *string `json:"status,omitempty" jsonschema:"Restrict the listing to comments with this status, for example LIVE or SPAM."`
	View        *string `json:"view,omitempty" jsonschema:"Access level for the returned resources. Valid values are READER, AUTHOR, or ADMIN."`
	StartDate   *string `json:"start_date,omitempty" jsonschema:"Restrict the listing to comments published no earlier than this RFC 3339 date-time."`
	EndDate     *string `json:"end_date,omitempty" jsonschema:"Restrict the listing to comments published no later than this RFC 3339 date-time."`
	FetchBodies *bool   `json:"fetch_bodies,omitempty" jsonschema:"Whether to include the comment bodies in the response."`
}

type getCommentInput struct {
	BlogID    string `json:"blog_id" jsonschema:"Required. The ID of the blog containing the comment."`
	PostID    string `json:"post_id" jsonschema:"Required. The ID of the post containing the comment."`
	CommentID string `json:"comment_id" jsonschema:"Required. The ID of the comment to retrieve."`
}

type approveCommentInput struct {
	BlogID    string `json:"blog_id" jsonschema:"Required. The ID of the blog containing the comment."`
	PostID    string `json:"post_id" jsonschema:"Required. The ID of the post containing the comment."`
	CommentID string `json:"comment_id" jsonschema:"Required. The ID of the comment to approve."`
}

type deleteCommentInput struct {
	BlogID    string `json:"blog_id" jsonschema:"Required. The ID of the blog containing the comment."`
	PostID    string `json:"post_id" jsonschema:"Required. The ID of the post containing the comment."`
	CommentID string `json:"comment_id" jsonschema:"Required. The ID of the comment to delete."`
}

type markCommentSpamInput struct {
	BlogID    string `json:"blog_id" jsonschema:"Required. The ID of the blog containing the comment."`
	PostID    string `json:"post_id" jsonschema:"Required. The ID of the post containing the comment."`
	CommentID string `json:"comment_id" jsonschema:"Required. The ID of the comment to mark as spam."`
}

type removeCommentContentInput struct {
	BlogID    string `json:"blog_id" jsonschema:"Required. The ID of the blog containing the comment."`
	PostID    string `json:"post_id" jsonschema:"Required. The ID of the post containing the comment."`
	CommentID string `json:"comment_id" jsonschema:"Required. The ID of the comment to remove the content from."`
}

func Register(s *mcp.Server, rt *tools.Runtime) {
	h := &handlers{rt: rt}

	mcp.AddTool(s, &mcp.Tool{
		Name:        "blogger_list_comments",
		Description: "List comments on a blog or on a single post, with filters for status, date range, and pagination.",
	}, h.listComments)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "blogger_get_comment",
		Description: "Get a comment by its ID.",
	}, h.getComment)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "blogger_approve_comment",
		Description: "Approve a comment, marking it as visible.",
	}, h.approveComment)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "blogger_delete_comment",
		Description: "Delete a comment.",
	}, h.deleteComment)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "blogger_mark_comment_spam",
		Description: "Mark a comment as spam.",
	}, h.markCommentSpam)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "blogger_remove_comment_content",
		Description: "Remove a comment's content while keeping the comment itself.",
	}, h.removeCommentContent)
}

func (h *handlers) listComments(ctx context.Context, req *mcp.CallToolRequest, in listCommentsInput) (*mcp.CallToolResult, *schemas.CommentList, error) {
	opts := commentListOptions(in)

	c, err := h.rt.Client()
	if err != nil {
		return nil, nil, fmt.Errorf("blogger_list_comments: getting client: %w", err)
	}

	var list *schemas.CommentList
	if in.PostID != nil && *in.PostID != "" {
		list, err = c.Comments().List(ctx, in.BlogID, *in.PostID, opts...)
	} else {
		list, err = c.Comments().ListByBlog(ctx, in.BlogID, opts...)
	}
	if err != nil {
		return nil, nil, fmt.Errorf("blogger_list_comments: listing comments: %w", err)
	}

	return nil, list, nil
}

func (h *handlers) getComment(ctx context.Context, req *mcp.CallToolRequest, in getCommentInput) (*mcp.CallToolResult, *schemas.Comment, error) {
	c, err := h.rt.Client()
	if err != nil {
		return nil, nil, fmt.Errorf("blogger_get_comment: getting client: %w", err)
	}

	comment, err := c.Comments().Get(ctx, in.BlogID, in.PostID, in.CommentID)
	if err != nil {
		return nil, nil, fmt.Errorf("blogger_get_comment: getting comment: %w", err)
	}

	return nil, comment, nil
}

func (h *handlers) approveComment(ctx context.Context, req *mcp.CallToolRequest, in approveCommentInput) (*mcp.CallToolResult, *schemas.Comment, error) {
	c, err := h.rt.Client()
	if err != nil {
		return nil, nil, fmt.Errorf("blogger_approve_comment: getting client: %w", err)
	}

	comment, err := c.Comments().Approve(ctx, in.BlogID, in.PostID, in.CommentID)
	if err != nil {
		return nil, nil, fmt.Errorf("blogger_approve_comment: approving comment: %w", err)
	}

	return nil, comment, nil
}

func (h *handlers) deleteComment(ctx context.Context, req *mcp.CallToolRequest, in deleteCommentInput) (*mcp.CallToolResult, struct {
	Deleted   bool   `json:"deleted"`
	BlogID    string `json:"blog_id"`
	PostID    string `json:"post_id"`
	CommentID string `json:"comment_id"`
}, error) {
	out := struct {
		Deleted   bool   `json:"deleted"`
		BlogID    string `json:"blog_id"`
		PostID    string `json:"post_id"`
		CommentID string `json:"comment_id"`
	}{}

	c, err := h.rt.Client()
	if err != nil {
		return nil, out, fmt.Errorf("blogger_delete_comment: getting client: %w", err)
	}

	if err := c.Comments().Delete(ctx, in.BlogID, in.PostID, in.CommentID); err != nil {
		return nil, out, fmt.Errorf("blogger_delete_comment: deleting comment: %w", err)
	}

	out.Deleted = true
	out.BlogID = in.BlogID
	out.PostID = in.PostID
	out.CommentID = in.CommentID

	return nil, out, nil
}

func (h *handlers) markCommentSpam(ctx context.Context, req *mcp.CallToolRequest, in markCommentSpamInput) (*mcp.CallToolResult, *schemas.Comment, error) {
	c, err := h.rt.Client()
	if err != nil {
		return nil, nil, fmt.Errorf("blogger_mark_comment_spam: getting client: %w", err)
	}

	comment, err := c.Comments().MarkAsSpam(ctx, in.BlogID, in.PostID, in.CommentID)
	if err != nil {
		return nil, nil, fmt.Errorf("blogger_mark_comment_spam: marking comment as spam: %w", err)
	}

	return nil, comment, nil
}

func (h *handlers) removeCommentContent(ctx context.Context, req *mcp.CallToolRequest, in removeCommentContentInput) (*mcp.CallToolResult, *schemas.Comment, error) {
	c, err := h.rt.Client()
	if err != nil {
		return nil, nil, fmt.Errorf("blogger_remove_comment_content: getting client: %w", err)
	}

	comment, err := c.Comments().RemoveContent(ctx, in.BlogID, in.PostID, in.CommentID)
	if err != nil {
		return nil, nil, fmt.Errorf("blogger_remove_comment_content: removing comment content: %w", err)
	}

	return nil, comment, nil
}

func commentListOptions(in listCommentsInput) []services.CommentListOption {
	var opts []services.CommentListOption

	if in.MaxResults != nil {
		opts = append(opts, services.WithCommentMaxResults(*in.MaxResults))
	}
	if in.PageToken != nil {
		opts = append(opts, services.WithCommentPageToken(*in.PageToken))
	}
	if in.Status != nil {
		opts = append(opts, services.WithCommentStatus(*in.Status))
	}
	if in.View != nil {
		opts = append(opts, services.WithCommentView(*in.View))
	}
	if in.StartDate != nil {
		opts = append(opts, services.WithCommentStartDate(*in.StartDate))
	}
	if in.EndDate != nil {
		opts = append(opts, services.WithCommentEndDate(*in.EndDate))
	}
	if in.FetchBodies != nil {
		opts = append(opts, services.WithCommentFetchBodies(*in.FetchBodies))
	}

	return opts
}

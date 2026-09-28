# blogger-mcp

blogger-mcp is an MCP stdio server for the Blogger v3 REST API, built on the
[blogger-go](https://github.com/itokun99/blogger-go) SDK. It exposes Blogger
blogs, posts, pages, comments, and stats as MCP tools so an AI coding agent
can develop and maintain a blog end to end.

## Features

- **Blogs (3 tools)**: list the blogs a user can reach, then fetch one by ID or
  by URL.
- **Posts (10 tools)**: list, search, fetch, create, update, patch, delete,
  publish, and revert. Listing and search share filters for status, labels,
  date range, ordering, and pagination.
- **Pages (8 tools)**: the same lifecycle for static pages, with a body
  inclusion switch on the listing.
- **Comments (6 tools)**: list and fetch comments, then approve, mark as spam,
  strip content, or delete them.
- **Stats and users (5 tools)**: user profiles, per-user blog and post info,
  and page view counts.

Behavior worth knowing up front:

- **Creates default to DRAFT.** `blogger_create_post` and `blogger_create_page`
  write a draft unless you pass `draft: false`, which publishes immediately.
- **Content is HTML.** Blogger's native format, not Markdown. See
  [Content format](#content-format).
- **Listings return page tokens.** Paginated tools hand back a
  `nextPageToken` you pass straight back as `page_token` on the next call.
- **Comment moderation** covers approve, spam, remove content, and delete.
- **Page view stats** come back as one count per available time range.

## Requirements

- Go 1.27.1 or later. That is the `go` directive in `go.mod`; older toolchains
  refuse to build.
- A Google Cloud project with the Blogger API enabled.
- An OAuth client of type **Desktop app**, downloaded from Google Cloud as
  `credentials.json`.

## Build

```sh
go build -o bin/blogger-mcp .
```

## Authentication

Run the `auth` subcommand once. It opens your browser, waits for the Google
consent screen, and writes a token file that the server reuses from then on.

```sh
./bin/blogger-mcp auth -credentials /path/to/credentials.json -token /path/to/token.json
```

What happens during the run:

1. The OAuth consent URL prints to stderr and your browser opens to it.
2. A local callback server listens on `127.0.0.1:8085` (change with `-port`).
3. After you approve, the code is exchanged for a token, which is written to
   the token path with `0600` permissions.

The token carries a refresh token, so the server renews access on its own at
runtime. You only re-run `auth` if you revoke access or rotate credentials.

If you omit the paths, `auth` falls back to the defaults listed in
[Configuration](#configuration).

## Configuration

Every flag has an environment variable equivalent, and both fall back to a
path under the user config directory.

| Flag | Environment variable | Default |
| --- | --- | --- |
| `-credentials` | `BLOGGER_MCP_CREDENTIALS` | `<user config dir>/blogger-mcp/credentials.json` |
| `-token` | `BLOGGER_MCP_TOKEN` | `<user config dir>/blogger-mcp/token.json` |
| `auth -port` | none | `8085` |
| `-version` | none | `false` (prints the version and exits) |

The user config directory is `os.UserConfigDir()`: on macOS it is
`~/Library/Application Support`, on Linux `~/.config`. So the macOS defaults are
`~/Library/Application Support/blogger-mcp/credentials.json` and
`~/Library/Application Support/blogger-mcp/token.json`. On the rare system that
reports no config directory at all, the server falls back to the current
working directory.

Precedence runs flag first, then environment variable, then the default path.
`-port` and `-version` exist only in the subcommand and main command
respectively, and neither reads an environment variable.

## Using with an MCP client

Clients that read an `mcpServers` block, such as Claude Code and Cursor, need
the absolute path to the binary:

```json
{
  "mcpServers": {
    "blogger-mcp": {
      "command": "/absolute/path/to/bin/blogger-mcp",
      "args": [
        "-credentials",
        "/absolute/path/to/credentials.json",
        "-token",
        "/absolute/path/to/token.json"
      ]
    }
  }
}
```

If you keep both files in the default locations, you can drop `args` entirely
and set the environment variables instead:

```json
{
  "mcpServers": {
    "blogger-mcp": {
      "command": "/absolute/path/to/bin/blogger-mcp",
      "env": {
        "BLOGGER_MCP_CREDENTIALS": "/absolute/path/to/credentials.json",
        "BLOGGER_MCP_TOKEN": "/absolute/path/to/token.json"
      }
    }
  }
}
```

The server speaks stdio, so keep `args` free of anything that writes to stdout.

## Tools reference

Parameters marked required must be present. Optional ones can be omitted.

### Blogs

| Tool | Description | Key parameters |
| --- | --- | --- |
| `blogger_list_blogs` | List the blogs the user can access. Defaults to the authenticated user (user_id self). | `user_id` |
| `blogger_get_blog` | Get a blog by its ID. | `blog_id` (required), `view`, `max_posts` |
| `blogger_get_blog_by_url` | Get a blog by its URL. | `url` (required), `view` |

`view` accepts `READER`, `AUTHOR`, or `ADMIN`.

### Posts

| Tool | Description | Key parameters |
| --- | --- | --- |
| `blogger_list_posts` | List the posts of a blog, with optional filters, ordering, and pagination. | `blog_id` (required), `max_results`, `page_token`, `status`, `labels`, `order_by`, `sort_order`, `start_date`, `end_date`, `fetch_body`, `fetch_images`, `view` |
| `blogger_search_posts` | Search the posts of a blog by a query string, with optional filters. | `blog_id` (required), `query` (required), plus the same filters as `blogger_list_posts` |
| `blogger_get_post` | Get a post by its ID. | `blog_id` (required), `post_id` (required) |
| `blogger_get_post_by_path` | Get a post by its blog path, for example /2024/01/hello.html. | `blog_id` (required), `path` (required) |
| `blogger_create_post` | Create a post on a blog. The post is created as a DRAFT unless draft is false. | `blog_id` (required), `title` (required), `content` (required), `labels`, `draft` |
| `blogger_update_post` | Update a post by reading it and applying only the provided fields. | `blog_id` (required), `post_id` (required), `title`, `content`, `labels`, `status` |
| `blogger_patch_post` | Patch a post, updating only the provided fields without reading it first. | `blog_id` (required), `post_id` (required), `title`, `content`, `labels` |
| `blogger_delete_post` | Delete a post. | `blog_id` (required), `post_id` (required) |
| `blogger_publish_post` | Publish a post, making it publicly visible. | `blog_id` (required), `post_id` (required) |
| `blogger_revert_post` | Revert a post, discarding its unpublished changes. | `blog_id` (required), `post_id` (required) |

`status` accepts `LIVE`, `DRAFT`, `SCHEDULED`, or `SOFT_TRASHED`. `order_by`
takes `PUBLISHED` or `UPDATED`, and `sort_order` takes `DESCENDING` or
`ASCENDING`. `max_results` runs 1 to 100. `blogger_update_post` can also change
status; `blogger_patch_post` cannot.

### Pages

| Tool | Description | Key parameters |
| --- | --- | --- |
| `blogger_list_pages` | List the pages of a blog, optionally filtered by status, view, and body inclusion. | `blog_id` (required), `max_results`, `page_token`, `status`, `view`, `fetch_body` |
| `blogger_get_page` | Get a page by its ID. | `blog_id` (required), `page_id` (required) |
| `blogger_create_page` | Create a page on a blog. Creates a DRAFT by default; set draft=false to create it as LIVE immediately. | `blog_id` (required), `title` (required), `content` (required), `draft` |
| `blogger_update_page` | Update a page: fetches the current page, applies the provided fields, then saves the merged page. | `blog_id` (required), `page_id` (required), `title`, `content`, `status` |
| `blogger_patch_page` | Patch a page by sending only the provided fields. | `blog_id` (required), `page_id` (required), `title`, `content` |
| `blogger_delete_page` | Delete a page. | `blog_id` (required), `page_id` (required) |
| `blogger_publish_page` | Publish a page, making it publicly visible. | `blog_id` (required), `page_id` (required) |
| `blogger_revert_page` | Revert a page, discarding its unpublished changes. | `blog_id` (required), `page_id` (required) |

Page `status` takes `LIVE`, `DRAFT`, or `SOFT_TRASHED`, and `blogger_update_page`
accepts `LIVE` or `DRAFT`.

### Comments

| Tool | Description | Key parameters |
| --- | --- | --- |
| `blogger_list_comments` | List comments on a blog or on a single post, with filters for status, date range, and pagination. | `blog_id` (required), `post_id`, `max_results`, `page_token`, `status`, `view`, `start_date`, `end_date`, `fetch_bodies` |
| `blogger_get_comment` | Get a comment by its ID. | `blog_id` (required), `post_id` (required), `comment_id` (required) |
| `blogger_approve_comment` | Approve a comment, marking it as visible. | `blog_id` (required), `post_id` (required), `comment_id` (required) |
| `blogger_delete_comment` | Delete a comment. | `blog_id` (required), `post_id` (required), `comment_id` (required) |
| `blogger_mark_comment_spam` | Mark a comment as spam. | `blog_id` (required), `post_id` (required), `comment_id` (required) |
| `blogger_remove_comment_content` | Remove a comment's content while keeping the comment itself. | `blog_id` (required), `post_id` (required), `comment_id` (required) |

Omit `post_id` on `blogger_list_comments` to list every comment on the blog.
Supply it to scope the listing to one post. The `status` filter accepts values
such as `LIVE` and `SPAM`.

### Stats and users

| Tool | Description | Key parameters |
| --- | --- | --- |
| `blogger_get_user` | Get a Blogger user profile. Defaults to the authenticated user (user_id self). | `user_id` |
| `blogger_get_blog_user_info` | Get the blog and per-user info pair for a blog and user. | `blog_id` (required), `user_id` |
| `blogger_get_page_views` | Get the page view counts of a blog, one count per available time range. | `blog_id` (required), `range` |
| `blogger_list_post_user_infos` | List posts with per-user info for a blog and user. | `blog_id` (required), `user_id` |
| `blogger_get_post_user_info` | Get the post and per-user info pair for a blog, post, and user. | `blog_id` (required), `post_id` (required), `user_id` |

`user_id` defaults to `self`. The `range` value is passed through to the API
unchanged, so use whatever it accepts, for example `ALL_TIME` or `7DAYS`.
Omitting it falls back to the API default.

## Development

Run the tests and the vet pass before opening a pull request:

```sh
go test ./...
go vet ./...
```

Project layout:

```
.
├── main.go                  # entry point, flag parsing, tool registration
├── bin/                     # build output (created by go build)
└── internal/
    ├── auth/auth.go         # OAuth flow behind the `auth` subcommand
    ├── config/config.go     # flag/env/default resolution and client build
    └── tools/
        ├── runtime.go       # lazy shared Blogger client
        ├── blogs.go         # blog tools
        ├── posts/           # post tools
        ├── pages/           # page tools
        ├── comments/        # comment tools
        ├── misc/            # user info and page view tools
        └── toolstest/       # shared test doubles
```

## Content format

Post and page content is **HTML**, because that is Blogger's native format. The
tools do no Markdown conversion, so write HTML into `content` when creating or
updating a post or page. `blogger_create_post` and `blogger_create_page` describe
the field as HTML markup, and a Markdown fragment stored as-is will render as
literal text on the live blog.

# PROJECT KNOWLEDGE BASE

**Generated:** 2026-09-29
**Commit:** (check git log)
**Branch:** main

## OVERVIEW
MCP stdio server for Blogger v3 REST API, built on blogger-go SDK. Exposes blogs, posts, pages, comments, and stats as MCP tools for AI coding agents. Go 1.27.1+, no Dockerfile, no Makefile.

## STRUCTURE
```
blogger-mcp/
├── main.go              # entry point, flag parsing, tool registration
├── internal/
│   ├── auth/auth.go     # OAuth flow behind `auth` subcommand
│   ├── config/config.go # flag/env/default resolution and client build
│   ├── toolstest/       # shared test doubles
│   └── tools/
│       ├── runtime.go   # lazy shared Blogger client
│       ├── blogs.go     # blog tools (3)
│       ├── posts/       # post tools (10)
│       ├── pages/       # page tools (8)
│       ├── comments/    # comment tools (6)
│       └── misc/        # user info and page view tools (5)
```

## WHERE TO LOOK
| Task | Location | Notes |
|------|----------|-------|
| Add new tool | `internal/tools/<domain>/` | Follow existing pattern: input struct + handler method + Register call |
| Change client init | `internal/tools/runtime.go` | Lazy initialization with mutex |
| Modify auth flow | `internal/auth/auth.go` | OAuth consent URL, callback server, token save |
| Run tests | `go test ./...` | Tests use mock client from toolstest |

## CODE MAP
| Symbol | Type | Location | Refs | Role |
|--------|------|----------|------|------|
| Runtime | struct | internal/tools/runtime.go | 5 | Lazy Blogger client singleton |
| Config | struct | internal/config/config.go | 3 | Flag/env/default resolution |
| RegisterBlogs | func | internal/tools/blogs.go | 1 | Tool registration (3 tools) |
| Register | func | internal/tools/posts/posts.go | 1 | Tool registration (10 tools) |
| Register | func | internal/tools/pages/pages.go | 1 | Tool registration (8 tools) |
| Register | func | internal/tools/comments/comments.go | 1 | Tool registration (6 tools) |
| Register | func | internal/tools/misc/misc.go | 1 | Tool registration (5 tools) |

## CONVENTIONS
- **Tool naming:** `blogger_<resource>_<action>` (e.g., `blogger_list_posts`)
- **Input structs:** named `<action>Input`, exported fields with `jsonschema:""` tags
- **Output structs:** anonymous in handler return, or named `<action>Output` only if JSON shape differs from schema
- **Error wrapping:** `fmt.Errorf("tool_name: %w", err)` - preserves tool name in error path
- **Status constants:** lowercase local aliases for `services.PostStatus*` / `services.PageStatus*` (e.g., `statusDraft`)

## ANTI-PATTERNS
- DO NOT call `blogger.Client` directly - use `Runtime.Client()` for lazy init
- DO NOT forget `jsonschema:` tags on input struct fields - required for MCP tool validation
- DO NOT modify status constants in handlers - use the pre-defined `statusDraft`/`statusLive` aliases

## UNIQUE STYLES
- `update_*` methods fetch current resource, merge fields, then save (read-modify-write)
- `patch_*` methods send only provided fields (partial update without read)
- Delete operations return custom output struct with `Deleted: true` flag

## COMMANDS
```bash
# Build
go build -o bin/blogger-mcp .

# Test
go test ./...

# Vet
go vet ./...

# Auth (one-time)
./bin/blogger-mcp auth -credentials /path/to/credentials.json -token /path/to/token.json
```

## NOTES
- Content is HTML, not Markdown - Blogger's native format
- `blogger_create_post` and `blogger_create_page` default to DRAFT unless `draft=false`
- Paginated listings return `nextPageToken` - pass back as `page_token` on next call
- OAuth callback server listens on `127.0.0.1:8085` during auth flow
- Token file has `0600` permissions after write

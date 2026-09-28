# TOOLS DIRECTORY

**Score:** 18 (high complexity: 4 subdomains + shared infrastructure + tests)

## OVERVIEW
MCP tool implementations for Blogger API resources. Each subdirectory is a distinct domain with its own Register function, input structs, and handlers.

## STRUCTURE
```
tools/
├── runtime.go      # lazy client initialization (shared)
├── blogs.go        # blog tools (no subdirectory)
├── posts/          # post CRUD + publish/revert
├── pages/          # page CRUD + publish/revert
├── comments/       # comment listing + moderation
└── misc/           # user info + page views
```

## WHERE TO LOOK
| Task | Location |
|------|----------|
| Add blog tool | `blogs.go` (inline, not subdirectory) |
| Add post tool | `posts/posts.go` |
| Add page tool | `pages/pages.go` |
| Add comment tool | `comments/comments.go` |
| Add user/stats tool | `misc/misc.go` |
| Shared test helpers | `../toolstest/mock.go` |

## CONVENTIONS
- Subdirectories use `package <name>` (not `package tools`)
- Each domain exports `Register(s *mcp.Server, rt *Runtime)`
- Handlers are methods on a domain-specific struct (e.g., `*handlers`, `*blogsHandlers`)
- Helper functions for option builders are package-private (e.g., `listPostOptions`)

## ANTI-PATTERNS
- DO NOT add new domains without tests in `<domain>_test.go`
- DO NOT modify `runtime.go` without checking all callers
- DO NOT change Register signatures - they are called from main.go in order

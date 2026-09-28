# Changelog

All notable changes to this project are documented in this file.
The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.1.0] - 2026-09-28

### Added

- MCP stdio server for the Blogger v3 REST API exposing 32 tools: blogs (3),
  posts (10), pages (8), comments (6), and users/stats (5), built on
  github.com/itokun99/blogger-go.
- `auth` subcommand running the OAuth desktop flow and writing a refreshable
  token (0600 permissions).
- Flag and environment configuration with `~/.config/blogger-mcp` defaults:
  `-credentials`/`BLOGGER_MCP_CREDENTIALS`, `-token`/`BLOGGER_MCP_TOKEN`,
  `auth -port` (default 8085).
- Safe creation defaults: posts and pages are created as `DRAFT` unless
  `draft=false`; content is HTML.
- Pagination passthrough: listing tools return `nextPageToken` to pass back as
  `page_token`.

### Verified

- gofmt clean; `go build ./...`, `go vet ./...`, `go test ./... -count=1` all
  exit 0 (mock-API tests assert request mapping).
- MCP stdio handshake (`initialize` then `tools/list`) returns exactly the 32
  tools; tool calls without credentials return a graceful setup error.

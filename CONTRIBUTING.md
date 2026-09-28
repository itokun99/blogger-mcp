# Contributing

Thanks for taking the time to contribute to blogger-mcp.

## Requirements

- Go 1.27.1 or newer (the `go` directive in `go.mod`).
- No credentials or network access are needed for the test suite; tests run
  against an in-process mock of the Blogger API (`internal/toolstest`).

## Getting started

```sh
git clone git@github.com:itokun99/blogger-mcp.git
cd blogger-mcp
go build ./...
go test ./...
```

## Before opening a pull request

All four gates must pass:

```sh
gofmt -l .          # must print nothing
go build ./...
go vet ./...
go test ./... -count=1
```

## Adding or changing tools

- Read `internal/tools/blogs.go` first: it is the reference implementation for
  the typed `mcp.AddTool` pattern (exported input-struct fields, `json` and
  `jsonschema` tags, schema inference, error wrapping with the tool name).
- One tool package per resource (`internal/tools/{posts,pages,comments,misc}`);
  register new tools from the package's `Register` function.
- The Blogger API client comes from
  [blogger-go](https://github.com/itokun99/blogger-go). Never guess SDK method
  or option names: read `gen/services/*.go` in that repository. SDK fixes
  belong there, not here.
- Tests must be deterministic: stub HTTP calls with `toolstest.NewMockAPI`,
  assert the outgoing request (path, query, body), and never sleep or hit the
  network.

## Commits

Use Conventional Commits (`feat:`, `fix:`, `docs:`, ...), imperative mood,
English. Keep a change and its tests in the same commit.

## Reporting issues

Open a GitHub issue with the tool name, the arguments you passed, and the exact
error output.

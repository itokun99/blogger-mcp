package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/itokun99/blogger-mcp/internal/auth"
	"github.com/itokun99/blogger-mcp/internal/config"
	"github.com/itokun99/blogger-mcp/internal/tools"
	"github.com/itokun99/blogger-mcp/internal/tools/comments"
	"github.com/itokun99/blogger-mcp/internal/tools/misc"
	"github.com/itokun99/blogger-mcp/internal/tools/pages"
	"github.com/itokun99/blogger-mcp/internal/tools/posts"
)

const version = "0.1.0"

func main() {
	if len(os.Args) > 1 && os.Args[1] == "auth" {
		fs := flag.NewFlagSet("auth", flag.ExitOnError)
		creds := fs.String("credentials", "", "Path to OAuth2 credentials JSON file")
		token := fs.String("token", "", "Path to token JSON file")
		port := fs.Int("port", 8085, "Port for OAuth callback server")
		fs.Parse(os.Args[2:])
		cfg := config.Load(*creds, *token)
		if err := auth.Run(context.Background(), cfg, *port); err != nil {
			fmt.Fprintf(os.Stderr, "auth: %v\n", err)
			os.Exit(1)
		}
		return
	}

	fs := flag.NewFlagSet("blogger-mcp", flag.ExitOnError)
	creds := fs.String("credentials", "", "Path to OAuth2 credentials JSON file")
	token := fs.String("token", "", "Path to token JSON file")
	versionFlag := fs.Bool("version", false, "Print version and exit")
	fs.Parse(os.Args[1:])

	if *versionFlag {
		fmt.Println("blogger-mcp", version)
		os.Exit(0)
	}

	cfg := config.Load(*creds, *token)
	rt := tools.NewRuntime(cfg)
	srv := mcp.NewServer(&mcp.Implementation{Name: "blogger-mcp", Version: version}, nil)

	tools.RegisterBlogs(srv, rt)
	posts.Register(srv, rt)
	pages.Register(srv, rt)
	comments.Register(srv, rt)
	misc.Register(srv, rt)

	if err := srv.Run(context.Background(), &mcp.StdioTransport{}); err != nil {
		fmt.Fprintf(os.Stderr, "server error: %v\n", err)
		os.Exit(1)
	}
}

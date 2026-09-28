package tools

import (
	"context"
	"sync"

	"github.com/itokun99/blogger-go"
	"github.com/itokun99/blogger-mcp/internal/config"
)

type Runtime struct {
	mu          sync.Mutex
	newClient   func(context.Context) (*blogger.Client, error)
	client      *blogger.Client
	err         error
	initialized bool
}

func NewRuntime(cfg config.Config) *Runtime {
	return &Runtime{
		newClient: cfg.Client,
	}
}

func NewRuntimeWithClient(client *blogger.Client) *Runtime {
	return &Runtime{
		client:      client,
		initialized: true,
	}
}

func (rt *Runtime) Client() (*blogger.Client, error) {
	rt.mu.Lock()
	defer rt.mu.Unlock()

	if rt.initialized {
		return rt.client, rt.err
	}

	client, err := rt.newClient(context.Background())
	rt.client = client
	rt.err = err
	rt.initialized = true
	return client, err
}

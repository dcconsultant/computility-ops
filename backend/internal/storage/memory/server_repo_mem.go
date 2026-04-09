package memory

import (
	"context"
	"sync"

	"computility-ops/backend/internal/domain"
)

type ServerRepo struct {
	mu      sync.RWMutex
	servers []domain.Server
}

func NewServerRepo() *ServerRepo { return &ServerRepo{} }

func (r *ServerRepo) ReplaceAll(_ context.Context, servers []domain.Server) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.servers = make([]domain.Server, len(servers))
	copy(r.servers, servers)
	return nil
}

func (r *ServerRepo) List(_ context.Context) ([]domain.Server, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]domain.Server, len(r.servers))
	copy(out, r.servers)
	return out, nil
}

func (r *ServerRepo) Clear(_ context.Context) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.servers = nil
	return nil
}

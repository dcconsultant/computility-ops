package mysql

import (
	"context"
	"errors"

	"computility-ops/backend/internal/domain"
)

type ServerRepo struct{}

func NewServerRepo(_ string) *ServerRepo { return &ServerRepo{} }

func (r *ServerRepo) ReplaceAll(_ context.Context, _ []domain.Server) error {
	return errors.New("mysql repo not implemented in phase 1")
}

func (r *ServerRepo) List(_ context.Context) ([]domain.Server, error) {
	return nil, errors.New("mysql repo not implemented in phase 1")
}

func (r *ServerRepo) Clear(_ context.Context) error {
	return errors.New("mysql repo not implemented in phase 1")
}

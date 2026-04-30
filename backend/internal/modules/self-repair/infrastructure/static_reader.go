package infrastructure

import (
	"context"

	srapp "computility-ops/backend/internal/modules/self-repair/application"
	srdomain "computility-ops/backend/internal/modules/self-repair/domain"
)

var _ srapp.CaseReader = (*StaticReader)(nil)

type StaticReader struct{}

func NewStaticReader() *StaticReader { return &StaticReader{} }

func (r *StaticReader) ListCases(ctx context.Context) ([]srdomain.SelfRepairCase, error) {
	_ = ctx
	return []srdomain.SelfRepairCase{}, nil
}

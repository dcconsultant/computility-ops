package infrastructure

import (
	"context"

	rpapp "computility-ops/backend/internal/modules/replacement-planning/application"
	rpdomain "computility-ops/backend/internal/modules/replacement-planning/domain"
)

var _ rpapp.CandidateReader = (*StaticReader)(nil)

type StaticReader struct{}

func NewStaticReader() *StaticReader { return &StaticReader{} }

func (r *StaticReader) ListCandidates(ctx context.Context) ([]rpdomain.ReplacementCandidate, error) {
	_ = ctx
	return []rpdomain.ReplacementCandidate{}, nil
}

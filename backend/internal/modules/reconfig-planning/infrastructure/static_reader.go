package infrastructure

import (
	"context"

	rcapp "computility-ops/backend/internal/modules/reconfig-planning/application"
	rcdomain "computility-ops/backend/internal/modules/reconfig-planning/domain"
)

var _ rcapp.CandidateReader = (*StaticReader)(nil)

type StaticReader struct{}

func NewStaticReader() *StaticReader { return &StaticReader{} }

func (r *StaticReader) ListCandidates(ctx context.Context) ([]rcdomain.ReconfigCandidate, error) {
	_ = ctx
	return []rcdomain.ReconfigCandidate{}, nil
}

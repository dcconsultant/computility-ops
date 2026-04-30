package infrastructure

import (
	"context"
	"strings"
	"time"

	"computility-ops/backend/internal/domain"
	rpapp "computility-ops/backend/internal/modules/replacement-planning/application"
	rpdomain "computility-ops/backend/internal/modules/replacement-planning/domain"
	"computility-ops/backend/internal/repository"
)

var _ rpapp.CandidateReader = (*LegacyReader)(nil)

type LegacyReader struct {
	serverRepo  repository.ServerRepo
	datasetRepo repository.DatasetRepo
}

func NewLegacyReader(serverRepo repository.ServerRepo, datasetRepo repository.DatasetRepo) *LegacyReader {
	return &LegacyReader{serverRepo: serverRepo, datasetRepo: datasetRepo}
}

func (r *LegacyReader) ListCandidates(ctx context.Context) ([]rpdomain.ReplacementCandidate, error) {
	servers, err := r.serverRepo.List(ctx)
	if err != nil {
		return nil, err
	}
	pkgRates, err := r.datasetRepo.ListPackageFailureRates(ctx)
	if err != nil {
		return nil, err
	}
	pkgMap := map[string]float64{}
	for _, p := range pkgRates {
		cfg := strings.TrimSpace(p.ConfigType)
		if cfg == "" {
			continue
		}
		rate := p.Recent1YFailureRate
		if rate <= 0 {
			rate = p.FailureRate
		}
		if old, ok := pkgMap[cfg]; !ok || old <= 0 {
			pkgMap[cfg] = rate
		}
	}

	now := time.Now()
	out := make([]rpdomain.ReplacementCandidate, 0, len(servers))
	for _, s := range servers {
		age := yearsSince(s.LaunchDate, now)
		afr := pkgMap[strings.TrimSpace(s.ConfigType)]
		maintCost := estimateMaintCost(s, afr)
		currentTCO := maintCost * 3
		out = append(out, rpdomain.ReplacementCandidate{
			AssetID:           strings.TrimSpace(s.SN),
			Model:             strings.TrimSpace(s.Model),
			AgeYears:          age,
			AnnualFailureRate: afr,
			AnnualMaintCost:   maintCost,
			CurrentTCO:        currentTCO,
		})
	}
	return out, nil
}

func yearsSince(launchDate string, now time.Time) float64 {
	if strings.TrimSpace(launchDate) == "" {
		return 0
	}
	if t, ok := parseFlexibleDate(launchDate); ok {
		if now.After(t) {
			return now.Sub(t).Hours() / 24 / 365
		}
	}
	return 0
}

func estimateMaintCost(s domain.Server, afr float64) float64 {
	base := 1000.0
	if strings.Contains(strings.ToLower(strings.TrimSpace(s.Model)), "gpu") {
		base = 3000
	}
	if afr <= 0 {
		afr = 0.02
	}
	return base * (1 + afr*10)
}

func parseFlexibleDate(raw string) (time.Time, bool) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return time.Time{}, false
	}
	layouts := []string{"2006-01-02", "2006/01/02", "2006/1/2", "2006-1-2", "20060102"}
	for _, layout := range layouts {
		if t, err := time.ParseInLocation(layout, raw, time.Local); err == nil {
			return t, true
		}
	}
	return time.Time{}, false
}

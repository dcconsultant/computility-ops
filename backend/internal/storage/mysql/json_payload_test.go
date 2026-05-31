package mysql

import (
	"strings"
	"testing"

	"computility-ops/backend/internal/domain"
)

func TestJSONPayloadRoundTripSupportsCompressedAndPlainPayloads(t *testing.T) {
	small := domain.RenewalPlan{PlanID: "small", TargetDate: "2026-12-31", TargetCores: 128}
	smallPayload, err := marshalJSONPayload(small)
	if err != nil {
		t.Fatalf("marshal small payload: %v", err)
	}
	if strings.HasPrefix(smallPayload, compressedJSONPayloadPrefix) {
		t.Fatalf("small payload should stay plain JSON")
	}
	var decodedSmall domain.RenewalPlan
	if err := unmarshalJSONPayload(smallPayload, &decodedSmall); err != nil {
		t.Fatalf("unmarshal small payload: %v", err)
	}
	if decodedSmall.PlanID != small.PlanID || decodedSmall.TargetCores != small.TargetCores {
		t.Fatalf("small round trip mismatch: %+v", decodedSmall)
	}

	large := domain.RenewalPlan{PlanID: "large", TargetDate: "2026-12-31"}
	for i := 0; i < 2000; i++ {
		large.Items = append(large.Items, domain.RenewalItem{
			Rank:            i + 1,
			SN:              strings.Repeat("SN", 20),
			Manufacturer:    strings.Repeat("vendor", 8),
			Model:           strings.Repeat("model", 8),
			DetailedConfig:  strings.Repeat("cpu/mem/disk;", 16),
			IDC:             "BJ01",
			ConfigType:      "compute",
			CPULogicalCores: 64,
			FinalScore:      99.9,
		})
	}
	largePayload, err := marshalJSONPayload(large)
	if err != nil {
		t.Fatalf("marshal large payload: %v", err)
	}
	if !strings.HasPrefix(largePayload, compressedJSONPayloadPrefix) {
		t.Fatalf("large payload should be compressed")
	}
	var decodedLarge domain.RenewalPlan
	if err := unmarshalJSONPayload(largePayload, &decodedLarge); err != nil {
		t.Fatalf("unmarshal large payload: %v", err)
	}
	if decodedLarge.PlanID != large.PlanID || len(decodedLarge.Items) != len(large.Items) {
		t.Fatalf("large round trip mismatch: plan_id=%s items=%d", decodedLarge.PlanID, len(decodedLarge.Items))
	}
}

func TestUnmarshalRenewalPlanListSummaryOmitsHeavyDetails(t *testing.T) {
	plan := domain.RenewalPlan{
		PlanID:        "summary",
		Status:        "effective",
		EffectiveAt:   "2026-01-01T00:00:00Z",
		TargetDate:    "2026-12-31",
		TargetCores:   128,
		SelectedCores: 64,
		SelectedCount: 1,
		Items:         []domain.RenewalItem{{SN: strings.Repeat("SN", 128), CPULogicalCores: 64}},
		Sections: []domain.RenewalPlanSection{{
			Bucket:        "compute",
			SelectedCores: 64,
			SelectedCount: 1,
			Items:         []domain.RenewalItem{{SN: strings.Repeat("SEC", 128)}},
		}},
		FullRenewal:     &domain.RenewalPlanVariant{Items: []domain.RenewalItem{{SN: strings.Repeat("FULL", 128)}}},
		MinimalRenewal:  &domain.RenewalPlanVariant{Items: []domain.RenewalItem{{SN: strings.Repeat("MIN", 128)}}},
		NonRenewalItems: []domain.NonRenewalItem{{SN: strings.Repeat("NON", 128)}},
		Comparison:      &domain.RenewalComparison{ReducedRenewalItems: []domain.ReducedRenewalItem{{SN: strings.Repeat("RED", 128)}}},
	}
	payload, err := marshalJSONPayload(plan)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}

	summary, err := unmarshalRenewalPlanListSummary(payload)
	if err != nil {
		t.Fatalf("unmarshal summary: %v", err)
	}
	if summary.PlanID != plan.PlanID || summary.Status != plan.Status || summary.SelectedCores != plan.SelectedCores {
		t.Fatalf("summary scalar fields mismatch: %+v", summary)
	}
	if len(summary.Items) != 0 || len(summary.NonRenewalItems) != 0 || summary.FullRenewal != nil || summary.MinimalRenewal != nil || summary.Comparison != nil {
		t.Fatalf("summary should omit heavy fields: %+v", summary)
	}
	if len(summary.Sections) != 1 || summary.Sections[0].SelectedCores != 64 || len(summary.Sections[0].Items) != 0 {
		t.Fatalf("summary sections should keep aggregate fields only: %+v", summary.Sections)
	}
}

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

package service

import (
	"context"
	"testing"

	"computility-ops/backend/internal/domain"
	"computility-ops/backend/internal/storage/memory"
)

func TestNormalizeHeaderName(t *testing.T) {
	got := NormalizeHeaderName("  CPU_Logical-Cores ")
	if got != "cpulogicalcores" {
		t.Fatalf("NormalizeHeaderName() = %q, want %q", got, "cpulogicalcores")
	}
}

func TestMapHeaders_WithAliases(t *testing.T) {
	headers := []string{"序列号", "套餐", "PSA", "保修结束日期", "备注"}
	mapped := MapHeaders(headers, serverHeaderMap)

	want := []string{"sn", "config_type", "psa", "warranty_end_date", "备注"}
	for i := range want {
		if mapped[i] != want[i] {
			t.Fatalf("mapped[%d]=%q, want %q", i, mapped[i], want[i])
		}
	}
}

func TestValidateRequiredHeaders(t *testing.T) {
	err := ValidateRequiredHeaders([]string{"sn", "psa"}, "sn", "psa", "config_type")
	if err == nil {
		t.Fatal("expected missing header error, got nil")
	}

	if err.Error() != "缺少必填列: config_type" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestNormalizeSpecialPolicy(t *testing.T) {
	cases := map[string]string{
		"白名单":      "whitelist",
		"renew":     "whitelist",
		"黑名单":      "blacklist",
		"norenew":   "blacklist",
		"something": "",
	}

	for in, want := range cases {
		got := normalizeSpecialPolicy(in)
		if got != want {
			t.Fatalf("normalizeSpecialPolicy(%q)=%q, want %q", in, got, want)
		}
	}
}

func TestValidateAndReplaceSpecialRules_OnlySNAndPolicy_EnrichFromServers(t *testing.T) {
	ctx := context.Background()
	serverRepo := memory.NewServerRepo()
	datasetRepo := memory.NewDatasetRepo()
	svc := NewImportService(serverRepo, datasetRepo)

	if err := serverRepo.ReplaceAll(ctx, []domain.Server{{
		SN:              "SN001",
		Manufacturer:    "Dell",
		Model:           "R760",
		PSA:             "10",
		IDC:             "SG1",
		ConfigType:      "compute-a",
		WarrantyEndDate: "2027-12-31",
		LaunchDate:      "2023-01-01",
	}}); err != nil {
		t.Fatalf("seed servers failed: %v", err)
	}

	res, err := svc.ValidateAndReplaceSpecialRules(ctx, []map[string]string{{
		"sn":     "SN001",
		"policy": "加白",
	}})
	if err != nil {
		t.Fatalf("ValidateAndReplaceSpecialRules error: %v", err)
	}
	if res.Success != 1 || res.Failed != 0 {
		t.Fatalf("unexpected result: %+v", res)
	}

	rules, err := datasetRepo.ListSpecialRules(ctx)
	if err != nil {
		t.Fatalf("ListSpecialRules error: %v", err)
	}
	if len(rules) != 1 {
		t.Fatalf("rules len=%d, want 1", len(rules))
	}
	r := rules[0]
	if r.SN != "SN001" || r.Policy != "whitelist" {
		t.Fatalf("rule basic fields = %+v", r)
	}
	if r.Manufacturer != "Dell" || r.Model != "R760" || r.PSA != "10" || r.IDC != "SG1" || r.PackageType != "compute-a" || r.WarrantyEndDate != "2027-12-31" || r.LaunchDate != "2023-01-01" {
		t.Fatalf("rule not enriched from server table: %+v", r)
	}
}

func TestValidateAndReplaceSpecialRules_SNNotFound(t *testing.T) {
	ctx := context.Background()
	svc := NewImportService(memory.NewServerRepo(), memory.NewDatasetRepo())

	res, err := svc.ValidateAndReplaceSpecialRules(ctx, []map[string]string{{
		"sn":     "MISSING",
		"policy": "blacklist",
	}})
	if err != nil {
		t.Fatalf("ValidateAndReplaceSpecialRules error: %v", err)
	}
	if res.Success != 0 || res.Failed != 1 || len(res.Errors) != 1 {
		t.Fatalf("unexpected result: %+v", res)
	}
	if res.Errors[0].Reason != "SN 不存在于服务器管理表" {
		t.Fatalf("unexpected error reason: %s", res.Errors[0].Reason)
	}
}

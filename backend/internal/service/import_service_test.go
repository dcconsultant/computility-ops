package service

import (
	"context"
	"testing"
	"time"

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

func TestMapHeaders_SpecialRulesWithAnnotatedHeaders(t *testing.T) {
	headers := []string{"SN", "策略（加白/加黑）", "原因（可选）"}
	mapped := MapHeaders(headers, specialHeaderMap)

	want := []string{"sn", "policy", "reason"}
	for i := range want {
		if mapped[i] != want[i] {
			t.Fatalf("mapped[%d]=%q, want %q", i, mapped[i], want[i])
		}
	}
}

func TestNormalizeSpecialPolicy(t *testing.T) {
	cases := map[string]string{
		"白名单":       "whitelist",
		"renew":     "whitelist",
		"黑名单":       "blacklist",
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
		"reason": "业务连续性保障",
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
	if r.SN != "SN001" || r.Policy != "whitelist" || r.Reason != "业务连续性保障" {
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

func TestBuildStorageTopServerRates_IncludesWarmAndHot(t *testing.T) {
	now := time.Date(2026, 4, 16, 0, 0, 0, 0, time.Local)
	ptrTime := func(v time.Time) *time.Time { return &v }

	servers := []domain.Server{
		{SN: "W-001", Manufacturer: "A", Model: "W", ConfigType: "cfg-warm"},
		{SN: "H-001", Manufacturer: "B", Model: "H", ConfigType: "cfg-hot"},
		{SN: "C-001", Manufacturer: "C", Model: "C", ConfigType: "cfg-compute"},
	}
	faultEventsBySN := map[string][]faultEvent{
		"W-001": {{createdAt: ptrTime(now.AddDate(0, 0, -10))}, {createdAt: ptrTime(now.AddDate(-2, 0, 0))}},
		"H-001": {{createdAt: ptrTime(now.AddDate(0, 0, -20))}, {createdAt: ptrTime(now.AddDate(0, 0, -5))}},
		"C-001": {{createdAt: ptrTime(now.AddDate(0, 0, -3))}},
	}
	pkgMap := map[string]domain.HostPackageConfig{
		"cfg-warm":    {ConfigType: "cfg-warm", SceneCategory: "warm_storage", DataDiskCount: 3, StorageCapacityTB: 24},
		"cfg-hot":     {ConfigType: "cfg-hot", SceneCategory: "hot_storage", DataDiskCount: 1, StorageCapacityTB: 8},
		"cfg-compute": {ConfigType: "cfg-compute", SceneCategory: "compute", DataDiskCount: 0, StorageCapacityTB: 0},
	}

	got := buildStorageTopServerRates(servers, faultEventsBySN, pkgMap, now)
	if len(got) != 2 {
		t.Fatalf("len(got)=%d, want 2", len(got))
	}
	if got[0].SN != "H-001" || got[1].SN != "W-001" {
		t.Fatalf("unexpected order or rows: %+v", got)
	}
	if got[0].FaultCount != 2 || got[0].Denominator != 2 || got[0].FaultRate != 1 {
		t.Fatalf("unexpected hot row metrics: %+v", got[0])
	}
	if got[1].FaultCount != 1 || got[1].Denominator != 4 || got[1].FaultRate != 0.25 {
		t.Fatalf("unexpected warm row metrics: %+v", got[1])
	}
}

func TestListStorageTopServerRatesByBucket(t *testing.T) {
	ctx := context.Background()
	serverRepo := memory.NewServerRepo()
	datasetRepo := memory.NewDatasetRepo()
	svc := NewImportService(serverRepo, datasetRepo)

	if err := serverRepo.ReplaceAll(ctx, []domain.Server{
		{SN: "W-001", ConfigType: "cfg-warm", WarrantyEndDate: "2028-01-01"},
		{SN: "H-001", ConfigType: "cfg-hot", WarrantyEndDate: "2028-01-02"},
	}); err != nil {
		t.Fatalf("seed servers failed: %v", err)
	}
	if err := datasetRepo.ReplaceHostPackages(ctx, []domain.HostPackageConfig{
		{ConfigType: "cfg-warm", SceneCategory: "warm_storage"},
		{ConfigType: "cfg-hot", SceneCategory: "hot_storage"},
	}); err != nil {
		t.Fatalf("seed host packages failed: %v", err)
	}
	if err := datasetRepo.ReplaceStorageTopServerRates(ctx, []domain.StorageTopServerRate{
		{SN: "W-001", ConfigType: "cfg-warm", FaultRate: 0.4},
		{SN: "H-001", ConfigType: "cfg-hot", FaultRate: 0.8},
	}); err != nil {
		t.Fatalf("seed top rates failed: %v", err)
	}

	warmRows, err := svc.ListStorageTopServerRatesByBucket(ctx, "warm_storage")
	if err != nil {
		t.Fatalf("ListStorageTopServerRatesByBucket(warm_storage) error: %v", err)
	}
	if len(warmRows) != 1 || warmRows[0].SN != "W-001" {
		t.Fatalf("unexpected warm rows: %+v", warmRows)
	}

	hotRows, err := svc.ListStorageTopServerRatesByBucket(ctx, "hot_storage")
	if err != nil {
		t.Fatalf("ListStorageTopServerRatesByBucket(hot_storage) error: %v", err)
	}
	if len(hotRows) != 1 || hotRows[0].SN != "H-001" {
		t.Fatalf("unexpected hot rows: %+v", hotRows)
	}

	defaultRows, err := svc.ListStorageTopServerRatesByBucket(ctx, "unknown")
	if err != nil {
		t.Fatalf("ListStorageTopServerRatesByBucket(default) error: %v", err)
	}
	if len(defaultRows) != 1 || defaultRows[0].SN != "W-001" {
		t.Fatalf("unexpected default rows: %+v", defaultRows)
	}
}

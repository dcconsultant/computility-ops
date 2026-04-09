package service

import "testing"

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

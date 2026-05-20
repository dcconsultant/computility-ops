package service

import "testing"

func TestIsPSAExcluded_JSONArrayAndCSV(t *testing.T) {
	rules := splitCSV("/server-decommission")
	if !isPSAExcluded("[\"/server-decommission/cn-decommission/reuse\"]", rules) {
		t.Fatalf("json array psa should be matched")
	}
	if !isPSAExcluded("/server-decommission/cn-decommission/reuse,/x/y", rules) {
		t.Fatalf("csv psa should be matched")
	}
	if isPSAExcluded("/online-product/abc", rules) {
		t.Fatalf("unexpected match")
	}
}

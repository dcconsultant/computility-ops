package service

import (
	"testing"

	"computility-ops/backend/internal/domain"
)

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

func TestCalcReplacementNeedInt_Compute(t *testing.T) {
	servers := make([]domain.Server, 20)
	for i := range servers {
		servers[i] = domain.Server{ConfigType: "c-old", LaunchDate: "2018-01-01", PSA: "/biz/a"}
	}
	pkgByConfig := map[string]domain.HostPackageConfig{
		"c-old": {ConfigType: "c-old", SceneCategory: "compute", CPULogicalCores: 64},
	}
	originByConfig := map[string]float64{"c-old": 10000} // score=0.0001

	need := calcReplacementNeedInt("compute", 100, 100, 0.0002, 0.00015, servers, pkgByConfig, originByConfig, nil, 0)
	if need != 25 { // routine=10, max=25, eligible enough to reach cap
		t.Fatalf("unexpected compute replacement need: %d", need)
	}
}

func TestCalcReplacementNeedFloat_WarmStorage(t *testing.T) {
	servers := []domain.Server{
		{ConfigType: "w-old", LaunchDate: "2017-01-01", PSA: "/biz/a"},
		{ConfigType: "w-old", LaunchDate: "2017-01-01", PSA: "/biz/b"},
		{ConfigType: "w-old", LaunchDate: "2017-01-01", PSA: "/biz/c"},
	}
	pkgByConfig := map[string]domain.HostPackageConfig{
		"w-old": {ConfigType: "w-old", SceneCategory: "warm_storage", StorageCapacityTB: 20},
	}
	originByConfig := map[string]float64{"w-old": 12000} // score≈0.0000833

	need := calcReplacementNeedFloat("warm_storage", 40, 40, 0.0002, 0.0001, servers, pkgByConfig, originByConfig, nil, 0)
	if need != 7 { // routine=4 + eligible3
		t.Fatalf("unexpected warm replacement need: %.2f", need)
	}
}

func TestCalcReplacementNeedFloat_HotStorage(t *testing.T) {
	servers := []domain.Server{
		{ConfigType: "h-old", LaunchDate: "2017-01-01", PSA: "/biz/a"},
		{ConfigType: "h-old", LaunchDate: "2017-01-01", PSA: "/biz/b"},
	}
	pkgByConfig := map[string]domain.HostPackageConfig{
		"h-old": {ConfigType: "h-old", SceneCategory: "hot_storage", StorageCapacityTB: 10},
	}
	originByConfig := map[string]float64{"h-old": 9000} // score≈0.000111

	need := calcReplacementNeedFloat("hot_storage", 30, 30, 0.0002, 0.00012, servers, pkgByConfig, originByConfig, nil, 0)
	if need != 5 { // routine=3 + eligible2
		t.Fatalf("unexpected hot replacement need: %.2f", need)
	}
}

func TestCalcReplacementNeedInt_GPU(t *testing.T) {
	servers := []domain.Server{
		{ConfigType: "g-old", LaunchDate: "2017-01-01", PSA: "/biz/a"},
		{ConfigType: "g-old", LaunchDate: "2017-01-01", PSA: "/biz/b"},
		{ConfigType: "g-old", LaunchDate: "2017-01-01", PSA: "/biz/c"},
	}
	pkgByConfig := map[string]domain.HostPackageConfig{
		"g-old": {ConfigType: "g-old", SceneCategory: "gpu", GPUCardCount: 8},
	}
	originByConfig := map[string]float64{"g-old": 15000} // score≈0.0000667

	need := calcReplacementNeedInt("gpu", 40, 40, 0.0002, 0.0001, servers, pkgByConfig, originByConfig, nil, 0)
	if need != 7 { // routine=4 + eligible3
		t.Fatalf("unexpected gpu replacement need: %d", need)
	}
}

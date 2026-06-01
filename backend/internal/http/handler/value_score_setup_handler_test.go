package handler

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"computility-ops/backend/internal/domain"
	"computility-ops/backend/internal/service"
	"computility-ops/backend/internal/storage/memory"
	"github.com/gin-gonic/gin"
	"github.com/xuri/excelize/v2"
)

func TestValueScoreSetupHandler_ExportMonthlyTCO_ExportsFrontendColumns(t *testing.T) {
	ctx := context.Background()
	repo := memory.NewDatasetRepo()
	if err := repo.SetValueScoreCostParams(ctx, domain.ValueScoreCostParams{
		DepreciationMonths: 60,
		CabinetUtilization: 1,
		RatedPowerKW:       10,
		MonthlyRentCNY:     1000,
	}); err != nil {
		t.Fatal(err)
	}
	if err := repo.ReplaceHostPackages(ctx, []domain.HostPackageConfig{{
		ConfigType:       "compute-a",
		SceneCategory:    "计算",
		CPULogicalCores:  8,
		MemoryCapacityGB: 64,
		PowerWatts:       1000,
	}}); err != nil {
		t.Fatal(err)
	}
	if err := repo.ReplaceValueScoreOriginalValues(ctx, []domain.ValueScoreOriginalValue{{ConfigType: "compute-a", ServerOriginalCNY: 60000}}); err != nil {
		t.Fatal(err)
	}
	if err := repo.ReplaceValueScorePerformanceParams(ctx, []domain.ValueScorePerformanceParam{{ConfigType: "compute-a", PerformanceScore: 2277}}); err != nil {
		t.Fatal(err)
	}

	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := NewValueScoreSetupHandler(service.NewValueScoreSetupService(repo, nil))
	r.GET("/export", h.ExportMonthlyTCO)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/export", nil))
	if w.Code != http.StatusOK {
		t.Fatalf("status=%d, want 200", w.Code)
	}

	xf, err := excelize.OpenReader(bytes.NewReader(w.Body.Bytes()))
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = xf.Close() }()
	rows, err := xf.GetRows("价值分")
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 2 {
		t.Fatalf("len(rows)=%d, want 2", len(rows))
	}
	if len(rows[0]) != 28 {
		t.Fatalf("len(headers)=%d, want 28", len(rows[0]))
	}
	if rows[0][0] != "配置类型" || rows[0][27] != "价值分v1" {
		t.Fatalf("unexpected headers: first=%q last=%q", rows[0][0], rows[0][27])
	}
	if rows[1][0] != "compute-a" || rows[1][26] != "133.75 /核" {
		t.Fatalf("unexpected exported row: config_type=%q unit_tco=%q", rows[1][0], rows[1][26])
	}
}

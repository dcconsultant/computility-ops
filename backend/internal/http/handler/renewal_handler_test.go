package handler

import (
	"bytes"
	"testing"

	"computility-ops/backend/internal/domain"
	"github.com/xuri/excelize/v2"
)

func TestBuildXLSX_IncludesPerformanceAndSingleDiskCapacityColumns(t *testing.T) {
	buf, err := buildXLSX(domain.RenewalPlan{
		PlanID:        "100",
		TargetDate:    "2026-12-31",
		SelectedCount: 1,
		Items: []domain.RenewalItem{{
			SN:                   "SN-1",
			ConfigType:           "compute-a",
			Model:                "M1",
			CPUPerformanceScore:  2500,
			SingleDiskCapacityTB: 4,
		}},
	})
	if err != nil {
		t.Fatalf("buildXLSX() error = %v", err)
	}

	xf, err := excelize.OpenReader(bytes.NewReader(buf.Bytes()))
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = xf.Close() }()

	rows, err := xf.GetRows("国内")
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 2 {
		t.Fatalf("len(rows)=%d, want 2", len(rows))
	}
	if rows[0][4] != "CPU性能分" || rows[0][5] != "单盘容量（TB）" {
		t.Fatalf("unexpected headers: %+v", rows[0])
	}
	if rows[1][4] != "2500" || rows[1][5] != "4" {
		t.Fatalf("unexpected metric cells: %+v", rows[1])
	}
}

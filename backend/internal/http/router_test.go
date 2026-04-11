package http

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"computility-ops/backend/internal/http/handler"
	"computility-ops/backend/internal/service"
	mem "computility-ops/backend/internal/storage/memory"
)

type healthResp struct {
	Code int    `json:"code"`
	Msg  string `json:"message"`
	Data struct {
		Status        string `json:"status"`
		StorageDriver string `json:"storage_driver"`
	} `json:"data"`
}

func TestNewRouter_Healthz(t *testing.T) {
	serverRepo := mem.NewServerRepo()
	datasetRepo := mem.NewDatasetRepo()
	renewalRepo := mem.NewRenewalRepo()

	importSvc := service.NewImportService(serverRepo, datasetRepo)
	renewalSvc := service.NewRenewalService(serverRepo, datasetRepo, renewalRepo)

	r := NewRouter(Handlers{
		Import:        handler.NewImportHandler(importSvc),
		Renewal:       handler.NewRenewalHandler(renewalSvc),
		System:        handler.NewSystemHandler(),
		StorageDriver: "memory",
	})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/healthz", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status=%d, want 200", w.Code)
	}

	var got healthResp
	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
		t.Fatalf("unmarshal response error: %v", err)
	}
	if got.Code != 0 || got.Msg != "ok" {
		t.Fatalf("unexpected response envelope: %+v", got)
	}
	if got.Data.Status != "ok" || got.Data.StorageDriver != "memory" {
		t.Fatalf("unexpected response data: %+v", got.Data)
	}
	if w.Header().Get("X-Request-Id") == "" {
		t.Fatal("missing X-Request-Id header")
	}
}

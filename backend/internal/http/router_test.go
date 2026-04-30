package http

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"computility-ops/backend/internal/http/handler"
	rcapi "computility-ops/backend/internal/modules/reconfig-planning/api"
	rcapp "computility-ops/backend/internal/modules/reconfig-planning/application"
	rcinfra "computility-ops/backend/internal/modules/reconfig-planning/infrastructure"
	rpapi "computility-ops/backend/internal/modules/replacement-planning/api"
	rpapp "computility-ops/backend/internal/modules/replacement-planning/application"
	rpinfra "computility-ops/backend/internal/modules/replacement-planning/infrastructure"
	srapi "computility-ops/backend/internal/modules/self-repair/api"
	srapp "computility-ops/backend/internal/modules/self-repair/application"
	srinfra "computility-ops/backend/internal/modules/self-repair/infrastructure"
	"computility-ops/backend/internal/service"
	mem "computility-ops/backend/internal/storage/memory"
	"github.com/gin-gonic/gin"
)

type healthResp struct {
	Code int    `json:"code"`
	Msg  string `json:"message"`
	Data struct {
		Status        string `json:"status"`
		StorageDriver string `json:"storage_driver"`
	} `json:"data"`
}

type envelopeResp struct {
	Code int             `json:"code"`
	Msg  string          `json:"message"`
	Data json.RawMessage `json:"data"`
}

func buildTestRouter() *gin.Engine {
	serverRepo := mem.NewServerRepo()
	datasetRepo := mem.NewDatasetRepo()
	renewalRepo := mem.NewRenewalRepo()

	importSvc := service.NewImportService(serverRepo, datasetRepo)
	renewalSvc := service.NewRenewalService(serverRepo, datasetRepo, renewalRepo)

	return NewRouter(Handlers{
		Import:              handler.NewImportHandler(importSvc),
		Renewal:             handler.NewRenewalHandler(renewalSvc),
		System:              handler.NewSystemHandler(),
		StorageDriver:       "memory",
		ReplacementPlanning: rpapi.NewHandler(rpapp.NewService(rpinfra.NewStaticReader())),
		ReconfigPlanning:    rcapi.NewHandler(rcapp.NewService(rcinfra.NewStaticReader())),
		SelfRepair:          srapi.NewHandler(srapp.NewService(srinfra.NewStaticReader())),
	})
}

func TestNewRouter_Healthz(t *testing.T) {
	r := buildTestRouter()

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

func TestNewRouter_DecisionRoutesContract(t *testing.T) {
	r := buildTestRouter()
	paths := []string{
		"/api/v1/ops/decisions/replacement",
		"/api/v1/ops/decisions/reconfig",
		"/api/v1/ops/decisions/self-repair",
	}

	for _, p := range paths {
		req := httptest.NewRequest(http.MethodGet, p, nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("path=%s status=%d, want 200", p, w.Code)
		}
		var got envelopeResp
		if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
			t.Fatalf("path=%s unmarshal error: %v", p, err)
		}
		if got.Code != 0 || got.Msg != "ok" {
			t.Fatalf("path=%s unexpected envelope: %+v", p, got)
		}
	}
}

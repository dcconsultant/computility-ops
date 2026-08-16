package http

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
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
	contractRepo := mem.NewContractRepo()
	supplierRepo := mem.NewSupplierRepo()
	deliveryRepo := mem.NewDeliveryRepo()

	importSvc := service.NewImportService(serverRepo, datasetRepo)
	renewalSvc := service.NewRenewalService(serverRepo, datasetRepo, renewalRepo)
	contractSvc := service.NewContractService(contractRepo)
	supplierSvc := service.NewSupplierService(supplierRepo, contractRepo)
	deliverySvc := service.NewDeliveryService(deliveryRepo)
	deliveryDecisionSvc := service.NewDeliveryDecisionService()

	return NewRouter(Handlers{
		Import:              handler.NewImportHandler(importSvc),
		Renewal:             handler.NewRenewalHandler(renewalSvc),
		Contract:            handler.NewContractHandler(contractSvc),
		Supplier:            handler.NewSupplierHandler(supplierSvc),
		Delivery:            handler.NewDeliveryHandler(deliverySvc),
		DeliveryDecision:    handler.NewDeliveryDecisionHandler(deliveryDecisionSvc),
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
		"/api/v1/delivery-decision/defaults",
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

func TestNewRouter_DeliveryDecisionCalculate(t *testing.T) {
	r := buildTestRouter()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/delivery-decision/calculate", strings.NewReader(`{"input":{"hw_total":277000,"hw_cores":128,"hw_tax_rate":0.13,"idc_rent_monthly":4639,"idc_rack_kw":3.52,"idc_fill_rate":1.2,"idc_server_power_w":570,"idc_network_depreciation":3.62,"cloud_memory_ratio":8,"cloud_disk_ratio":0,"cloud_cpu_daily_price":0.78543,"cloud_memory_daily_price":0.1122,"cloud_disk_daily_price":0.00732,"cloud_tax_rate":0.06,"depreciation_years":7,"wacc_rate":0.03,"residual_rate":0.05,"country":"China","currency":"CNY","cloud_current_discount":0.25}}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status=%d, want 200", w.Code)
	}
	var got envelopeResp
	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
		t.Fatalf("unmarshal response error: %v", err)
	}
	if got.Code != 0 || got.Msg != "ok" {
		t.Fatalf("unexpected response envelope: %+v", got)
	}
}

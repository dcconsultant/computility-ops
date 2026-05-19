package http

import (
	"computility-ops/backend/internal/http/handler"
	"computility-ops/backend/internal/http/middleware"
	rcapi "computility-ops/backend/internal/modules/reconfig-planning/api"
	renewalapi "computility-ops/backend/internal/modules/renewal/api"
	rpapi "computility-ops/backend/internal/modules/replacement-planning/api"
	srapi "computility-ops/backend/internal/modules/self-repair/api"
	"github.com/gin-gonic/gin"
)

type Handlers struct {
	Import              *handler.ImportHandler
	Renewal             *handler.RenewalHandler
	Contract            *handler.ContractHandler
	Cabinet             *handler.CabinetHandler
	System              *handler.SystemHandler
	ValueScoreSetup     *handler.ValueScoreSetupHandler
	MetaData            *handler.MetaHandler
	ReconfigMgmt        *handler.ReconfigHandler
	ResourcePlanning    *handler.ResourcePlanningHandler
	StorageDriver       string
	ReplacementPlanning *rpapi.Handler
	ReconfigPlanning    *rcapi.Handler
	SelfRepair          *srapi.Handler
	RenewalRead         *renewalapi.LegacyQueryAdapter
}

func NewRouter(h Handlers) *gin.Engine {
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(middleware.RequestID())
	r.Use(middleware.Audit())

	v1 := r.Group("/api/v1")
	{
		v1.GET("/healthz", func(c *gin.Context) {
			c.Set("audit_action", "healthz")
			c.Set("audit_result", "ok")
			c.JSON(200, gin.H{"code": 0, "message": "ok", "data": gin.H{"status": "ok", "storage_driver": h.StorageDriver}})
		})

		v1.POST("/servers/import", h.Import.ImportServers)
		v1.GET("/servers", h.Import.ListServers)
		v1.GET("/servers/package-anomalies/export", h.Import.ExportServerPackageAnomalies)

		v1.POST("/host-packages/import", h.Import.ImportHostPackages)
		v1.GET("/host-packages", h.Import.ListHostPackages)
		v1.GET("/host-packages/template/export", h.Import.ExportHostPackagesTemplate)

		v1.GET("/value-score/cabinet-baseline", h.ValueScoreSetup.GetCabinetBaseline)
		v1.GET("/value-score/cost-params", h.ValueScoreSetup.GetCostParams)
		v1.PUT("/value-score/cost-params", h.ValueScoreSetup.UpdateCostParams)
		v1.POST("/value-score/cost-params/import", h.ValueScoreSetup.ImportCostParams)
		v1.GET("/value-score/cost-params/template/export", h.ValueScoreSetup.ExportCostParamsTemplate)
		v1.GET("/value-score/original-values", h.ValueScoreSetup.ListOriginalValues)
		v1.POST("/value-score/original-values/import", h.ValueScoreSetup.ImportOriginalValues)
		v1.GET("/value-score/original-values/template/export", h.ValueScoreSetup.ExportOriginalValuesTemplate)
		v1.GET("/value-score/performance-params", h.ValueScoreSetup.ListPerformanceParams)
		v1.POST("/value-score/performance-params/import", h.ValueScoreSetup.ImportPerformanceParams)
		v1.POST("/value-score/performance-params/preview", h.ValueScoreSetup.PreviewPerformanceParams)
		v1.GET("/value-score/performance-params/template/export", h.ValueScoreSetup.ExportPerformanceParamsTemplate)
		v1.POST("/value-score/performance/calculate", h.ValueScoreSetup.CalculatePerformance)
		v1.POST("/value-score/tco/calculate", h.ValueScoreSetup.CalculateMonthlyTCO)
		v1.GET("/value-score/tco/export", h.ValueScoreSetup.ExportMonthlyTCO)

		v1.GET("/cabinet-config/utilization", h.Cabinet.GetUtilization)
		v1.PUT("/cabinet-config/utilization", h.Cabinet.UpdateUtilization)
		v1.POST("/cabinet-config/import", h.Cabinet.Import)
		v1.GET("/cabinet-config/template/export", h.Cabinet.ExportTemplate)
		v1.GET("/cabinet-config", h.Cabinet.List)
		v1.POST("/cabinet-config", h.Cabinet.Create)
		v1.PUT("/cabinet-config/:id", h.Cabinet.Update)
		v1.DELETE("/cabinet-config/:id", h.Cabinet.Delete)

		v1.POST("/special-rules/import", h.Import.ImportSpecialRules)
		v1.GET("/special-rules", h.Import.ListSpecialRules)

		v1.POST("/failure-rates/model/import", h.Import.ImportModelFailureRates)
		v1.GET("/failure-rates/model", h.Import.ListModelFailureRates)
		v1.POST("/failure-rates/package/import", h.Import.ImportPackageFailureRates)
		v1.GET("/failure-rates/package", h.Import.ListPackageFailureRates)
		v1.POST("/failure-rates/package-model/import", h.Import.ImportPackageModelFailureRates)
		v1.GET("/failure-rates/package-model", h.Import.ListPackageModelFailureRates)
		v1.GET("/failure-rates/overall", h.Import.ListOverallFailureRates)
		v1.GET("/failure-rates/overview-cards", h.Import.ListFailureOverviewCards)
		v1.GET("/failure-rates/age-trend", h.Import.ListFailureAgeTrendPoints)
		v1.GET("/failure-rates/features", h.Import.ListFailureFeatureFacts)
		v1.GET("/failure-rates/storage-top-servers", h.Import.ListStorageTopServerRates)
		v1.GET("/failure-rates/storage-top-servers/export", h.Import.ExportWarmStorageServers)
		v1.POST("/failure-rates/analyze/import", h.Import.AnalyzeFaultRates)
		v1.POST("/failure-rates/year-fault-analysis/export", h.Import.ExportYearFaultAnalysis)

		v1.POST("/system/mysql/test", h.System.TestMySQLConnection)
		v1.GET("/system/import-errors", h.System.ListImportErrors)

		v1.POST("/contracts", h.Contract.CreateContract)
		v1.GET("/contracts", h.Contract.ListContracts)
		v1.GET("/contracts/:contract_id", h.Contract.GetContract)
		v1.PUT("/contracts/:contract_id", h.Contract.UpdateContract)
		v1.DELETE("/contracts/:contract_id", h.Contract.DeleteContract)
		v1.POST("/contracts/:contract_id/attachments", h.Contract.UploadAttachment)
		v1.GET("/contracts/:contract_id/attachments/:attachment_id/download", h.Contract.DownloadAttachment)
		v1.DELETE("/contracts/:contract_id/attachments/:attachment_id", h.Contract.DeleteAttachment)

		v1.POST("/renewals/plan", h.Renewal.CreatePlan)
		v1.GET("/renewals/plans", h.Renewal.ListPlans)
		v1.GET("/renewals/plans/:plan_id", h.Renewal.GetPlan)
		v1.DELETE("/renewals/plans/:plan_id", h.Renewal.DeletePlan)
		v1.GET("/renewals/plans/:plan_id/export", h.Renewal.ExportPlan)
		v1.GET("/renewals/plans/:plan_id/non-renewal/export", h.Renewal.ExportNonRenewal)
		v1.GET("/renewals/settings", h.Renewal.GetSettings)
		v1.PUT("/renewals/settings", h.Renewal.UpdateSettings)
		v1.GET("/renewals/unit-prices", h.Renewal.ListUnitPrices)
		v1.PUT("/renewals/unit-prices", h.Renewal.UpdateUnitPrices)

		v1.GET("/ops/decisions/replacement", h.ReplacementPlanning.ListSuggestions)
		v1.GET("/ops/decisions/reconfig", h.ReconfigPlanning.ListSuggestions)
		v1.GET("/resource-planning/config", h.ResourcePlanning.GetConfig)
		v1.POST("/resource-planning/config", h.ResourcePlanning.SaveConfig)
		v1.POST("/resource-planning/calculate", h.ResourcePlanning.Calculate)
		v1.POST("/reconfig/plan/calculate", h.ReconfigMgmt.CalculatePlan)
		v1.POST("/reconfig/plan/start", h.ReconfigMgmt.StartPlan)
		v1.GET("/reconfig/plan/progress/:job_id", h.ReconfigMgmt.GetPlanProgress)
		v1.GET("/reconfig/plan/result/:job_id", h.ReconfigMgmt.GetPlanResult)
		v1.GET("/reconfig/plan/result/:job_id/actions/export", h.ReconfigMgmt.ExportPlanResultActions)
		v1.GET("/reconfig/plans", h.ReconfigMgmt.ListSavedPlans)
		v1.GET("/reconfig/plans/:plan_id", h.ReconfigMgmt.GetSavedPlan)
		v1.GET("/reconfig/plans/:plan_id/actions/export", h.ReconfigMgmt.ExportSavedPlanActions)
		v1.POST("/meta/models", h.MetaData.CreateModel)
		v1.GET("/meta/models", h.MetaData.ListModels)
		v1.GET("/meta/models/:model_id", h.MetaData.GetModel)
		v1.PUT("/meta/models/:model_id", h.MetaData.UpdateModel)
		v1.POST("/meta/models/:model_id/archive", h.MetaData.ArchiveModel)
		v1.POST("/meta/models/:model_id/clone", h.MetaData.CloneModel)
		v1.DELETE("/meta/models/:model_id", h.MetaData.DeleteModel)

		v1.POST("/meta/models/:model_id/fields", h.MetaData.CreateField)
		v1.PUT("/meta/models/:model_id/fields/:field_id", h.MetaData.UpdateField)
		v1.DELETE("/meta/models/:model_id/fields/:field_id", h.MetaData.DeleteField)
		v1.PUT("/meta/models/:model_id/fields/reorder", h.MetaData.ReorderFields)

		v1.POST("/meta/models/:model_id/references", h.MetaData.CreateReference)
		v1.PUT("/meta/models/:model_id/references/:ref_id", h.MetaData.UpdateReference)
		v1.DELETE("/meta/models/:model_id/references/:ref_id", h.MetaData.DeleteReference)
		v1.GET("/meta/models/:model_id/references", h.MetaData.ListReferences)

		v1.POST("/meta/models/:model_id/publish", h.MetaData.PublishModel)
		v1.GET("/meta/models/:model_id/versions", h.MetaData.ListVersions)
		v1.GET("/meta/models/:model_id/versions/:version", h.MetaData.GetVersion)
		v1.POST("/meta/models/:model_id/rollback", h.MetaData.RollbackModel)
		v1.GET("/meta/models/:model_id/records", h.MetaData.ListRecords)
		v1.POST("/meta/models/:model_id/records", h.MetaData.CreateRecord)
		v1.PUT("/meta/models/:model_id/records/:record_id", h.MetaData.UpdateRecord)
		v1.DELETE("/meta/models/:model_id/records/:record_id", h.MetaData.DeleteRecord)
		v1.GET("/meta/models/:model_id/records/template/export", h.MetaData.ExportRecordTemplate)
		v1.POST("/meta/models/:model_id/records/import", h.MetaData.ImportRecords)
		v1.GET("/meta/import-jobs/:job_id", h.MetaData.GetImportJob)
		v1.GET("/meta/import-jobs/:job_id/errors/export", h.MetaData.ExportImportJobErrorsCSV)

		v1.GET("/ops/decisions/self-repair", h.SelfRepair.ListSuggestions)
	}
	return r
}

package http

import (
	"computility-ops/backend/internal/http/handler"
	"computility-ops/backend/internal/http/middleware"
	"github.com/gin-gonic/gin"
)

type Handlers struct {
	Import        *handler.ImportHandler
	Renewal       *handler.RenewalHandler
	StorageDriver string
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

		v1.POST("/host-packages/import", h.Import.ImportHostPackages)
		v1.GET("/host-packages", h.Import.ListHostPackages)

		v1.POST("/special-rules/import", h.Import.ImportSpecialRules)
		v1.GET("/special-rules", h.Import.ListSpecialRules)

		v1.POST("/failure-rates/model/import", h.Import.ImportModelFailureRates)
		v1.GET("/failure-rates/model", h.Import.ListModelFailureRates)
		v1.POST("/failure-rates/package/import", h.Import.ImportPackageFailureRates)
		v1.GET("/failure-rates/package", h.Import.ListPackageFailureRates)
		v1.POST("/failure-rates/package-model/import", h.Import.ImportPackageModelFailureRates)
		v1.GET("/failure-rates/package-model", h.Import.ListPackageModelFailureRates)
		v1.GET("/failure-rates/overall", h.Import.ListOverallFailureRates)
		v1.POST("/failure-rates/analyze/import", h.Import.AnalyzeFaultRates)

		v1.POST("/renewals/plan", h.Renewal.CreatePlan)
		v1.GET("/renewals/plans", h.Renewal.ListPlans)
		v1.GET("/renewals/plans/:plan_id", h.Renewal.GetPlan)
		v1.DELETE("/renewals/plans/:plan_id", h.Renewal.DeletePlan)
		v1.GET("/renewals/plans/:plan_id/export", h.Renewal.ExportPlan)
	}
	return r
}

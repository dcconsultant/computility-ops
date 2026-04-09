package app

import (
	"fmt"

	"computility-ops/backend/internal/config"
	httpapi "computility-ops/backend/internal/http"
	"computility-ops/backend/internal/http/handler"
	"computility-ops/backend/internal/repository"
	"computility-ops/backend/internal/service"
	mem "computility-ops/backend/internal/storage/memory"
	mysql "computility-ops/backend/internal/storage/mysql"
	"github.com/gin-gonic/gin"
)

func Build(cfg config.Config) (*gin.Engine, error) {
	serverRepo, datasetRepo, renewalRepo, driver, err := buildRepos(cfg)
	if err != nil {
		return nil, err
	}

	importSvc := service.NewImportService(serverRepo, datasetRepo)
	renewalSvc := service.NewRenewalService(serverRepo, datasetRepo, renewalRepo)

	h := httpapi.Handlers{
		Import:        handler.NewImportHandler(importSvc),
		Renewal:       handler.NewRenewalHandler(renewalSvc),
		StorageDriver: driver,
	}
	return httpapi.NewRouter(h), nil
}

func buildRepos(cfg config.Config) (repository.ServerRepo, repository.DatasetRepo, repository.RenewalPlanRepo, string, error) {
	switch cfg.StorageDriver {
	case "memory", "":
		return mem.NewServerRepo(), mem.NewDatasetRepo(), mem.NewRenewalRepo(), "memory", nil
	case "mysql":
		if cfg.MySQLDSN == "" {
			return nil, nil, nil, "", fmt.Errorf("MYSQL_DSN is required when STORAGE_DRIVER=mysql")
		}
		return mysql.NewServerRepo(cfg.MySQLDSN), mysql.NewDatasetRepo(cfg.MySQLDSN), mysql.NewRenewalRepo(cfg.MySQLDSN), "mysql", nil
	default:
		return nil, nil, nil, "", fmt.Errorf("unsupported STORAGE_DRIVER: %s", cfg.StorageDriver)
	}
}

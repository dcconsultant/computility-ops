package app

import (
	"fmt"
	"os"

	"computility-ops/backend/internal/config"
	httpapi "computility-ops/backend/internal/http"
	"computility-ops/backend/internal/http/handler"
	rcapi "computility-ops/backend/internal/modules/reconfig-planning/api"
	rcapp "computility-ops/backend/internal/modules/reconfig-planning/application"
	rcinfra "computility-ops/backend/internal/modules/reconfig-planning/infrastructure"
	renewalapi "computility-ops/backend/internal/modules/renewal/api"
	renewalapp "computility-ops/backend/internal/modules/renewal/application"
	rpapi "computility-ops/backend/internal/modules/replacement-planning/api"
	rpapp "computility-ops/backend/internal/modules/replacement-planning/application"
	rpinfra "computility-ops/backend/internal/modules/replacement-planning/infrastructure"
	srapi "computility-ops/backend/internal/modules/self-repair/api"
	srapp "computility-ops/backend/internal/modules/self-repair/application"
	srinfra "computility-ops/backend/internal/modules/self-repair/infrastructure"
	"computility-ops/backend/internal/repository"
	"computility-ops/backend/internal/service"
	mem "computility-ops/backend/internal/storage/memory"
	mysql "computility-ops/backend/internal/storage/mysql"
	"github.com/gin-gonic/gin"
)

func Build(cfg config.Config) (*gin.Engine, error) {
	serverRepo, datasetRepo, renewalRepo, contractRepo, driver, err := buildRepos(cfg)
	if err != nil {
		return nil, err
	}

	importSvc := service.NewImportService(serverRepo, datasetRepo)
	renewalSvc := service.NewRenewalService(serverRepo, datasetRepo, renewalRepo)
	contractSvc := service.NewContractService(contractRepo)

	rules, err := rpinfra.LoadScoringRules(os.Getenv("REPLACEMENT_RULES_FILE"))
	if err != nil {
		return nil, fmt.Errorf("load replacement rules: %w", err)
	}
	replacementSvc := rpapp.NewServiceWithRules(rpinfra.NewLegacyReader(serverRepo, datasetRepo), rules)
	reconfigSvc := rcapp.NewService(rcinfra.NewStaticReader())
	selfRepairSvc := srapp.NewService(srinfra.NewStaticReader())
	renewalReadSvc := renewalapp.NewService(renewalRepo)

	h := httpapi.Handlers{
		Import:               handler.NewImportHandler(importSvc),
		Renewal:              handler.NewRenewalHandler(renewalSvc),
		Contract:             handler.NewContractHandler(contractSvc),
		System:               handler.NewSystemHandler(),
		StorageDriver:        driver,
		ReplacementPlanning:  rpapi.NewHandler(replacementSvc),
		ReconfigPlanning:     rcapi.NewHandler(reconfigSvc),
		SelfRepair:           srapi.NewHandler(selfRepairSvc),
		RenewalRead:          renewalapi.NewLegacyQueryAdapter(renewalReadSvc),
	}
	return httpapi.NewRouter(h), nil
}

func buildRepos(cfg config.Config) (repository.ServerRepo, repository.DatasetRepo, repository.RenewalPlanRepo, repository.ContractRepo, string, error) {
	switch cfg.StorageDriver {
	case "memory", "":
		return mem.NewServerRepo(), mem.NewDatasetRepo(), mem.NewRenewalRepo(), mem.NewContractRepo(), "memory", nil
	case "mysql":
		if cfg.MySQLDSN == "" {
			return nil, nil, nil, nil, "", fmt.Errorf("MYSQL_DSN is required when STORAGE_DRIVER=mysql")
		}
		return mysql.NewServerRepo(cfg.MySQLDSN), mysql.NewDatasetRepo(cfg.MySQLDSN), mysql.NewRenewalRepo(cfg.MySQLDSN), mysql.NewContractRepo(cfg.MySQLDSN), "mysql", nil
	default:
		return nil, nil, nil, nil, "", fmt.Errorf("unsupported STORAGE_DRIVER: %s", cfg.StorageDriver)
	}
}

package application

import (
	"context"
	"strings"

	"computility-ops/backend/internal/domain"
	assetdomain "computility-ops/backend/internal/modules/asset-config/domain"
)

type ServerReader interface {
	List(ctx context.Context) ([]domain.Server, error)
}

type DatasetReader interface {
	ListHostPackages(ctx context.Context) ([]domain.HostPackageConfig, error)
	ListSpecialRules(ctx context.Context) ([]domain.SpecialRule, error)
}

type Service struct {
	servers  ServerReader
	datasets DatasetReader
}

func NewService(servers ServerReader, datasets DatasetReader) *Service {
	return &Service{servers: servers, datasets: datasets}
}

func (s *Service) ListServers(ctx context.Context) ([]assetdomain.Server, error) {
	rows, err := s.servers.List(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]assetdomain.Server, 0, len(rows))
	for _, row := range rows {
		out = append(out, assetdomain.Server{
			AssetID:        strings.TrimSpace(row.SN),
			SN:             row.SN,
			Manufacturer:   row.Manufacturer,
			Model:          row.Model,
			DetailedConfig: row.DetailedConfig,
			PSA:            row.PSA,
			IDC:            row.IDC,
			Environment:    row.Environment,
			ConfigType:     row.ConfigType,
			WarrantyEnd:    row.WarrantyEndDate,
			LaunchDate:     row.LaunchDate,
		})
	}
	return out, nil
}

func (s *Service) ListHostPackages(ctx context.Context) ([]assetdomain.HostPackage, error) {
	rows, err := s.datasets.ListHostPackages(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]assetdomain.HostPackage, 0, len(rows))
	for _, row := range rows {
		out = append(out, assetdomain.HostPackage{
			ConfigType:             row.ConfigType,
			SceneCategory:          row.SceneCategory,
			CPULogicalCores:        row.CPULogicalCores,
			GPUCardCount:           row.GPUCardCount,
			DataDiskType:           row.DataDiskType,
			DataDiskCount:          row.DataDiskCount,
			StorageCapacityTB:      row.StorageCapacityTB,
			ServerValueScore:       row.ServerValueScore,
			ArchStandardizedFactor: row.ArchStandardizedFactor,
		})
	}
	return out, nil
}

func (s *Service) ListSpecialRules(ctx context.Context) ([]assetdomain.SpecialRule, error) {
	rows, err := s.datasets.ListSpecialRules(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]assetdomain.SpecialRule, 0, len(rows))
	for _, row := range rows {
		out = append(out, assetdomain.SpecialRule{
			SN:      row.SN,
			Policy:  row.Policy,
			Reason:  row.Reason,
			PSA:     row.PSA,
			IDC:     row.IDC,
			Env:     row.IDC,
			Model:   row.Model,
			Vendor:  row.Manufacturer,
			Package: row.PackageType,
		})
	}
	return out, nil
}

package api

import (
	"computility-ops/backend/internal/domain"
	masterdomain "computility-ops/backend/internal/modules/master-data/domain"
)

// LegacyServerToAsset keeps old domain structs compatible while migrating routes.
// Phase 1 placeholder: no route switch yet.
func LegacyServerToAsset(s domain.Server) masterdomain.Asset {
	return masterdomain.Asset{
		AssetID:      s.SN,
		SN:           s.SN,
		Manufacturer: s.Manufacturer,
		Model:        s.Model,
		ConfigType:   s.ConfigType,
		IDC:          s.IDC,
		Environment:  s.Environment,
	}
}

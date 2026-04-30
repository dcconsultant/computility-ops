package domain

// Asset is the normalized master-data view for server assets.
// It is intentionally minimal in Phase 1.
type Asset struct {
	AssetID      string `json:"asset_id"`
	SN           string `json:"sn"`
	Manufacturer string `json:"manufacturer,omitempty"`
	Model        string `json:"model,omitempty"`
	ConfigType   string `json:"config_type,omitempty"`
	IDC          string `json:"idc,omitempty"`
	Environment  string `json:"environment,omitempty"`
}

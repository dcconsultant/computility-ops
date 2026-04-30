package domain

type Server struct {
	AssetID        string `json:"asset_id"`
	SN             string `json:"sn"`
	Manufacturer   string `json:"manufacturer"`
	Model          string `json:"model"`
	DetailedConfig string `json:"detailed_config,omitempty"`
	PSA            string `json:"psa"`
	IDC            string `json:"idc,omitempty"`
	Environment    string `json:"environment,omitempty"`
	ConfigType     string `json:"config_type"`
	WarrantyEnd    string `json:"warranty_end_date,omitempty"`
	LaunchDate     string `json:"launch_date,omitempty"`
}

type HostPackage struct {
	ConfigType             string  `json:"config_type"`
	SceneCategory          string  `json:"scene_category,omitempty"`
	CPULogicalCores        int     `json:"cpu_logical_cores"`
	GPUCardCount           int     `json:"gpu_card_count,omitempty"`
	DataDiskType           string  `json:"data_disk_type,omitempty"`
	DataDiskCount          int     `json:"data_disk_count,omitempty"`
	StorageCapacityTB      float64 `json:"storage_capacity_tb,omitempty"`
	ServerValueScore       float64 `json:"server_value_score,omitempty"`
	ArchStandardizedFactor float64 `json:"arch_standardized_factor"`
}

type SpecialRule struct {
	SN      string `json:"sn"`
	Policy  string `json:"policy"`
	Reason  string `json:"reason,omitempty"`
	PSA     string `json:"psa,omitempty"`
	IDC     string `json:"idc,omitempty"`
	Env     string `json:"environment,omitempty"`
	Model   string `json:"model,omitempty"`
	Vendor  string `json:"manufacturer,omitempty"`
	Package string `json:"package_type,omitempty"`
}

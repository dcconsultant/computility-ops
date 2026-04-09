package domain

type Server struct {
	SN              string  `json:"sn"`
	Manufacturer    string  `json:"manufacturer"`
	Model           string  `json:"model"`
	PSA             float64 `json:"psa"`
	IDC             string  `json:"idc,omitempty"`
	Environment     string  `json:"environment,omitempty"`
	ConfigType      string  `json:"config_type"`
	WarrantyEndDate string  `json:"warranty_end_date,omitempty"`
	LaunchDate      string  `json:"launch_date,omitempty"`
}

type HostPackageConfig struct {
	ConfigType             string  `json:"config_type"`
	SceneCategory          string  `json:"scene_category,omitempty"`
	CPULogicalCores        int     `json:"cpu_logical_cores"`
	DataDiskType           string  `json:"data_disk_type,omitempty"`
	DataDiskCount          int     `json:"data_disk_count,omitempty"`
	StorageCapacityTB      float64 `json:"storage_capacity_tb,omitempty"`
	ArchStandardizedFactor float64 `json:"arch_standardized_factor"`
}

type SpecialRule struct {
	SN              string `json:"sn"`
	Manufacturer    string `json:"manufacturer,omitempty"`
	Model           string `json:"model,omitempty"`
	PSA             string `json:"psa,omitempty"`
	IDC             string `json:"idc,omitempty"`
	PackageType     string `json:"package_type,omitempty"`
	WarrantyEndDate string `json:"warranty_end_date,omitempty"`
	LaunchDate      string `json:"launch_date,omitempty"`
	Policy          string `json:"policy"` // whitelist | blacklist
}

type ModelFailureRate struct {
	Manufacturer string  `json:"manufacturer"`
	Model        string  `json:"model"`
	FailureRate  float64 `json:"failure_rate"`
}

type PackageFailureRate struct {
	ConfigType  string  `json:"config_type"`
	FailureRate float64 `json:"failure_rate"`
}

type PackageModelFailureRate struct {
	ConfigType   string  `json:"config_type"`
	Manufacturer string  `json:"manufacturer"`
	Model        string  `json:"model"`
	FailureRate  float64 `json:"failure_rate"`
}

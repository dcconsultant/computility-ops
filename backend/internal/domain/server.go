package domain

type Server struct {
	SN              string `json:"sn"`
	Manufacturer    string `json:"manufacturer"`
	Model           string `json:"model"`
	PSA             string `json:"psa"`
	IDC             string `json:"idc,omitempty"`
	Environment     string `json:"environment,omitempty"`
	ConfigType      string `json:"config_type"`
	WarrantyEndDate string `json:"warranty_end_date,omitempty"`
	LaunchDate      string `json:"launch_date,omitempty"`
}

type HostPackageConfig struct {
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
	Manufacturer            string  `json:"manufacturer"`
	Model                   string  `json:"model"`
	FailureRate             float64 `json:"failure_rate"`
	OverWarrantyFailureRate float64 `json:"over_warranty_failure_rate,omitempty"`
}

type PackageFailureRate struct {
	ConfigType              string  `json:"config_type"`
	FailureRate             float64 `json:"failure_rate"`
	OverWarrantyFailureRate float64 `json:"over_warranty_failure_rate,omitempty"`
}

type PackageModelFailureRate struct {
	ConfigType              string  `json:"config_type"`
	Manufacturer            string  `json:"manufacturer"`
	Model                   string  `json:"model"`
	FailureRate             float64 `json:"failure_rate"`
	OverWarrantyFailureRate float64 `json:"over_warranty_failure_rate,omitempty"`
}

type FailureRateSummary struct {
	Period               string  `json:"period"`
	Year                 int     `json:"year,omitempty"`
	Scope                string  `json:"scope"`
	Segment              string  `json:"segment"`
	FullCycleFailureRate float64 `json:"full_cycle_failure_rate"`
	OverWarrantyRate     float64 `json:"over_warranty_failure_rate"`
	FaultCount           int     `json:"fault_count"`
	OverWarrantyFaults   int     `json:"over_warranty_fault_count"`
	ServerYears          float64 `json:"server_years"`
	OverWarrantyYears    float64 `json:"over_warranty_years"`
}

type FailureOverviewCard struct {
	Segment                string  `json:"segment"`
	Year                   int     `json:"year"`
	CurrentYearFaultRate   float64 `json:"current_year_fault_rate"`
	HistoryAvgFaultRate    float64 `json:"history_avg_fault_rate"`
	CurrentYearFaultCount  int     `json:"current_year_fault_count"`
	CurrentYearDenominator float64 `json:"current_year_denominator"`
	HistoryFaultCount      int     `json:"history_fault_count"`
	HistoryDenominator     float64 `json:"history_denominator"`
}

type FailureAgeTrendPoint struct {
	Segment             string  `json:"segment"`
	AgeBucket           int     `json:"age_bucket"`
	NumeratorFaultCount int     `json:"numerator_fault_count"`
	DenominatorExposure float64 `json:"denominator_exposure"`
	FaultRate           float64 `json:"fault_rate"`
}

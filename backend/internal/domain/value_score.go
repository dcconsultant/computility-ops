package domain

type ValueScoreCabinetBaseline struct {
	Status             string  `json:"status"`
	IDC                string  `json:"idc"`
	CabinetUtilization float64 `json:"cabinet_utilization"`
	MinRatedPowerKW    float64 `json:"min_rated_power_kw"`
	MonthlyRentCNY     float64 `json:"monthly_rent_cny"`
	Formula            string  `json:"formula"`
	SourceCount        int     `json:"source_count"`
	Note               string  `json:"note,omitempty"`
}

type ValueScoreCostParams struct {
	DepreciationMonths    int     `json:"depreciation_months"`
	NetworkDeviceShareCNY float64 `json:"network_device_share_cny"`
	ServerRenewalFeeCNY   float64 `json:"server_renewal_fee_cny"`
	CabinetUtilization    float64 `json:"cabinet_utilization"`
	RatedPowerKW          float64 `json:"rated_power_kw"`
	MonthlyRentCNY        float64 `json:"monthly_rent_cny"`
}

type ValueScoreOriginalValue struct {
	ConfigType        string  `json:"config_type"`
	ServerOriginalCNY float64 `json:"server_original_cny"`
}

type ValueScoreTableItem struct {
	ConfigType              string  `json:"config_type"`
	SceneCategory           string  `json:"scene_category,omitempty"`
	SceneType               string  `json:"scene_type,omitempty"`
	CPULogicalCores         int     `json:"cpu_logical_cores"`
	MemoryCapacityGB        float64 `json:"memory_capacity_gb"`
	CapacityStorageTB       float64 `json:"capacity_storage_tb"`
	CountGPU                int     `json:"count_gpu"`
	UnavailableCores        int     `json:"unavailable_cores"`
	UnavailableMemoryGB     float64 `json:"unavailable_memory_gb"`
	PerformanceScore        float64 `json:"performance_score"`
	AvailableCores          int     `json:"available_cores"`
	AvailableMemoryGB       float64 `json:"available_memory_gb"`
	StandardScore           float64 `json:"standard_score"`
	CPUPerformanceFactor    float64 `json:"cpu_performance_factor"`
	MemoryRatio             float64 `json:"memory_ratio"`
	MemoryRatioFactor       float64 `json:"memory_ratio_factor"`
	OverallPerformanceRatio float64 `json:"overall_performance_ratio"`
	PowerWatts              float64 `json:"power_watts"`
	PowerKW                 float64 `json:"power_kw"`
	CabinetCostMonthly      float64 `json:"cabinet_cost_monthly"`
	ServerOriginalCNY       float64 `json:"server_original_cny"`
	DepreciationMonthly     float64 `json:"depreciation_monthly"`
	NetworkDeviceMonthly    float64 `json:"network_device_monthly"`
	NetworkCabinetMonthly   float64 `json:"network_cabinet_monthly"`
	ServerRenewalMonthly    float64 `json:"server_renewal_monthly"`
	OtherFixedCostMonthly   float64 `json:"other_fixed_cost_monthly"`
	TotalTCOMonthly         float64 `json:"total_tco_monthly"`
	UnitTCO                 float64 `json:"unit_tco"`
	ValueScoreV1            float64 `json:"value_score_v1"`
}

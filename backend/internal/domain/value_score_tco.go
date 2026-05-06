package domain

type ValueScoreTCOCalculateRequest struct {
	ConfigTypes []string `json:"config_types,omitempty"`
}

type ValueScoreTCOItem struct {
	ConfigType             string  `json:"config_type"`
	PowerWatts             float64 `json:"power_watts"`
	PowerKW                float64 `json:"power_kw"`
	CabinetCostMonthly     float64 `json:"cabinet_cost_monthly"`
	DepreciationMonthly    float64 `json:"depreciation_monthly"`
	NetworkDeviceMonthly   float64 `json:"network_device_monthly"`
	NetworkCabinetMonthly  float64 `json:"network_cabinet_monthly"`
	ServerRenewalMonthly   float64 `json:"server_renewal_monthly"`
	OtherFixedCostMonthly  float64 `json:"other_fixed_cost_monthly"`
	TotalTCOMonthly        float64 `json:"total_tco_monthly"`
}

type ValueScoreTCOCalculateResult struct {
	Status             string              `json:"status"`
	IDC                string              `json:"idc"`
	CabinetUtilization float64             `json:"cabinet_utilization"`
	MinRatedPowerKW    float64             `json:"min_rated_power_kw"`
	MonthlyRentCNY     float64             `json:"monthly_rent_cny"`
	DepreciationMonths int                 `json:"depreciation_months"`
	Formula            string              `json:"formula"`
	Items              []ValueScoreTCOItem `json:"items"`
	Note               string              `json:"note,omitempty"`
}

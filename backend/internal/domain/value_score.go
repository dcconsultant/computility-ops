package domain

type ValueScoreCostSettings struct {
	ElectricityPriceCNYPerKWh float64 `json:"electricity_price_cny_per_kwh"`
	DepreciationMonths        int     `json:"depreciation_months"`
	CabinetUtilization        float64 `json:"cabinet_utilization"`
}

type PackageCabinetCheckItem struct {
	ConfigType   string  `json:"config_type"`
	IDC          string  `json:"idc"`
	PowerWatts   float64 `json:"power_watts"`
	PowerKW      float64 `json:"power_kw"`
	Matched      bool    `json:"matched"`
	Reason       string  `json:"reason,omitempty"`
}

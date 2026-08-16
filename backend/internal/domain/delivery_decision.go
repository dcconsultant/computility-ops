package domain

type DeliveryDecisionInput struct {
	HardwareTotalCNY       float64 `json:"hw_total"`
	HardwareCores          float64 `json:"hw_cores"`
	HardwareTaxRate        float64 `json:"hw_tax_rate"`
	IDCRentMonthly         float64 `json:"idc_rent_monthly"`
	IDCRackKW              float64 `json:"idc_rack_kw"`
	IDCFillRate            float64 `json:"idc_fill_rate"`
	IDCServerPowerW        float64 `json:"idc_server_power_w"`
	IDCNetworkDepreciation float64 `json:"idc_network_depreciation"`
	CloudMemoryRatio       float64 `json:"cloud_memory_ratio"`
	CloudDiskRatio         float64 `json:"cloud_disk_ratio"`
	CloudCPUDailyPrice     float64 `json:"cloud_cpu_daily_price"`
	CloudMemoryDailyPrice  float64 `json:"cloud_memory_daily_price"`
	CloudDiskDailyPrice    float64 `json:"cloud_disk_daily_price"`
	CloudTaxRate           float64 `json:"cloud_tax_rate"`
	DepreciationYears      int     `json:"depreciation_years"`
	WACCRate               float64 `json:"wacc_rate"`
	ResidualRate           float64 `json:"residual_rate"`
	Country                string  `json:"country"`
	Currency               string  `json:"currency"`
	CloudCurrentDiscount   float64 `json:"cloud_current_discount"`
}

type DeliveryDecisionCalculateRequest struct {
	Input DeliveryDecisionInput `json:"input"`
}

type DeliveryDecisionDefaults struct {
	Country  string                `json:"country"`
	Currency string                `json:"currency"`
	Input    DeliveryDecisionInput `json:"input"`
}

type DeliveryDecisionFormulaTrace struct {
	CloudGross          float64  `json:"cloud_gross"`
	CloudDailyNet       float64  `json:"cloud_daily_net"`
	HardwareNetPerCore  float64  `json:"hardware_net_per_core"`
	ServerKW            float64  `json:"server_kw"`
	PhysicalMonthlyNet  float64  `json:"physical_monthly_net"`
	PhysicalDailyNet    float64  `json:"physical_daily_net"`
	DailyDepreciation   float64  `json:"daily_depreciation"`
	DailyWACC           float64  `json:"daily_wacc"`
	DailyDepreciation3Y float64  `json:"daily_depreciation_3y"`
	SelfDailyTCO        float64  `json:"self_daily_tco"`
	SelfDailyTCO3Y      float64  `json:"self_daily_tco_3y"`
	PremiumRatioR       float64  `json:"premium_ratio_r"`
	SelfWeight          float64  `json:"self_weight"`
	CloudWeight         float64  `json:"cloud_weight"`
	FormulaSelfShare    float64  `json:"formula_self_share"`
	FormulaCloudShare   float64  `json:"formula_cloud_share"`
	FinalSelfShare      float64  `json:"final_self_share"`
	FinalCloudShare     float64  `json:"final_cloud_share"`
	DailyMargin         float64  `json:"daily_margin"`
	BreakEvenYears      *float64 `json:"break_even_years"`
	CloudHedgeLost      bool     `json:"cloud_hedge_lost"`
	EqualityTolerance   float64  `json:"equality_tolerance"`
}

type DeliveryDecisionSensitivityPoint struct {
	Curve            string  `json:"curve"`
	Label            string  `json:"label"`
	XValue           float64 `json:"x_value"`
	CloudDailyNet    float64 `json:"cloud_daily_net"`
	SelfDailyTCO     float64 `json:"self_daily_tco"`
	SelfDailyTCO3Y   float64 `json:"self_daily_tco_3y"`
	FormulaSelfShare float64 `json:"formula_self_share"`
	FinalSelfShare   float64 `json:"final_self_share"`
	CloudHedgeLost   bool    `json:"cloud_hedge_lost"`
}

type DeliveryDecisionSnapshot struct {
	Operator       string `json:"operator"`
	FormulaVersion string `json:"formula_version"`
	CalculatedAt   string `json:"calculated_at"`
}

type DeliveryDecisionResult struct {
	Input             DeliveryDecisionInput              `json:"input"`
	Formula           DeliveryDecisionFormulaTrace       `json:"formula"`
	SensitivityPoints []DeliveryDecisionSensitivityPoint `json:"sensitivity_points"`
	Snapshot          DeliveryDecisionSnapshot           `json:"snapshot"`
}

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

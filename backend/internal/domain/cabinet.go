package domain

// CabinetUtilizationSetting stores the global cabinet utilization value in decimal form.
// Example: 0.8 means 80%, 1.2 means 120%.
type CabinetUtilizationSetting struct {
	Utilization float64 `json:"utilization"`
}

// CabinetConfig maps cabinet price/capacity by IDC and rated power.
type CabinetConfig struct {
	ID          int64   `json:"id"`
	IDC         string  `json:"idc"`
	RatedPowerKW float64 `json:"rated_power_kw"`
	MonthlyRent float64 `json:"monthly_rent"`
	CreatedAt   string  `json:"created_at,omitempty"`
	UpdatedAt   string  `json:"updated_at,omitempty"`
}

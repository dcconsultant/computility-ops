package domain

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

type FailureFeatureFact struct {
	RecordYearIndex     int     `json:"record_year_index"`
	RecordYearStart     string  `json:"record_year_start"`
	RecordYearEnd       string  `json:"record_year_end"`
	Scope               string  `json:"scope"`
	SceneGroup          string  `json:"scene_group"`
	AgeBucket           int     `json:"age_bucket"`
	DenominatorWeighted float64 `json:"denominator_weighted"`
	FaultCount          int     `json:"fault_count"`
	FaultRate           float64 `json:"fault_rate"`
}

type StorageTopServerRate struct {
	SN                   string  `json:"sn"`
	Manufacturer         string  `json:"manufacturer,omitempty"`
	Model                string  `json:"model,omitempty"`
	ConfigType           string  `json:"config_type,omitempty"`
	Environment          string  `json:"environment,omitempty"`
	IDC                  string  `json:"idc,omitempty"`
	WarrantyEndDate      string  `json:"warranty_end_date,omitempty"`
	DataDiskCount        int     `json:"data_disk_count"`
	SingleDiskCapacityTB float64 `json:"single_disk_capacity_tb"`
	TotalCapacityTB      float64 `json:"total_capacity_tb"`
	FaultCount           int     `json:"fault_count"`
	Denominator          float64 `json:"denominator"`
	FaultRate            float64 `json:"fault_rate"`
}

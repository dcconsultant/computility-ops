package domain

type ValueScorePerformanceParam struct {
	ConfigType          string  `json:"config_type"`
	UnavailableCores    int     `json:"unavailable_cores"`
	UnavailableMemoryGB float64 `json:"unavailable_memory_gb"`
	PerformanceScore    float64 `json:"performance_score"`
}

type ValueScorePerformanceAlert struct {
	ConfigType   string `json:"config_type"`
	ErrorCode    string `json:"error_code"`
	Field        string `json:"field"`
	CurrentValue string `json:"current_value"`
	Suggestion   string `json:"suggestion"`
}

type ValueScorePerformanceCalcItem struct {
	ConfigType             string                        `json:"config_type"`
	CPULogicalCores        int                           `json:"cpu_logical_cores"`
	MemoryCapacityGB       float64                       `json:"memory_capacity_gb"`
	UnavailableCores       int                           `json:"unavailable_cores"`
	UnavailableMemoryGB    float64                       `json:"unavailable_memory_gb"`
	PerformanceScore       float64                       `json:"performance_score"`
	AvailableCores         int                           `json:"available_cores"`
	AvailableMemoryGB      float64                       `json:"available_memory_gb"`
	StandardScore          float64                       `json:"standard_score"`
	CPUPerformanceFactor   float64                       `json:"cpu_performance_factor"`
	MemoryRatio            float64                       `json:"memory_ratio"`
	MemoryRatioFactor      float64                       `json:"memory_ratio_factor"`
	OverallPerformanceRatio float64                      `json:"overall_performance_ratio"`
	Alerts                 []ValueScorePerformanceAlert   `json:"alerts,omitempty"`
}

type ValueScorePerformanceCalcResult struct {
	Items      []ValueScorePerformanceCalcItem `json:"items"`
	AlertCount int                             `json:"alert_count"`
	Note       string                          `json:"note,omitempty"`
}

package domain

type RenewalUnitPrice struct {
	Country       string  `json:"country"`
	SceneCategory string  `json:"scene_category"`
	UnitPrice     float64 `json:"unit_price"`
}

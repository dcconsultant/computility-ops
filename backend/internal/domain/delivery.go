package domain

type ArrivalPlan struct {
	PlanID               string `json:"plan_id"`
	Category             string `json:"category"`
	MaterialCode         string `json:"material_code"`
	MaterialName         string `json:"material_name"`
	Quantity             int    `json:"quantity"`
	ReceivingAddress     string `json:"receiving_address"`
	Supplier             string `json:"supplier"`
	OrderNo              string `json:"order_no"`
	AssetCodeRange       string `json:"asset_code_range"`
	EstimatedArrivalTime string `json:"estimated_arrival_time"`
	Remark               string `json:"remark,omitempty"`
	CreatedAt            string `json:"created_at,omitempty"`
	UpdatedAt            string `json:"updated_at,omitempty"`
}

type DeviceArrivalRecord struct {
	RecordID                   string  `json:"record_id"`
	Category                   string  `json:"category"`
	PackageCode                string  `json:"package_code"`
	PackageType                string  `json:"package_type"`
	MaterialServiceCode        string  `json:"material_service_code"`
	MaterialServiceDescription string  `json:"material_service_description"`
	RackUnits                  float64 `json:"rack_units"`
	Manufacturer               string  `json:"manufacturer"`
	Quantity                   int     `json:"quantity"`
	ReceivingLocation          string  `json:"receiving_location"`
	PurchaseRequestNo          string  `json:"purchase_request_no"`
	SRMRequirementSubmittedAt  string  `json:"srm_requirement_submitted_at"`
	PONo                       string  `json:"po_no"`
	ActualArrivalTime          string  `json:"actual_arrival_time"`
	CreatedAt                  string  `json:"created_at,omitempty"`
	UpdatedAt                  string  `json:"updated_at,omitempty"`
}

type AccessoryArrivalRecord struct {
	RecordID                   string `json:"record_id"`
	PurchaseRequestNo          string `json:"purchase_request_no"`
	MaterialServiceCode        string `json:"material_service_code"`
	MaterialServiceDescription string `json:"material_service_description"`
	Quantity                   int    `json:"quantity"`
	Supplier                   string `json:"supplier"`
	IDCRoom                    string `json:"idc_room"`
	PurchaseBackground         string `json:"purchase_background"`
	SRMRequirementSubmittedAt  string `json:"srm_requirement_submitted_at"`
	PONo                       string `json:"po_no"`
	ArrivalTime                string `json:"arrival_time"`
	CreatedAt                  string `json:"created_at,omitempty"`
	UpdatedAt                  string `json:"updated_at,omitempty"`
}

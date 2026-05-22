package domain

type Supplier struct {
	SupplierID        string `json:"supplier_id"`
	CompanyFullName   string `json:"company_full_name"`
	TaxNumber         string `json:"tax_number"`
	ProjectOwner      string `json:"project_owner"`
	ProjectOwnerPhone string `json:"project_owner_phone"`
	TechContact       string `json:"tech_contact"`
	TechContactPhone  string `json:"tech_contact_phone"`
	BusinessScope     string `json:"business_scope"`
	CreatedAt         string `json:"created_at,omitempty"`
	UpdatedAt         string `json:"updated_at,omitempty"`
}

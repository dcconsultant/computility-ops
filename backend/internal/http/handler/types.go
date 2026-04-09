package handler

type CreatePlanReq struct {
	TargetCores int `json:"target_cores" binding:"required,min=1"`
}

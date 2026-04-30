package api

import (
	rcapp "computility-ops/backend/internal/modules/reconfig-planning/application"
	"github.com/gin-gonic/gin"
)

type Handler struct{ svc *rcapp.Service }

func NewHandler(svc *rcapp.Service) *Handler { return &Handler{svc: svc} }

// Phase 1/2 placeholder: not mounted to router yet.
func (h *Handler) ListSuggestions(c *gin.Context) {
	rows, err := h.svc.ListSuggestions(c.Request.Context())
	if err != nil {
		c.JSON(200, gin.H{"code": 50001, "message": err.Error(), "data": gin.H{}})
		return
	}
	c.JSON(200, gin.H{"code": 0, "message": "ok", "data": gin.H{"list": rows, "total": len(rows)}})
}

package api

import (
	"net/http"

	masterapp "computility-ops/backend/internal/modules/master-data/application"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	svc *masterapp.Service
}

func NewHandler(svc *masterapp.Service) *Handler {
	return &Handler{svc: svc}
}

// ListAssets is intentionally not mounted in router during Phase 1.
// It is prepared for Phase 2 endpoint cutover.
func (h *Handler) ListAssets(c *gin.Context) {
	assets, err := h.svc.ListAssets(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 1, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "ok", "data": assets})
}

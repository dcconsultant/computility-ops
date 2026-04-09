package handler

import (
	"sort"
	"strconv"
	"strings"

	"computility-ops/backend/internal/domain"
	"computility-ops/backend/internal/repository"
	"github.com/gin-gonic/gin"
)

type ServerHandler struct {
	repo repository.ServerRepo
}

func NewServerHandler(repo repository.ServerRepo) *ServerHandler { return &ServerHandler{repo: repo} }

func (h *ServerHandler) List(c *gin.Context) {
	list, err := h.repo.List(c.Request.Context())
	if err != nil {
		fail(c, 50001, err.Error())
		return
	}
	keyword := strings.ToLower(strings.TrimSpace(c.Query("keyword")))
	if keyword != "" {
		filtered := make([]domain.Server, 0, len(list))
		for _, s := range list {
			if strings.Contains(strings.ToLower(s.SN), keyword) || strings.Contains(strings.ToLower(s.Model), keyword) {
				filtered = append(filtered, s)
			}
		}
		list = filtered
	}
	sort.Slice(list, func(i, j int) bool { return list[i].PSA > list[j].PSA })

	page := atoiDefault(c.DefaultQuery("page", "1"), 1)
	if page < 1 {
		page = 1
	}
	pageSize := atoiDefault(c.DefaultQuery("page_size", "20"), 20)
	if pageSize < 1 {
		pageSize = 20
	}
	if pageSize > 200 {
		pageSize = 200
	}
	total := len(list)
	start := (page - 1) * pageSize
	if start > total {
		start = total
	}
	end := start + pageSize
	if end > total {
		end = total
	}
	ok(c, gin.H{"list": list[start:end], "total": total, "page": page, "page_size": pageSize})
}

func atoiDefault(v string, d int) int {
	n, err := strconv.Atoi(v)
	if err != nil {
		return d
	}
	return n
}

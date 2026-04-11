package handler

import (
	"context"
	"database/sql"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	"computility-ops/backend/internal/diagnose"
	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
)

type SystemHandler struct{}

func NewSystemHandler() *SystemHandler { return &SystemHandler{} }

type mysqlTestRequest struct {
	DSN      string `json:"dsn"`
	Host     string `json:"host"`
	Port     int    `json:"port"`
	User     string `json:"user"`
	Password string `json:"password"`
	Database string `json:"database"`
	Params   string `json:"params"`
}

func (h *SystemHandler) ListImportErrors(c *gin.Context) {
	c.Set("audit_action", "system.import_errors.list")
	limit := 20
	if q := strings.TrimSpace(c.Query("limit")); q != "" {
		if v, err := strconv.Atoi(q); err == nil {
			limit = v
		}
	}
	rows := diagnose.ListImportErrors(limit)
	ok(c, gin.H{"list": rows, "total": len(rows), "page": 1, "page_size": len(rows)})
}

func (h *SystemHandler) TestMySQLConnection(c *gin.Context) {
	c.Set("audit_action", "system.mysql.test")
	var req mysqlTestRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, 40001, "请求参数无效")
		return
	}
	dsn, err := buildMySQLDSN(req)
	if err != nil {
		fail(c, 40004, err.Error())
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	start := time.Now()
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		fail(c, 50001, fmt.Sprintf("连接失败: %v", err))
		return
	}
	defer db.Close()
	if err := db.PingContext(ctx); err != nil {
		fail(c, 50001, fmt.Sprintf("连接失败: %v", err))
		return
	}

	ok(c, gin.H{
		"reachable":  true,
		"latency_ms": time.Since(start).Milliseconds(),
		"message":    "MySQL连接成功",
	})
}

func buildMySQLDSN(req mysqlTestRequest) (string, error) {
	if strings.TrimSpace(req.DSN) != "" {
		return strings.TrimSpace(req.DSN), nil
	}
	host := strings.TrimSpace(req.Host)
	if host == "" {
		host = "127.0.0.1"
	}
	port := req.Port
	if port <= 0 {
		port = 3306
	}
	user := strings.TrimSpace(req.User)
	if user == "" {
		return "", fmt.Errorf("user 不能为空")
	}
	database := strings.TrimSpace(req.Database)
	if database == "" {
		return "", fmt.Errorf("database 不能为空")
	}
	params := strings.TrimSpace(req.Params)
	if params == "" {
		params = "parseTime=true&loc=Local&charset=utf8mb4"
	}
	password := url.QueryEscape(req.Password)
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?%s", user, password, host, port, database, params), nil
}

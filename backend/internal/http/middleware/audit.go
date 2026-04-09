package middleware

import (
	"encoding/json"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

type AuditRecord struct {
	Timestamp string `json:"ts"`
	RequestID string `json:"request_id"`
	Operator  string `json:"operator"`
	Method    string `json:"method"`
	Path      string `json:"path"`
	Status    int    `json:"status"`
	LatencyMs int64  `json:"latency_ms"`
	ClientIP  string `json:"client_ip"`
	UserAgent string `json:"user_agent"`
	Action    string `json:"action,omitempty"`
	Result    string `json:"result,omitempty"`
	Detail    string `json:"detail,omitempty"`
}

func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		reqID := c.GetHeader("X-Request-ID")
		if reqID == "" {
			reqID = strconv.FormatInt(time.Now().UnixNano(), 36) + "-" + strconv.Itoa(rand.Intn(10000))
		}
		c.Set("request_id", reqID)
		c.Writer.Header().Set("X-Request-ID", reqID)
		c.Next()
	}
}

func Audit() gin.HandlerFunc {
	logger := newAuditLogger()
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()

		requestID, _ := c.Get("request_id")
		action, _ := c.Get("audit_action")
		result, _ := c.Get("audit_result")
		detail, _ := c.Get("audit_detail")

		record := AuditRecord{
			Timestamp: time.Now().Format(time.RFC3339),
			RequestID: toString(requestID),
			Operator:  c.GetHeader("X-Operator"),
			Method:    c.Request.Method,
			Path:      c.FullPath(),
			Status:    c.Writer.Status(),
			LatencyMs: time.Since(start).Milliseconds(),
			ClientIP:  c.ClientIP(),
			UserAgent: c.Request.UserAgent(),
			Action:    toString(action),
			Result:    toString(result),
			Detail:    toString(detail),
		}
		if record.Operator == "" {
			record.Operator = "anonymous"
		}

		b, err := json.Marshal(record)
		if err != nil {
			log.Printf("audit marshal error: %v", err)
			return
		}
		logger.Println(string(b))
	}
}

func newAuditLogger() *log.Logger {
	path := os.Getenv("AUDIT_LOG_PATH")
	if path == "" {
		return log.New(os.Stdout, "audit ", 0)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		log.Printf("audit mkdir error: %v", err)
		return log.New(os.Stdout, "audit ", 0)
	}
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		log.Printf("audit open error: %v", err)
		return log.New(os.Stdout, "audit ", 0)
	}
	return log.New(f, "", 0)
}

func toString(v any) string {
	s, _ := v.(string)
	return s
}

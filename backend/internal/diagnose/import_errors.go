package diagnose

import (
	"strings"
	"sync"
	"time"
)

type ImportErrorItem struct {
	Time      string `json:"time"`
	RequestID string `json:"request_id"`
	Action    string `json:"action"`
	Reason    string `json:"reason"`
	Hint      string `json:"hint"`
}

var importErrors = struct {
	sync.Mutex
	items []ImportErrorItem
}{}

const maxItems = 100

func RecordImportError(action, requestID string, err error) {
	if err == nil {
		return
	}
	reason := strings.TrimSpace(err.Error())
	item := ImportErrorItem{
		Time:      time.Now().Format(time.RFC3339),
		RequestID: requestID,
		Action:    action,
		Reason:    reason,
		Hint:      AnalyzeReason(reason),
	}
	importErrors.Lock()
	defer importErrors.Unlock()
	importErrors.items = append([]ImportErrorItem{item}, importErrors.items...)
	if len(importErrors.items) > maxItems {
		importErrors.items = importErrors.items[:maxItems]
	}
}

func ListImportErrors(limit int) []ImportErrorItem {
	if limit <= 0 {
		limit = 20
	}
	if limit > maxItems {
		limit = maxItems
	}
	importErrors.Lock()
	defer importErrors.Unlock()
	if len(importErrors.items) < limit {
		limit = len(importErrors.items)
	}
	out := make([]ImportErrorItem, limit)
	copy(out, importErrors.items[:limit])
	return out
}

func AnalyzeReason(reason string) string {
	r := strings.ToLower(reason)
	switch {
	case strings.Contains(r, "doesn't exist") || strings.Contains(r, "does not exist"):
		return "疑似缺表或表名不一致，请执行 mysql_v3_ops_repo_tables.sql（以及 v2）迁移后重试"
	case strings.Contains(r, "access denied") || strings.Contains(r, "permission"):
		return "数据库账号权限不足，请给 computility_ops 库授予 SELECT/INSERT/UPDATE/DELETE/CREATE/ALTER 权限"
	case strings.Contains(r, "unknown column"):
		return "表结构版本落后（缺列），请重新执行最新迁移脚本"
	case strings.Contains(r, "duplicate"):
		return "数据存在重复键冲突，请检查唯一键字段（如 SN）是否重复"
	case strings.Contains(r, "data too long"):
		return "字段长度超限，请检查导入文件中的异常长文本"
	case strings.Contains(r, "invalid") && strings.Contains(r, "date"):
		return "日期格式不符合要求，请使用 yyyy-mm-dd 或系统支持的日期格式"
	default:
		return "请根据请求ID在后端日志定位详情；若为 MySQL 模式，优先检查表结构和权限"
	}
}

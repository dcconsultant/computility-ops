package handler

import "github.com/gin-gonic/gin"

func ok(c *gin.Context, data any) {
	c.Set("audit_result", "ok")
	c.JSON(200, gin.H{"code": 0, "message": "ok", "data": data})
}

func fail(c *gin.Context, code int, msg string) {
	requestID, _ := c.Get("request_id")
	c.Set("audit_result", "failed")
	c.Set("audit_detail", msg)
	c.JSON(200, gin.H{
		"code":    code,
		"message": msg,
		"data": gin.H{
			"request_id": requestID,
		},
	})
}

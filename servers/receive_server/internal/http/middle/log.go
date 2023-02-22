package middle

import (
	"fmt"
	"github.com/gin-gonic/gin"
	log "lalala_im/pkg/la_log"
	"net/http"
	"time"
)

// LogMiddle 日志拦截中间件
func LogMiddle() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		latency := time.Now().Sub(start)
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery
		clientIP := c.GetHeader("X-Forwarded-For")
		if len(clientIP) == 0 {
			clientIP = c.ClientIP()
		}
		if raw != "" {
			path = path + "?" + raw
		}
		method := c.Request.Method
		statusCode := c.Writer.Status()
		if method != http.MethodHead {
			log.Info(fmt.Sprintf("|STATUS: %d	|Latency: %v	|Client ip: %s	|method: %s	|path: %s	",
				statusCode,
				latency,
				clientIP,
				method,
				path))
		}

	}

}

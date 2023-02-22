package middle

import (
	"fmt"
	"github.com/gin-gonic/gin"
	log "lalala_im/pkg/la_log"
	"lalala_im/pkg/la_rsp"
	"net/http"
	"runtime/debug"
)

func Cors() gin.HandlerFunc {
	return func(c *gin.Context) {
		method := c.Request.Method
		origin := c.Request.Header.Get("Origin") //请求头部
		if origin != "" {
			//接收客户端发送的origin （重要！）
			c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
			//服务器支持的所有跨域请求的方法
			c.Header("Access-Control-Allow-Methods", "*")
			//允许跨域设置可以返回其他子段，可以自定义字段
			c.Header("Access-Control-Allow-Headers", "Authorization, Content-Length, X-CSRF-Token, Token,session,Content-Type,User-Agent,Referer,Accept")
			// 允许浏览器（客户端）可以解析的头部 （重要）
			c.Header("Access-Control-Expose-Headers", "Content-Length, Access-Control-Allow-Origin, Access-Control-Allow-Headers")
			//设置缓存时间
			c.Header("Access-Control-Max-Age", "172800")
			//允许客户端传递校验信息比如 cookie (重要)
			c.Header("Access-Control-Allow-Credentials", "true")
		}

		//允许类型校验
		if method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
		}

		//捕获panic输出到日志
		defer func() {
			if err := recover(); err != nil {
				log.Error(string(debug.Stack()), fmt.Sprintf("Panic info is: %v", err))
				la_rsp.Failed(c, la_rsp.ERROR, err.(error))
			}
		}()

		c.Next()
	}
}

package route

import (
	"github.com/gin-gonic/gin"
	"lalala_im/servers/pool_server/internal/http/middle"
	"lalala_im/servers/pool_server/internal/http/route/ws"
)

func InitRouters() *gin.Engine {
	r := gin.New()
	//Recovery
	r.Use(gin.Recovery())
	//跨域和补货panic
	r.Use(middle.Cors())
	//访问日志打印
	r.Use(middle.LogMiddle())

	//sayHai
	r.HEAD("/", sayHai)
	r.GET("/", sayHai)

	g := r.Group("")
	//ws服务
	ws.InitRouterWithOutAuth(g)

	//用户认证
	//g.Use(middle.AdminAuth())

	return r
}

func sayHai(ctx *gin.Context) {
	_, err := ctx.Writer.Write([]byte("hello,la→la↑la↓!"))
	if err != nil {
		return
	}
	return
}

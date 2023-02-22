package route

import (
	"github.com/gin-gonic/gin"
	"lalala_im/servers/receive_server/internal/http/middle"
	"lalala_im/servers/receive_server/internal/http/route/add_req"
	"lalala_im/servers/receive_server/internal/http/route/chat_log"
	"lalala_im/servers/receive_server/internal/http/route/friend"
	"lalala_im/servers/receive_server/internal/http/route/group"
	"lalala_im/servers/receive_server/internal/http/route/services"
	"lalala_im/servers/receive_server/internal/http/route/user"
	"lalala_im/servers/receive_server/internal/http/route/verify_code"
)

func InitRouters() *gin.Engine {
	r := gin.New()
	//Recovery
	r.Use(gin.Recovery())
	//跨域和捕获panic输出panic错误日志
	r.Use(middle.Cors())
	//访问日志打印
	r.Use(middle.LogMiddle())

	//sayHai
	r.HEAD("/", sayHai)
	r.GET("/", sayHai)

	g := r.Group("")
	//用户接口
	user.InitRouterWithoutAuth(g)
	verify_code.InitRouterWithoutAuth(g)
	chat_log.InitRouterWithoutAuth(g)

	//用户认证
	g.Use(middle.AdminAuth())

	user.InitRouter(g)
	friend.InitRouter(g)
	group.InitRouter(g)
	services.InitRouter(g)
	chat_log.InitRouter(g)
	add_req.InitRouter(g)

	return r
}

func sayHai(ctx *gin.Context) {
	_, err := ctx.Writer.Write([]byte("hello,la→la↑la↓!"))
	if err != nil {
		return
	}
	return
}

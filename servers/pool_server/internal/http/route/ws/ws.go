package ws

import (
	"github.com/gin-gonic/gin"
	"lalala_im/servers/pool_server/internal/conf"
	"lalala_im/servers/pool_server/internal/controller/http/ws"
)

func InitRouterWithOutAuth(g *gin.RouterGroup) {
	//初始化ws服务
	ws.Init(conf.Bootstrap.Server.Http.OutAddr, conf.Bootstrap.Grpc.IP+":"+conf.Bootstrap.Grpc.Port, conf.Bootstrap.Ws.WebsocketMaxConnNum, conf.Bootstrap.Ws.WebsocketTimeOut, conf.Bootstrap.Ws.WebsocketMaxMsgLen)
	userG := g.Group("/ws")
	userG.GET("", ws.WS.WsHandler)
}

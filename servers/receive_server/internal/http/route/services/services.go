package services

import (
	"github.com/gin-gonic/gin"
	"lalala_im/servers/receive_server/internal/controller/services"
)

func InitRouterWithoutAuth(g *gin.RouterGroup) {
}

func InitRouter(g *gin.RouterGroup) {
	sG := g.Group("/services")
	//获取websocket链接池服务IP地址列表
	sG.GET("/ws/list", services.GetPoolServiceAddrList)
}

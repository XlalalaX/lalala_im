package add_req

import (
	"github.com/gin-gonic/gin"
	"lalala_im/servers/receive_server/internal/controller/add_req"
)

func InitRouterWithoutAuth(g *gin.RouterGroup) {
}

func InitRouter(g *gin.RouterGroup) {
	addG := g.Group("/add_req")
	//获取自己的好友申请列表
	addG.GET("/self", add_req.GetSelfAddReqList)
	//获取群的申请列表
	addG.GET("/group", add_req.GetGroupAddReqList)
	//发送好友申请
	addG.POST("/send", add_req.SendAddReq)
	//发送入群申请
	addG.POST("/send_group", add_req.SendGroupAddReq)
	//处理申请
	addG.POST("/handle", add_req.HandleAddReq)
}

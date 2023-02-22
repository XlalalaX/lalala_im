package friend

import (
	"github.com/gin-gonic/gin"
	"lalala_im/servers/receive_server/internal/controller/friend"
)

func InitRouterWithoutAuth(g *gin.RouterGroup) {
}

func InitRouter(g *gin.RouterGroup) {
	fG := g.Group("/friend")
	//添加好友
	fG.POST("/add", friend.AddFriend)
	//删除好友
	fG.DELETE("/del", friend.DelFriend)
	//判断是否是好友
	fG.POST("/is", friend.IsFriend)
	//获取好友列表
	fG.GET("/list", friend.GetFriendList)
}

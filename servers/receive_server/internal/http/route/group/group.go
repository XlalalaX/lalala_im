package group

import (
	"github.com/gin-gonic/gin"
	"lalala_im/servers/receive_server/internal/controller/group"
)

func InitRouterWithoutAuth(g *gin.RouterGroup) {
}

func InitRouter(g *gin.RouterGroup) {
	groupG := g.Group("/group")
	//创建群
	groupG.POST("/create", group.CreatGroup)
	//获取群信息
	groupG.GET("/info", group.GetGroupInfo)
	//搜索群
	groupG.GET("/find_group_list", group.GetGroupList)
	//添加管理员
	groupG.POST("/admin/add", group.AddAdmin)
	//删除管理员
	groupG.DELETE("/admin/del", group.DelAdmin)
	//添加群成员
	groupG.POST("/member/add", group.AddMember)
	//删除群成员
	groupG.DELETE("/member/del", group.DelMember)
	//自己退群
	groupG.DELETE("/self/remove", group.RemoveSelf)
	//获取自己已经加入的所有群
	groupG.GET("self/list", group.GetSelfGroupList)
	//禁言群成员
	groupG.POST("/ban/add", group.AddBan)
	//解除禁言
	groupG.DELETE("/ban/del", group.DelBan)

}

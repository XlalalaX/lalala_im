package user

import (
	"github.com/gin-gonic/gin"
	"lalala_im/servers/receive_server/internal/controller/user"
)

func InitRouterWithoutAuth(g *gin.RouterGroup) {
	userG := g.Group("/user_not")
	//注册
	userG.POST("/register", user.Register)
	//登陆
	userG.POST("/login", user.Login)
	//用户头像
	userG.GET("/user_face_url", user.GetFaceUrl)
	//找回密码
	userG.POST("/user_find_back_password", user.FindBackPassWord)
}

func InitRouter(g *gin.RouterGroup) {
	userG := g.Group("/user")
	//userG.GET("/test", user.LoginTest)
	//用户信息
	userG.GET("/user_info", user.GetUserInfo)
	//用户展示信息
	userG.GET("/user_show_info", user.GetUserShowInfo)
	//修改用户信息
	userG.POST("/change_data", user.ChangeUserData)
	//修改密码
	userG.POST("/change_pw", user.ChangePassWord)
	//修改头像
	userG.POST("/user_face_url", user.UpdateFaceUrl)
	//查找用户
	userG.GET("/find_user_list", user.GetUserList)
}

package verify_code

import (
	"github.com/gin-gonic/gin"
	"lalala_im/servers/receive_server/internal/controller/verify_code"
)

func InitRouterWithoutAuth(g *gin.RouterGroup) {
	//获取邮箱验证码
	g.POST("/verify_code", verify_code.GetVerifyCode)
}

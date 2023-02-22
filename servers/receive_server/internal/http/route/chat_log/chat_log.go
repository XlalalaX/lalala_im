package chat_log

import (
	"github.com/gin-gonic/gin"
	"lalala_im/servers/receive_server/internal/controller/chat_log"
)

func InitRouterWithoutAuth(g *gin.RouterGroup) {
	chatG := g.Group("/chat_log")
	//获取文件临时链接
	chatG.GET("/file", chat_log.GetChatFileUrl)
}

func InitRouter(g *gin.RouterGroup) {
	chatG := g.Group("/chat_log")
	//拉取临时聊天记录
	chatG.GET("/temp", chat_log.GetRecentChatLogList)
	//拉取保存的聊天记录（所有的聊天记录，分页返回）
	chatG.GET("/save", chat_log.GetSaveLogList)
	//上传聊天文件
	chatG.POST("/file", chat_log.UploadChatFile)
}

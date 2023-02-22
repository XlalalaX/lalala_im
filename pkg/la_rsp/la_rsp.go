package la_rsp

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
)

type Response struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data"`
}

// HttpResponse  setting gin.JSON
func HttpResponse(ctx *gin.Context, httpCode, errCode int, data interface{}) {
	switch data.(type) {
	case error:
		ctx.JSON(httpCode, Response{
			Code: errCode,
			Msg:  getMsg(errCode),
			Data: fmt.Sprint(data),
		})
	default:
		ctx.JSON(httpCode, Response{
			Code: errCode,
			Msg:  getMsg(errCode),
			Data: data,
		})
	}
}

func Success(ctx *gin.Context, data interface{}) {
	HttpResponse(ctx, http.StatusOK, 0, data)
	ctx.Abort()
}

func Failed(ctx *gin.Context, errCode int, err error) {
	HttpResponse(ctx, http.StatusOK, errCode, err)
	ctx.Abort()
}

func FailedWithData(ctx *gin.Context, errCode int, data interface{}) {
	HttpResponse(ctx, http.StatusOK, errCode, data)
	ctx.Abort()
}

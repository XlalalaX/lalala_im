package verify_code

import (
	"github.com/gin-gonic/gin"
	"lalala_im/pkg/la_email"
	log "lalala_im/pkg/la_log"
	"lalala_im/pkg/la_rsp"
	"lalala_im/servers/receive_server/internal/db"
	"math/rand"
	"time"
)

type getVerifyCode struct {
	Email string `json:"email"`
}

// GetVerifyCode 获取验证码接口
func GetVerifyCode(ctx *gin.Context) {
	req := &getVerifyCode{}
	err := ctx.ShouldBind(req)
	if err != nil || len(req.Email) < 5 {
		log.Error("获取验证码请求，参数校验错误：", err)
		la_rsp.Failed(ctx, la_rsp.ParameterValidationError, err)
		return
	}
	randCode := make([]byte, 0, 6)
	for i := 0; i < cap(randCode); i++ {
		randCode = append(randCode, '0'+byte(rand.Intn(9)))
	}
	err = la_email.SendMail([]string{req.Email}, "验证码", "验证码："+string(randCode)+"\n验证码有效时长：5分钟")
	if err != nil {
		log.Error("获取验证码请求，发送邮箱验证码失败：", err)
		la_rsp.Failed(ctx, la_rsp.ERROR, err)
		return
	}
	err = db.DB.Rdb.Set(ctx, "verify_code:"+req.Email, string(randCode), time.Minute*5).Err()
	if err != nil {
		log.Error("获取验证码请求，邮箱验证码写入redis失败：", err)
		la_rsp.Failed(ctx, la_rsp.ERROR, err)
		return
	}
	la_rsp.Success(ctx, map[string]interface{}{
		"verify_code": string(randCode),
	})
}

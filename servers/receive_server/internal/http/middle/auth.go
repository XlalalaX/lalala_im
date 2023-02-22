package middle

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/qiniu/qmgo"
	"lalala_im/data/Repo/User"
	"lalala_im/data/entity"
	"lalala_im/pkg/la_auth"
	log "lalala_im/pkg/la_log"
	"lalala_im/servers/receive_server/internal/db"
	"net/http"
	"time"
)

var (
	userRepo = User.NewIUserRepo(db.DB.MongoDB)
)

func AdminAuth() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		token := ctx.GetHeader("token")
		userClaims, err := la_auth.GetClaimFromToken(token)
		if userClaims == nil || err != nil {
			if err != nil {
				ctx.JSON(http.StatusInternalServerError, err)
			} else {
				ctx.JSON(http.StatusUnauthorized, err)
			}

			log.Error("用户鉴权，提取token信息错误")
			ctx.Abort()
			return
		}
		userJson, err := db.DB.Rdb.Get(ctx, "auth:"+userClaims.UID).Result()
		if err != nil && err != redis.Nil {
			ctx.JSON(http.StatusInternalServerError, err)
			log.Error("用户鉴权，redis取数据错误")
			ctx.Abort()
			return
		}
		user := &entity.User{}
		if userJson != "" {
			err = json.Unmarshal([]byte(userJson), user)
			if err != nil || user.Version != userClaims.Version {
				if err != nil {
					ctx.JSON(http.StatusInternalServerError, err)
					log.Error("用户鉴权，redis数据json解析错误")
				} else {
					ctx.JSON(http.StatusUnauthorized, err)
				}

				ctx.Abort()
				return
			}
		} else {
			userInfo, err := userRepo.GetUserInfoByUID(ctx, userClaims.UID)
			if err != nil && err != qmgo.ErrNoSuchDocuments {
				ctx.JSON(http.StatusInternalServerError, err)
				log.Error("用户鉴权，数据库查询错误")
				ctx.Abort()
				return
			}
			if userInfo == nil || userInfo.Id.IsZero() || userClaims.Version != userInfo.Version {
				ctx.JSON(http.StatusUnauthorized, err)
				log.Error("用户鉴权，数据库查询错误")
				ctx.Abort()
				return
			}
			userJsonBytes, err := json.Marshal(entity.TransformFromModel(userInfo))
			if err != nil {
				ctx.JSON(http.StatusInternalServerError, err)
				log.Error("用户鉴权，用户数据解析到json错误")
				ctx.Abort()
				return
			}

			err = db.DB.Rdb.Set(ctx, "auth:"+userInfo.UID, string(userJsonBytes), 3600000*time.Second).Err()
			if err != nil {
				ctx.JSON(http.StatusInternalServerError, err)
				log.Error("用户鉴权，用户信息存入redis缓存错误")
				ctx.Abort()
				return
			}
			user = entity.TransformFromModel(userInfo)
		}

		ctx.Set("user", user)
		ctx.Next()
	}
}

// GetUserFromCtx 从上下文中获取用户信息
func GetUserFromCtx(ctx *gin.Context) *entity.User {
	user, isexist := ctx.Get("user")
	if isexist {
		return user.(*entity.User)
	}
	return nil
}

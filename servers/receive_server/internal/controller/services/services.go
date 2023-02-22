package services

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	log "lalala_im/pkg/la_log"
	"lalala_im/pkg/la_rsp"
	"lalala_im/servers/receive_server/internal/db"
)

func GetPoolServiceAddrList(ctx *gin.Context) {
	keys, err := db.DB.Rdb.Keys(ctx, "ws_pool_out_addr:*").Result()
	if err != nil {
		log.Error(fmt.Sprintf("获取在线ws服务器列表失败，err:%+v", err))
		la_rsp.Failed(ctx, la_rsp.ERROR, err)
		return
	}
	for i := 0; i < len(keys); i++ {
		keys[i] = keys[i][17:]
	}
	//返回连接数在1w以下的ws服务器
	list, err := db.DB.Rdb.ZRange(ctx, "ws_pool_out_addr_list", 0, 10000).Result()
	m := map[string]bool{}
	for i := 0; i < len(keys); i++ {
		m[keys[i]] = true
	}
	liveList := []string{}
	//更新下线ws服务器，无视错误
	for i := 0; i < len(list); i++ {
		if !m[list[i]] {
			err = db.DB.Rdb.ZRem(ctx, "ws_pool_out_addr_list", redis.Z{Member: list[i]}).Err()
			if err != nil {
				log.Error(fmt.Sprintf("移除失活状态ws服务器失败，ip：%s，err:%+v", list[i], err))
			}
		} else {
			liveList = append(liveList, list[i])
		}
	}
	//返回ws服务器列表，已经按服务器连接数排序
	la_rsp.Success(ctx, map[string]interface{}{
		"ws_addr_list": liveList,
	})
}

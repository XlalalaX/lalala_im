package db

import (
	"context"
	"encoding/json"
	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/go-redis/redis/v8"
	"github.com/qiniu/qmgo"
	"lalala_im/data"
	log "lalala_im/pkg/la_log"
	"time"
)

var DB *ReceiveServerDB

type ReceiveServerDB struct {
	Rdb       *redis.Client
	MongoDB   *qmgo.Database
	AliyunOss *oss.Bucket
}

func InitDB(mongoUrl, mongoDB, redisAddr, redisPW string, ossEndpoint string, ossAccessKeyID string, ossAccessKeySecret string) {
	DB = &ReceiveServerDB{}
	DB.Rdb = data.NewRedis(redisAddr, redisPW)
	DB.MongoDB = data.NewMongoDB(mongoUrl, mongoDB)
	var err error
	DB.AliyunOss, err = data.NewAliyunOss(ossEndpoint, ossAccessKeyID, ossAccessKeySecret).Bucket("lalala-im")
	if err != nil {
		log.Panic("初始化阿里云OSS的Bucket失败：", err)
	}
}

// SetObjectToRedis 把结构体存入redis
func (db *ReceiveServerDB) SetObjectToRedis(ctx context.Context, key string, data interface{}, ex time.Duration) error {
	jsonStr, err := json.Marshal(data)
	if err != nil {
		return err
	}
	return db.Rdb.Set(ctx, key, jsonStr, ex).Err()
}

// GetObjectFromRedis 从redis获取结构体
func (db *ReceiveServerDB) GetObjectFromRedis(ctx context.Context, key string, out interface{}) error {
	jsonStr, err := db.Rdb.Get(ctx, key).Result()
	if err != nil {
		return err
	}
	return json.Unmarshal([]byte(jsonStr), out)
}

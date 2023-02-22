package data

import (
	"context"
	"github.com/Shopify/sarama"
	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	go_redis "github.com/go-redis/redis/v8"
	"github.com/qiniu/qmgo"
	log "lalala_im/pkg/la_log"
	"time"
)

func NewMongoDB(url string, db string) *qmgo.Database {
	MongoCli, err := qmgo.NewClient(context.Background(), &qmgo.Config{Uri: url})
	if err != nil {
		log.Panic("mongo初始化失败", err)
	}
	err = MongoCli.Ping(5)
	if err != nil {
		log.Panic("mongo初始化Ping失败", err)
	}

	MongoDB := MongoCli.Database(db)
	return MongoDB
}

func NewRedis(addr string, password string) *go_redis.Client {
	cli := go_redis.NewClient(&go_redis.Options{
		Addr:     addr,
		Password: password, // no password set
		DB:       0,        // use default DB
		PoolSize: 100,      // 连接池大小
	})
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	err := cli.Ping(ctx).Err()
	if err != nil {
		log.Panic("初始化redis失败", err)
	}
	return cli
}

func NewKafka(addrList []string) sarama.Client {
	// 配置Kafka连接信息
	config := sarama.NewConfig()
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Retry.Max = 5
	config.Producer.Return.Successes = true
	config.Producer.MaxMessageBytes = 1024 * 1024 * 1024 * 1024
	//config.Net.DialTimeout = time.Second * 10
	//config.Net.ReadTimeout = time.Second * 10
	//config.Net.WriteTimeout = time.Second * 10

	// 创建Kafka管理者
	admin, err := sarama.NewClient(addrList, config)
	if err != nil {
		log.Panic("初始化kafka失败", err)
	}
	return admin
}

func NewAliyunOss(endpoint string, accessKeyID string, accessKeySecret string) *oss.Client {
	client, err := oss.New(endpoint, accessKeyID, accessKeySecret)
	if err != nil {
		log.Panic("初始化阿里云OSS失败", err)
	}
	return client
}

//
//func NewKafkaConsumerGroup(addrList []string) sarama.ConsumerGroup {
//	// 配置Kafka连接信息
//	config := sarama.NewConfig()
//	config.Producer.Return.Successes = true
//	config.Consumer.Group.Rebalance.Strategy = sarama.BalanceStrategyRoundRobin
//	config.Version = sarama.V2_8_0_0
//	config.Net.DialTimeout = time.Second * 10
//	config.Net.ReadTimeout = time.Second * 10
//	config.Net.WriteTimeout = time.Second * 10
//	// 创建Consumer Group
//	groupName := "test-group"
//	consumer, err := sarama.NewConsumerGroup(addrList, groupName, config)
//	if err != nil {
//		log.Panic("初始化kafka_ConsumerGroup失败", err)
//	}
//	return consumer
//}
//
//func NewKafkaProducer(addrList []string) sarama.AsyncProducer {
//	// 配置Kafka连接信息
//	config := sarama.NewConfig()
//	config.Producer.Return.Successes = true
//	config.Consumer.Group.Rebalance.Strategy = sarama.BalanceStrategyRoundRobin
//	config.Version = sarama.V2_8_0_0
//	config.Net.DialTimeout = time.Second * 10
//	config.Net.ReadTimeout = time.Second * 10
//	config.Net.WriteTimeout = time.Second * 10
//	// 创建Consumer Group
//	groupName := "test-group"
//	consumer, err := sarama.NewAsyncProducer(addrList, groupName, config)
//	if err != nil {
//		log.Panic("初始化kafka_ConsumerGroup失败", err)
//	}
//	return consumer
//}

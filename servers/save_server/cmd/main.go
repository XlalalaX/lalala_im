package main

import (
	"context"
	log "lalala_im/pkg/la_log"
	"lalala_im/servers/save_server/internal/conf"
	"lalala_im/servers/save_server/internal/controller"
	"lalala_im/servers/save_server/internal/db"
	"os"
	"os/signal"
	"syscall"
)

func main() {

	// 等待中断信号来优雅停止服务器，设置的5秒延迟
	quit := make(chan os.Signal, 1)
	// kill （不带参数的）是默认发送 syscall.SIGTERM
	// kill -2 是 syscall.SIGINT
	// kill -9 是 syscall.SIGKILL，但是无法被捕获到，所以无需添加
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		if err := db.DB.KafkaConsumerGroup.Consume(context.Background(), conf.Bootstrap.Kafka.Topics, controller.NewMsgConsumer()); err != nil {
			log.Panic("消费消息持久化服务关闭，错误：", err)
		}
	}()
	log.Info("消费消息持久化服务开启,消费topic：", conf.Bootstrap.Kafka.Topics)
	<-quit
	log.Warn("关闭服务")

	//// ctx是用来通知服务器还有5秒的时间来结束当前正在处理的request
	//ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	//defer cancel()

	if err := db.DB.KafkaConsumerGroup.Close(); err != nil {
		log.Panic("强制关闭服务: ", err)
	}

	log.Info("服务退出")
}

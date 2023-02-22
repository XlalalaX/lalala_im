package main

import (
	"context"
	"fmt"
	"google.golang.org/grpc"
	log "lalala_im/pkg/la_log"
	"lalala_im/servers/pool_server/internal/conf"
	grpcRegister "lalala_im/servers/pool_server/internal/grpc"
	"lalala_im/servers/pool_server/internal/http/route"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

//ws测试url：ws://127.0.0.1:3002/ws?send_id=la1&token=lalala1
//groupId:5a1cc7a4-8f49-4847-ac0d-3c53c13297a1

func main() {
	router := route.InitRouters()
	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", conf.Bootstrap.Server.Http.Addr),
		Handler: router,
	}

	grpcListener, err := net.Listen("tcp", ":"+conf.Bootstrap.Grpc.Port)
	if err != nil {
		panic(err)
	}
	defer grpcListener.Close()

	var grpcOpts []grpc.ServerOption
	grpcOpts = append(grpcOpts, []grpc.ServerOption{
		//最大收发消息长度100MB
		grpc.MaxRecvMsgSize(1024 * 1024 * 100),
		grpc.MaxSendMsgSize(1024 * 1024 * 100),
		//暂时没啥加
	}...)

	grpcServer := grpc.NewServer(grpcOpts...)
	grpcRegister.RegisterServerList(grpcServer)

	// 等待中断信号来优雅停止服务器，设置的5秒延迟
	quit := make(chan os.Signal, 1)
	// kill （不带参数的）是默认发送 syscall.SIGTERM
	// kill -2 是 syscall.SIGINT
	// kill -9 是 syscall.SIGKILL，但是无法被捕获到，所以无需添加
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		if err := server.ListenAndServe(); err != nil {
			log.Panic("服务关闭，错误：", err)
		}
	}()
	go func() {
		if err := grpcServer.Serve(grpcListener); err != nil {
			log.Panic("grpc服务关闭，错误：", err)
		}
	}()
	log.Info("http服务监听端口：", server.Addr)
	log.Info("grpc服务监听端口：", grpcListener.Addr())
	<-quit
	log.Warn("关闭服务")

	// ctx是用来通知服务器还有5秒的时间来结束当前正在处理的request
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	grpcServer.GracefulStop()
	if err := server.Shutdown(ctx); err != nil {
		log.Panic("强制关闭服务: ", err)
	}

	log.Info("服务退出")
}

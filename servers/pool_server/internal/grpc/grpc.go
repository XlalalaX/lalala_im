package grpcRegister

import (
	"google.golang.org/grpc"
	"lalala_im/proto/pb_msg"
	"lalala_im/servers/pool_server/internal/controller/grpc/grpc_msg"
)

func RegisterServerList(s *grpc.Server) {
	pb_msg.RegisterMsgServerServer(s, &grpc_msg.MsgServer{})
}

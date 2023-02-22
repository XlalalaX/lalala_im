package grpc_msg

import (
	"context"
	"fmt"
	"github.com/gorilla/websocket"
	"google.golang.org/protobuf/proto"
	log "lalala_im/pkg/la_log"
	"lalala_im/proto/pb_msg"
	"lalala_im/servers/pool_server/internal/controller/http/ws"
)

var _ = pb_msg.MsgServerServer(&MsgServer{})

type MsgServer struct {
}

func (*MsgServer) Push(ctx context.Context, msg *pb_msg.Msg) (*pb_msg.ErrRsp, error) {
	var conn *ws.UserConn
	if msg.SelfID != "" {
		//历史消息用
		conn = ws.WS.GetUserConn(msg.SelfID)
		if conn == nil {
			return &pb_msg.ErrRsp{
				ErrCode: -1,
				ErrMsg:  "用户不存在",
			}, nil
		}
	} else {
		//实时推送用
		conn = ws.WS.GetUserConn(msg.RecvID)
		if conn == nil {
			return &pb_msg.ErrRsp{
				ErrCode: -1,
				ErrMsg:  "用户不存在",
			}, nil
		}
	}

	log.Info(fmt.Sprintf("接收消息：%+v", msg))
	byteMsg, err := proto.Marshal(msg)
	if err != nil {
		return &pb_msg.ErrRsp{
			ErrCode: -2,
			ErrMsg:  "消息解析错误",
		}, err
	}
	err = ws.WriteMsg(conn, websocket.BinaryMessage, byteMsg)
	if err != nil {
		return &pb_msg.ErrRsp{
			ErrCode: -3,
			ErrMsg:  "消息写入ws失败",
		}, err
	}
	return &pb_msg.ErrRsp{
		ErrCode: 0,
		ErrMsg:  "成功",
	}, nil
}

package controller

import (
	"context"
	"fmt"
	"github.com/Shopify/sarama"
	"google.golang.org/protobuf/proto"
	"lalala_im/data/Repo/ChatLog"
	"lalala_im/data/Repo/SaveLog"
	"lalala_im/data/model/MongoModel"
	log "lalala_im/pkg/la_log"
	"lalala_im/proto/pb_msg"
	"lalala_im/servers/save_server/internal/db"
	"sort"
	"strings"
)

var _ = sarama.ConsumerGroupHandler(&SaveConsumer{})

type SaveConsumer struct {
	chatRepo ChatLog.IMongoChatLogRepo
	saveRepo SaveLog.ISaveLogRepo
}

func (m *SaveConsumer) Setup(session sarama.ConsumerGroupSession) error {
	return nil
}

func (m *SaveConsumer) Cleanup(session sarama.ConsumerGroupSession) error {
	return nil
}

func (m *SaveConsumer) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	defer func() {
		if err := recover(); err != nil {
			log.Error(fmt.Sprintf("消费消息持久化出现panic：%+v", err))
			return
		}
	}()
	for message := range claim.Messages() {
		ctx := context.Background()

		msg := &pb_msg.Msg{}
		err := proto.Unmarshal(message.Value, msg)
		if err != nil {
			log.Error(fmt.Sprintf("消费消息出错，消息：%+v，错误：%+v", message, err))
			// 处理失败，标记消息为失败
			session.MarkMessage(message, "processing failed")
			continue
		}
		log.Info(fmt.Sprintf("消费 msg:%+v", msg))

		//语音消息不存储
		if msg.ContentType == 9 {
			continue
		}

		ChatId := msg.GroupID
		//单聊
		if msg.SessionType == 0 {
			ss := []string{msg.SendID, msg.RecvID}
			sort.Strings(ss)
			ChatId = strings.Join(ss, "_")
		}

		err = m.chatRepo.AppendMsgByChatId(ctx, ChatId,
			&MongoModel.MsgInfo{
				SendTime: msg.SendTime,
				Msg:      &message.Value,
			})
		if err != nil {
			log.Error(fmt.Sprintf("添加消息进临时聊天记录chatLog失败：%+v", err))
			// 处理失败，标记消息为失败
			session.MarkMessage(message, "processing failed")
			continue
		}
		//起协程去更新临时聊天记录，无视错误，只打印错误（错了也不会有大影响）
		go func() {
			defer func() {
				if err := recover(); err != nil {
					log.Error(fmt.Sprintf("更新临时聊天记录chatLog出现panic：%+v", err))
					return
				}
			}()
			err := m.chatRepo.PullMsgByChatId(ctx, ChatId)
			if err != nil {
				log.Error(fmt.Sprintf("更新临时聊天记录chatLog错误：%+v", err))
				return
			}
		}()
		newSaveLog := &MongoModel.SaveLog{
			ChatId:   ChatId,
			ChatType: int(msg.SessionType),
			SendTime: msg.SendTime,
			Msg:      &message.Value,
		}
		err = m.saveRepo.CreatSaveLog(ctx, newSaveLog)
		if err != nil {
			// 处理失败，标记消息为失败
			session.MarkMessage(message, "processing failed")
			continue
		}
		//保证已经消费消息
		session.MarkMessage(message, "")
	}
	return nil
}

//
//func (m *SaveConsumer) GetClient(addr string) pb_msg.MsgServerClient {
//	m.RLock()
//	cli, ok := m.ConnMap[addr]
//	m.RUnlock()
//	if !ok {
//
//		conn, err := grpc.Dial(addr, grpc.WithDefaultCallOptions(
//			grpc.MaxCallRecvMsgSize(1024*1024*10), // 最大消息接收大小为1MB
//			grpc.MaxCallSendMsgSize(1024*1024*10), // 最大消息发送大小为1MB
//		), grpc.WithInsecure())
//		if err != nil {
//			log.Error(fmt.Sprintf("创建grpcClient失败，addr：%s,错误：%+v", addr, err))
//			return nil
//		}
//		cli = pb_msg.NewMsgServerClient(conn)
//		m.Lock()
//		m.ConnMap[addr] = cli
//		m.Unlock()
//	}
//	return cli
//}
//
//func (m *SaveConsumer) DelClient(addr string) {
//	m.Lock()
//	delete(m.ConnMap, addr)
//	m.Unlock()
//}

func NewMsgConsumer() sarama.ConsumerGroupHandler {
	return &SaveConsumer{
		chatRepo: ChatLog.NewIChatLogRepo(db.DB.MongoDB),
		saveRepo: SaveLog.NewISaveLogRepo(db.DB.MongoDB),
	}
}

package test

import (
	"context"
	"fmt"
	"github.com/Shopify/sarama"
	"github.com/davecgh/go-spew/spew"
	"github.com/qiniu/qmgo"
	db "lalala_im/data"
	"lalala_im/data/Repo/ChatLog"
	"lalala_im/data/Repo/Friend"
	"lalala_im/data/Repo/User"
	"lalala_im/data/model/MongoModel"
	log "lalala_im/pkg/la_log"
	"strconv"
	"testing"
)

var MongoDB *qmgo.Database

func Init() {
	MongoDB = db.NewMongoDB("mongodb://admin:lalala@192.168.242.128:27017/?authSource=admin", "la_DB")
}

func TestDataChatLog(t *testing.T) {
	Init()
	chatRepo := ChatLog.NewIChatLogRepo(MongoDB)

	ctx := context.Background()

	for i := 0; i < 10; i++ {
		err := chatRepo.CreateChatLog(ctx, &MongoModel.ChatLog{
			ChatId:   "lalala_" + strconv.Itoa(i),
			ChatType: 0,
			Msg: []*MongoModel.MsgInfo{&MongoModel.MsgInfo{
				SendTime: int64(i),
				Msg:      nil,
			}},
		})
		if err != nil {
			panic(err)
		}
	}

	id, err := chatRepo.GetChatLogIDByChatId(ctx, "lalala_5")
	if err != nil {
		panic(err)
	}

	one_chatlog, err := chatRepo.GetChatLogById(ctx, id)
	if err != nil {
		panic(err)
	}
	log.Info(fmt.Sprintf("用ID查：%+v", one_chatlog))
	one_chatlog, err = chatRepo.GetChatLogByChatId(ctx, "lalala_5")
	if err != nil {
		panic(err)
	}
	log.Info(fmt.Sprintf("用ID查：%+v", one_chatlog))
}

func TestFriend(t *testing.T) {
	Init()
	ctx := context.Background()
	frepo := Friend.NewIFriendRepo(MongoDB)

	for i := 0; i < 10; i++ {
		for j := 0; j < 10; j++ {
			if i == j {
				continue
			}
			ii, jj := strconv.Itoa(i), strconv.Itoa(j)
			tempF, err := frepo.GetFriend(ctx, ii, jj)
			if err != nil {
				if err != qmgo.ErrNoSuchDocuments {
					panic(err)
				}
			}
			if tempF != nil && tempF.Id.IsZero() {
				err = frepo.CreateFriend(ctx, ii, jj)
				if err != nil {
					panic(err)
				}
			}

		}
	}

	flist, err := frepo.GetFriendListByUID(ctx, strconv.Itoa(1))
	if err != nil {
		panic(err)
	}
	for i := 0; i < len(flist); i++ {
		spew.Dump(flist[i])
	}

}

func TestUserRepo(t *testing.T) {
	Init()
	userRepo := User.NewIUserRepo(MongoDB)

	ctx := context.Background()
	info, err := userRepo.GetUserInfoByUID(ctx, "\u0013����)M���8{�d)g")
	if err != nil {
		panic(err)
	}
	fmt.Sprintf("ingo: %+v\n", info)
}

// 自定义ConsumerGroupHandler，用于处理消费者接收到的消息
type exampleConsumerGroupHandler struct{}

func (exampleConsumerGroupHandler) Setup(_ sarama.ConsumerGroupSession) error {
	return nil
}

func (exampleConsumerGroupHandler) Cleanup(_ sarama.ConsumerGroupSession) error {
	return nil
}

func (exampleConsumerGroupHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for message := range claim.Messages() {
		fmt.Printf("消费者 Message topic:%s partition:%d offset:%d value:%s\n", message.Topic, message.Partition, message.Offset, string(message.Value))
		session.MarkMessage(message, "")
	}
	return nil
}
func TestKafka(t *testing.T) {
	topicName := "la_msg"
	groupName := "lalala"
	cli := db.NewKafka([]string{"192.168.242.128:9092"})
	// 创建新topic
	//topicDetail := &sarama.TopicDetail{
	//	NumPartitions:     1,
	//	ReplicationFactor: 1,
	//}
	//admin, err := sarama.NewClusterAdminFromClient(cli)
	//if err != nil {
	//	panic(err)
	//}
	//topicList, err := admin.DescribeTopics([]string{topicName})
	//if err != nil {
	//	panic(err)
	//}
	//isExist := false
	//for i := 0; i < len(topicList); i++ {
	//	if topicList[i].Name == topicName {
	//		isExist = true
	//	}
	//}
	//if !isExist {
	//	err = admin.CreateTopic(topicName, topicDetail, false)
	//	if err != nil {
	//		panic(err)
	//	}
	//}

	//consumerGroupList,err:=admin.DescribeConsumerGroups([]string{groupName})
	//isExist=false
	//for i:=0;i<len(consumerGroupList);i++{
	//	if consumerGroupList[i].GroupId==topicName{
	//		isExist=true
	//	}
	//}
	//if !isExist{
	//	err=admin.CreateTopic(topicName,topicDetail,false)
	//	if err!=nil{
	//		panic(err)
	//	}
	//}
	// 创建新consumer group
	group, err := sarama.NewConsumerGroupFromClient(groupName, cli)
	if err != nil {
		panic(err)
	}

	topics := []string{topicName}
	handler := exampleConsumerGroupHandler{}

	err = group.Consume(context.Background(), topics, handler)
	if err != nil {
		panic(err)
	}

}

func TestDelTopics(t *testing.T) {
	topicName := []string{"lalala", "la_msg", "la_msg1", "123"}
	cli := db.NewKafka([]string{"192.168.242.128:9092"})
	//创建新topic
	topicDetail := &sarama.TopicDetail{
		NumPartitions:     1,
		ReplicationFactor: 1,
	}
	admin, err := sarama.NewClusterAdminFromClient(cli)
	if err != nil {
		panic(err)
	}
	topicList, err := admin.DescribeTopics(topicName)
	if err != nil {
		panic(err)
	}
	isExist := false
	for i := 0; i < len(topicList); i++ {
		if topicList[i].Err != 0 {
			//break
			err = admin.DeleteTopic(topicList[i].Name)
			if err != nil {
				panic(err)
			}
		}
	}
	if !isExist {
		err = admin.CreateTopic(topicName[1], topicDetail, false)
		if err != nil {
			panic(err)
		}
	}
}

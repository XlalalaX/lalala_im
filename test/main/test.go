package main

import (
	"fmt"
	"github.com/Shopify/sarama"
)

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

//func main() {
//
//	topicName := "lalala"
//	groupName := "lalala1"
//	cli := db.NewKafka([]string{"192.168.242.128:9092"})
//	// 创建新topic
//	topicDetail := &sarama.TopicDetail{
//		NumPartitions:     1,
//		ReplicationFactor: 1,
//	}
//	admin, err := sarama.NewClusterAdminFromClient(cli)
//	if err != nil {
//		panic(err)
//	}
//	topicList, err := admin.DescribeTopics([]string{topicName})
//	if err != nil {
//		panic(err)
//	}
//	isExist := false
//	for i := 0; i < len(topicList); i++ {
//		if topicList[i].Name == topicName {
//			isExist = true
//		}
//	}
//	if !isExist {
//		err = admin.CreateTopic(topicName, topicDetail, false)
//		if err != nil {
//			panic(err)
//		}
//	}
//
//	//consumerGroupList, err := admin.DescribeConsumerGroups([]string{groupName})
//	//isExist = false
//	//for i := 0; i < len(consumerGroupList); i++ {
//	//	if consumerGroupList[i].GroupId == groupName {
//	//		isExist = true
//	//	}
//	//}
//	//if !isExist {
//	//	err = admin.CreateTopic(topicName, topicDetail, false)
//	//	if err != nil {
//	//		panic(err)
//	//	}
//	//}
//
//	//创建新consumer group
//	group, err := sarama.NewConsumerGroupFromClient(groupName, cli)
//	if err != nil {
//		panic(err)
//	}
//
//	go func() {
//		topics := []string{topicName}
//		handler := exampleConsumerGroupHandler{}
//
//		err = group.Consume(context.Background(), topics, handler)
//		if err != nil {
//			panic(err)
//		}
//	}()
//
//	producer, err := sarama.NewSyncProducerFromClient(cli)
//	if err != nil {
//		panic(err)
//	}
//	i := 0
//	for {
//		msg := &sarama.ProducerMessage{
//			Topic:     topicName,
//			Key:       nil,
//			Value:     sarama.StringEncoder(strconv.Itoa(i)),
//			Headers:   nil,
//			Metadata:  nil,
//			Offset:    0,
//			Partition: 0,
//			Timestamp: time.Time{},
//		}
//		p, offset, err := producer.SendMessage(msg)
//		fmt.Printf("生产：%d 返回：p:%d offset:%d err:%v\n", i, p, offset, err)
//		time.Sleep(1 * time.Second)
//		i++
//	}
//}

func main() {
	s := "ws_pool_out_addr:12345"
	fmt.Println(s[17:])

}

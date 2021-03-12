package kafka

import (
	"github.com/Shopify/sarama"
	"github.com/sirupsen/logrus"
)

// kafka 相关操作

var (
	client  sarama.SyncProducer
	msgChan chan *sarama.ProducerMessage
)

// Init 初始化连接kafka的全局client
func Init(address []string, channelSize int64) (err error) {
	// 1、生产者配置
	config := sarama.NewConfig()
	// 发送完数据需要Leader和follower都确认
	config.Producer.RequiredAcks = sarama.WaitForAll
	// 随机选择一个Partition
	config.Producer.Partitioner = sarama.NewRandomPartitioner
	// 成功交付的消息将在success channel返回
	config.Producer.Return.Successes = true

	// 2、连接Kafka
	client, err = sarama.NewSyncProducer(address, config)
	if err != nil {
		logrus.Errorf("[kafka] Producer closed, err: %v", err)
		return
	}

	// 初始化MsgChan
	msgChan = make(chan *sarama.ProducerMessage, channelSize)
	// 起一个后台的goroutine从MsgChan中读数据
	go sendMsg()
	return
}

// sendMsg 从MsgChan中读取msg，发送给kafka
func sendMsg() {
	for {
		select {
		case msg := <-msgChan:
			pid, offset, err := client.SendMessage(msg)
			if err != nil {
				logrus.Warningf("[kafka] Send msg failed, err: %v", err)
				return
			}
			logrus.Infof("[kafka] Send msg to kafka success, topic: %s, pid: %v, offset: %v", msg.Topic, pid, offset)
		}
	}
}

// ToMsgChan 定义一个函数向外暴露msgChan
func ToMsgChan(msg *sarama.ProducerMessage) {
	msgChan <- msg
}

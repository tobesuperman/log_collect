package kafka

import (
	"encoding/json"
	"github.com/Shopify/sarama"
	"github.com/sirupsen/logrus"
	"log_transfer/es"
)

// kafkaConsumer kafka消费者的结构体
type kafkaConsumer struct {
	address  []string
	topic    string
	consumer sarama.Consumer
}

// newKafkaConsumer 根据address和topic创建消费者对象
func newKafkaConsumer(address []string, topic string) (kc *kafkaConsumer, err error) {
	kc = &kafkaConsumer{
		address: address,
		topic:   topic,
	}
	kc.consumer, err = sarama.NewConsumer(kc.address, nil)
	return
}

func (kc *kafkaConsumer) run() {
	logrus.Infof("[kafka] Consumer for topic: %s start", kc.topic)
	// 根据topic取到所有的分区
	partitionList, err := kc.consumer.Partitions(kc.topic)
	if err != nil {
		logrus.Errorf("[kafka] Get list of partition failed, err: %v", err)
		return
	}
	for partition := range partitionList {
		// 针对每个分区创建一个对应的分区消费者
		pc, err := kc.consumer.ConsumePartition(kc.topic, int32(partition), sarama.OffsetNewest)
		if err != nil {
			logrus.Errorf("[kafka] Start consumer for partition %d failed, err: %v", partition, err)
			continue
		}
		// 异步从每个分区消费消息
		go func(sarama.PartitionConsumer) {
			for msg := range pc.Messages() {
				// 将取出的日志数据先放到channel中将同步的流程异步化
				var m map[string]interface{}
				err = json.Unmarshal(msg.Value, &m)
				if err != nil {
					logrus.Errorf("[kafka] Json unmarshal msg failed, err: %v", err)
					continue
				}
				// 这里默认kafka中的topic和ES中的index相同
				es.SendLogData(kc.topic, m)
				//fmt.Printf("partition: %d, offset: %d, key: %s, value: %s\n", msg.Partition, msg.Offset, msg.Key, msg.Value)
			}
		}(pc)
	}
}

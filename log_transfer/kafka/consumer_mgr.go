package kafka

import (
	"github.com/sirupsen/logrus"
	"log_transfer/common"
)

type consumerMgr struct {
	address     []string                  // kafka连接的IP地址
	consumerMap map[string]*kafkaConsumer // 所有的kafka消费者
	topicList   []string                  // 所有的topic
	topicChan   chan []string             // 等待新topic的通道
}

var (
	cMgr *consumerMgr
)

// Init 初始化一个全局的kafka消费者管理器，为每一个topic创建一个单独的消费者，从kafka中取出日志数据
func Init(address []string, allConf []common.CollectionEntry) (err error) {
	cMgr = &consumerMgr{
		consumerMap: make(map[string]*kafkaConsumer, 20),
		topicList:   common.GetTopicFromCollectionEntry(allConf),
		// 初始化新配置的channel（无缓冲区的）
		topicChan: make(chan []string),
		address:   address,
	}
	for _, topic := range cMgr.topicList {
		// 创建一个kafka消费者
		kc, err := newKafkaConsumer(address, topic)
		if err != nil {
			logrus.Errorf("[kafka] Start consumer for topic: %s failed, err: %v", topic, err)
			continue
		}
		logrus.Infof("[kafka] Start a consumer for topic: %s success", topic)
		cMgr.consumerMap[topic] = kc
		go kc.run()
	}
	// 启动一个后台的goroutine等待新配置的到来
	go cMgr.watch()
	return
}

// watch 一直在等topicChan中有值，有值就开始去管理之前的kafka消费者
// 管理分两种情况：
// 1、原来有就什么都不干
// 2、原来没有现在有就新建
func (cMgr *consumerMgr) watch() {
	for {
		// 能取到值说明新的topic已到来
		newTopic := <-cMgr.topicChan
		// 新topic来了之后应该管理之前启动的那些kafka消费者
		logrus.Infof("[kafka] Get new conf from etcd success, conf: %v", newTopic)
		for _, topic := range newTopic {
			// 1、原来已经存在的消费者无需理会
			if cMgr.isExist(topic) {
				continue
			}
			// 2、原来没有的需要新创建一个消费者
			kc, err := newKafkaConsumer(cMgr.address, topic)
			if err != nil {
				logrus.Errorf("[kafka] Start consumer for topic: %s failed, err: %v", topic, err)
				continue
			}
			logrus.Infof("[kafka] Start a consumer for topic: %s success", topic)
			cMgr.consumerMap[topic] = kc
			go kc.run()
		}
	}
}

// isExist 判断consumerMap中是否存在该消费者
func (cMgr *consumerMgr) isExist(topic string) bool {
	_, ok := cMgr.consumerMap[topic]
	return ok
}

// SendNewTopic 把新的topic放到管理器的topicChan中
func SendNewTopic(newConf []common.CollectionEntry) {
	cMgr.topicChan <- common.GetTopicFromCollectionEntry(newConf)
}

package es

import (
	"context"
	"fmt"
	"github.com/olivere/elastic/v7"
	"github.com/sirupsen/logrus"
	"log_transfer/common"
)

type ESClient struct {
	client         *elastic.Client
	indexes        []string
	goroutineNum   int
	channelSize    int
	indexChan      chan []string
	logDataChanMap map[string]chan interface{} // 不同topic的数据将会发送到ES的不同index中
}

var (
	esClient *ESClient
)

// 将日志数据写入ES中
func Init(address string, indexes []string, channelSize, goroutineNum int) (err error) {
	esClient = &ESClient{
		logDataChanMap: make(map[string]chan interface{}),
		goroutineNum:   goroutineNum,
		channelSize:    channelSize,
		indexes:        indexes,
	}
	esClient.client, err = elastic.NewClient(elastic.SetURL("http://" + address))
	if err != nil {
		logrus.Errorf("[es] Connect to ES failed, err: %v", err)
		return
	}
	logrus.Info("[es] Connect to ES success!")
	// 默认Kafka中的topic就是ES中的index
	for _, index := range indexes {
		esClient.logDataChanMap[index] = make(chan interface{}, channelSize)
		// 从channel中取出数据，写入到ES中
		for i := 0; i < goroutineNum; i++ {
			go sendToES(index)
		}
	}
	// 启动一个后台的goroutine等待新配置的到来
	go esClient.watch()
	return
}

func sendToES(index string) {
	for {
		select {
		case msg := <-esClient.logDataChanMap[index]:
			put, err := esClient.client.Index().
				Index(index).
				BodyJson(msg).
				Do(context.Background())
			if err != nil {
				logrus.Errorf("[es] Send msg to ES failed, err: %v", err)
				continue
			}
			fmt.Printf("Indexed user %s to index %s, type is %s\n", put.Id, put.Index, put.Type)
		}
	}
}

// 通过一个可导出的函数从包外接收msg，放到channel中
func SendLogData(index string, msg interface{}) {
	esClient.logDataChanMap[index] <- msg
}

// watch 一直在等indexChan中有值，有值就开始去管理之前的index
func (ec *ESClient) watch() {
	for {
		// 能取到值说明新的配置已到来
		newIndexes := <-esClient.indexChan
		// 新配置来了之后应该管理之前建立的那些index
		logrus.Infof("[es] Get new conf from etcd success, conf: %v", newIndexes)
		for _, index := range newIndexes {
			// 1、原来已经存在无需理会
			if ec.isExist(index) {
				continue
			}
			// 2、原来没有的需要新创建一个index
			esClient.logDataChanMap[index] = make(chan interface{}, ec.channelSize)
			// 从channel中取出数据，写入到kafka中
			for i := 0; i < ec.goroutineNum; i++ {
				go sendToES(index)
			}
		}
	}
}

// SendNewTopic 把新的index放到indexChan中
func SendNewIndex(newConf []common.CollectionEntry) {
	esClient.indexChan <- common.GetTopicFromCollectionEntry(newConf)
}

// isExist 判断index是否已存在
func (ec *ESClient) isExist(index string) bool {
	for _, value := range ec.indexes {
		if value == index {
			return true
		}
	}
	return false
}

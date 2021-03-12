package main

import (
	"github.com/go-ini/ini"
	"github.com/sirupsen/logrus"
	"log_transfer/common"
	"log_transfer/es"
	"log_transfer/etcd"
	"log_transfer/kafka"
	"strings"
)

// 日志收集的log_transfer
// 从kafka消费日志数据，写入ES中

func main() {
	// 1、加载配置文件
	var config = new(common.Config)
	err := ini.MapTo(config, "./conf/config.ini")
	if err != nil {
		logrus.Errorf("[main] Load config failed, err: %v", err)
		return
	}
	logrus.Info("[main] Load config success")

	// 2、初始化etcd连接
	err = etcd.Init(strings.Split(config.EtcdConfig.Address, ","))
	if err != nil {
		logrus.Errorf("[main] Connect to etcd failed, err: %v", err)
		return
	}
	logrus.Info("[main] Connect to etcd success")

	// 3、从etcd中拉取要收集的日志配置项
	collectionKey := config.EtcdConfig.CollectionKey
	allConf, err := etcd.GetConf(collectionKey)
	topics := common.GetTopicFromCollectionEntry(allConf)

	// 4、启动一个后台的goroutine，监控etcd中的日志收集项是否发生变化
	go etcd.WatchConf(collectionKey)

	// 5、连接kafka
	err = kafka.Init(strings.Split(config.KafkaConfig.Address, ","), allConf)
	if err != nil {
		logrus.Errorf("[main] Connect to kafka failed, err: %v", err)
		return
	}
	logrus.Info("[main] Connect to kafka success")

	// 6、连接ES
	err = es.Init(config.ESConfig.Address, topics, config.ESConfig.ChannelSize, config.ESConfig.GoroutineNum)
	if err != nil {
		logrus.Errorf("[main] Connect to ES failed, err: %v", err)
		return
	}
	logrus.Info("[main] Connect to ES success")

	select {}
}

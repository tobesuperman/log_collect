package main

import (
	"github.com/go-ini/ini"
	"github.com/sirupsen/logrus"
	"log_agent/common"
	"log_agent/etcd"
	"log_agent/kafka"
	"log_agent/tailfile"
	"strings"
)

// 日志收集的客户端log_agent
// 收集指定目录下的日志文件，发送到kafka中

func main() {
	var config = new(common.Config)
	// 1、读配置文件
	err := ini.MapTo(config, "./conf/config.ini")
	if err != nil {
		logrus.Errorf("[main] Load config failed, err: %v", err)
		return
	}
	logrus.Info("[main] Load config success")

	// 2、初始化kafka连接
	err = kafka.Init(strings.Split(config.KafkaConfig.Address, ","), config.KafkaConfig.ChannelSize)
	if err != nil {
		logrus.Errorf("[main] Connect to kafka failed, err: %v", err)
		return
	}
	logrus.Info("[main] Connect to kafka success")

	// 3、初始化etcd连接
	err = etcd.Init(strings.Split(config.EtcdConfig.Address, ","))
	if err != nil {
		logrus.Errorf("[main] Connect to etcd failed, err: %v", err)
		return
	}
	logrus.Info("[main] Connect to etcd success")

	// 4、从etcd中拉取要收集日志的配置项
	collectionKey := config.EtcdConfig.CollectionKey
	allConf, err := etcd.GetConf(collectionKey)
	// 只获取属于本机的日志收集任务配置项
	allConf = common.FilterByIp(allConf, etcd.IP)
	if err != nil {
		logrus.Errorf("[main] Get conf from etcd failed, err: %v", err)
		return
	}

	// 5、启动一个后台的goroutine，监控etcd中的日志收集项是否发生变化
	go etcd.WatchConf(config.EtcdConfig.CollectionKey)

	// 6、根据配置中的日志路径初始化tailTask
	err = tailfile.Init(allConf)
	if err != nil {
		logrus.Errorf("[main] Init tail failed, err: %v", err)
		return
	}
	logrus.Info("[main] Init tail success")
	select {}
}

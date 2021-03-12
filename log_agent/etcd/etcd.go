package etcd

import (
	"context"
	"encoding/json"
	storagepb2 "github.com/coreos/etcd/storage/storagepb"
	"github.com/sirupsen/logrus"
	"go.etcd.io/etcd/clientv3"
	"log_agent/common"
	"log_agent/tailfile"
	"time"
)

// etcd相关操作

var (
	client *clientv3.Client
	IP     string
)

// Init 初始化连接etcd的全局client
func Init(address []string) (err error) {
	client, err = clientv3.New(clientv3.Config{
		Endpoints:   address,
		DialTimeout: time.Second * 5,
	})
	if err != nil {
		logrus.Errorf("[etcd] Connect to etcd failed, err: %v", err)
		return
	}
	// 获取本机IP，以便后续从etcd中获取配置项
	IP, err = common.GetOutboundIP()
	if err != nil {
		logrus.Errorf("[etcd] Get outboud ip failed, err: %v", err)
		return
	}
	return
}

// GetConf 从etcd中根据给定的key获取日志收集配置项
func GetConf(key string) (collectionEntryList []common.CollectionEntry, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
	defer cancel()
	resp, err := client.Get(ctx, key)
	if err != nil {
		logrus.Errorf("[etcd] Get conf from etcd by key: %s, failed, err: %v", key, err)
		return
	}
	if len(resp.Kvs) == 0 {
		logrus.Warningf("[etcd] Get len:0 conf from etcd by key: %s", key)
		return
	}
	ret := resp.Kvs[0]
	// json格式式字符串
	err = json.Unmarshal(ret.Value, &collectionEntryList)
	if err != nil {
		logrus.Errorf("[etcd] Json unmarshal conf failed, err: %v", err)
		return
	}
	return
}

// WatchConf 监控etcd中日志收集项配置变化
func WatchConf(key string) {
	for {
		// 监控
		watchChan := client.Watch(context.Background(), key)
		for wResp := range watchChan {
			logrus.Info("[etcd] Get new conf from etcd success!")
			for _, evt := range wResp.Events {
				//fmt.Printf("type:%s, key: %s, value: %s\n", evt.Type, evt.Kv.Key, evt.Kv.Value)
				var newConf []common.CollectionEntry
				if evt.Type == storagepb2.DELETE {
					// 如果是删除
					logrus.Warningf("[etcd] Delete the key!")
					tailfile.SendNewConf(newConf)
					continue
				}
				err := json.Unmarshal(evt.Kv.Value, &newConf)
				if err != nil {
					logrus.Errorf("[etcd] Json unmarshal new conf failed, err: %v", err)
					continue
				}
				// 通知tail模块应该启用新的配置
				newConf = common.FilterByIp(newConf, IP)
				tailfile.SendNewConf(newConf) // 没有任何接收就是阻塞的
			}
		}
	}
}

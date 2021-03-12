package main

import (
	"context"
	"encoding/json"
	"fmt"
	"go.etcd.io/etcd/clientv3"
	"net"
	"time"
)

// Conf 日志收集配置项
type Conf struct {
	Path  string `json:"path"`
	Topic string `json:"topic"`
}

func main() {
	// 获取本机的IP地址
	ip, err := getOutboundIP()
	if err != nil {
		panic(err)
		return
	}
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{"127.0.0.1:2379"},
		DialTimeout: time.Second * 5,
	})
	if err != nil {
		fmt.Printf("connect to etdc failed, err: %v\n", err)
		return
	}
	defer cli.Close()

	// 模拟两个日志收集任务，用于测试
	conf1 := Conf{
		Path:  "d:/logs/nginx.log",
		Topic: fmt.Sprintf("%s_nginx", ip),
	}
	conf2 := Conf{
		Path:  "d:/logs/mysql.log",
		Topic: fmt.Sprintf("%s_mysql", ip),
	}
	configs := [...]Conf{conf1, conf2}
	b, err := json.Marshal(&configs)
	if err != nil {
		fmt.Println(err)
		return
	}
	str := string(b)

	// 写入日志收集任务的配置项
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	key := "collection_log_conf"
	_, err = cli.Put(ctx, key, str)
	if err != nil {
		fmt.Printf("put to etdc failed, err: %v\n", err)
		return
	}
	cancel()

	// 获取日志收集任务的配置项
	ctx, cancel = context.WithTimeout(context.Background(), time.Second)
	gr, err := cli.Get(ctx, key)
	if err != nil {
		fmt.Printf("get from etdc failed, err: %v\n", err)
		return
	}
	for _, ev := range gr.Kvs {
		fmt.Printf("key: %s, value: %s\n", ev.Key, ev.Value)
	}
	cancel()
}

func getOutboundIP() (ip string, err error) {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return
	}
	defer conn.Close()
	localAddr := conn.LocalAddr().(*net.UDPAddr)
	ip = localAddr.IP.String()
	return
}

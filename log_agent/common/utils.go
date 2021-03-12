package common

import (
	"github.com/sirupsen/logrus"
	"net"
	"strings"
)

// GetOutboundIP 获取本机的IP
func GetOutboundIP() (ip string, err error) {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		logrus.Errorf("[utils] Get outbound ip failed, err: %v\n", err)
		return
	}
	defer conn.Close()
	localAddr := conn.LocalAddr().(*net.UDPAddr)
	ip = localAddr.IP.String()
	return
}

// FilterByIp 选择本机IP对应的日志收集配置项
func FilterByIp(allConf []CollectionEntry, ip string) []CollectionEntry {
	conf := make([]CollectionEntry, 0)
	for _, value := range allConf {
		if strings.HasPrefix(value.Topic, ip) {
			conf = append(conf, value)
		}
	}
	return conf
}

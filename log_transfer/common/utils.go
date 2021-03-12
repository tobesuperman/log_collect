package common

import (
	"github.com/sirupsen/logrus"
	"net"
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

func GetTopicFromCollectionEntry(conf []CollectionEntry) []string {
	topics := make([]string, len(conf))
	for i := 0; i < len(conf); i++ {
		topics[i] = conf[i].Topic
	}
	return topics
}

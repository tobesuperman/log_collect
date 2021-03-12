package common

type Config struct {
	KafkaConfig `ini:"kafka"`
	ESConfig    `ini:"es"`
	EtcdConfig  `ini:"etcd"`
}

type KafkaConfig struct {
	Address string `ini:"address"`
}

type ESConfig struct {
	Address      string `ini:"address"`
	ChannelSize  int    `ini:"channel_size"`
	GoroutineNum int    `ini:"goroutine_num"`
}

type EtcdConfig struct {
	Address       string `ini:"address"`
	CollectionKey string `ini:"collection_key"`
}

// CollectionEntry 要收集的日志的配置项结构体
type CollectionEntry struct {
	Path  string `json:"path"`
	Topic string `json:"topic"`
}

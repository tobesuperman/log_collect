package common

// 整个log_agent项目的配置结构体
type Config struct {
	KafkaConfig `ini:"kafka"`
	EtcdConfig  `ini:"etcd"`
}

type KafkaConfig struct {
	Address     string `ini:"address"`
	ChannelSize int64  `ini:"channel_size"`
}

type EtcdConfig struct {
	Address       string `ini:"address"`
	CollectionKey string `ini:"collection_key"`
}

// CollectionEntry 要收集的日志的配置项结构体
type CollectionEntry struct {
	Path  string `json:"path"`  // 从哪个路径读取日志文件
	Topic string `json:"topic"` // 日志文件中的内容发往kafka中的哪个topic
}

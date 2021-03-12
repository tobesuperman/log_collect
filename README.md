## Golang日志收集项目

### 项目介绍

使用Golang进行开发，采用生产者-消费者模式，实现对业务日志的实时收集与存储。

由于环境限制，Kafka、etcd、ElasticSearch这些中间件均部署在同一台电脑上。

使用到的第三方库：

* github.com/Shopify/sarama
* github.com/olivere/elastic/v7
* go.etcd.io/etcd
* github.com/hpcloud/tail
* github.com/go-ini/ini
* github.com/sirupsen/logrus

项目一共分为三个模块：

* log_agent模块：收集指定目录下的日志文件，发送（生产）至Kafka中
* log_transfer模块：从Kafka中读取（消费）日志数据，写入ElasticSearch中
* etcd_demo模块：模拟向etcd中写入日志收集任务的配置项，用来测试两个功能模块
### log_agent模块

具体流程：
1. 初始化工作，包括读取Kafka、etcd的连接配置，以及获取本机的IP地址
2. 连接Kafka，并开启goroutine异步地向Kafka写入读取到的日志数据
3. 连接etcd，获取日志收集任务的配置项，并开启一个goroutine监听etcd中配置项的变化（对配置项进行过滤，只关心本服务器对应的日志收集任务）
4. 初始化日志收集任务管理器，根据配置项开启goroutine，用于承载日志收集任务
5. 开启一个后台的goroutine监听新配置项的到来，日志收集任务管理器根据新的配置项动态新增、停止日志收集任务

### log_transfer模块

具体流程：
1. 初始化工作，包括读取Kafka、ElasticSearch的连接配置
2. 连接etcd，获取日志收集任务的配置项（只获取topic），并开启一个goroutine监听etcd中配置项的变化
3. 初始化Kafka消费者管理器，根据配置项中的topic开启goroutine，用于从Kafka中读取日志数据
4. 开启一个后台的goroutine监听新配置项的到来，Kafka消费者管理器根据新的配置项动态新增消费者
5. 连接ElasticSearch，根据index开启goroutine，异步地将日志数据写入ES中
6. 开启一个后台的goroutine监听新配置项的到来，ESCilent模块根据新的配置项动态新增index

### etcd_demo模块

初始化etcd连接，将日志收集任务的配置项序列化后，写入etcd中。
其中将服务器的IP地址与业务线名称拼接，作为Kafka中的topic，用以区分不同服务器上的不同日志收集需求。

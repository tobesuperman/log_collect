package tailfile

import (
	"context"
	"github.com/Shopify/sarama"
	"github.com/hpcloud/tail"
	"github.com/sirupsen/logrus"
	"log_agent/kafka"
	"strings"
	"time"
)

// tail 相关操作

// tailTask 日志收集任务的结构体
type tailTask struct {
	path   string
	topic  string
	tObj   *tail.Tail
	ctx    context.Context
	cancel context.CancelFunc
}

// newTailTask 根据path和topic创建tailTask对象
func newTailTask(path, topic string) *tailTask {
	ctx, cancel := context.WithCancel(context.Background())
	tt := &tailTask{
		path:   path,
		topic:  topic,
		ctx:    ctx,
		cancel: cancel,
	}
	return tt
}

// init 使用tail工具打开日志文件准备读
func (tt *tailTask) init() (err error) {
	config := tail.Config{
		ReOpen:    true,
		Follow:    true,
		Location:  &tail.SeekInfo{Offset: 0, Whence: 2},
		MustExist: false,
		Poll:      true,
	}
	// 打开文件开始读取数据
	tt.tObj, err = tail.TailFile(tt.path, config)
	return
}

// run 读取日志向kafka发送
func (tt *tailTask) run() {
	logrus.Infof("[tailfile] Tailtask with collection path: %s start", tt.path)
	for {
		select {
		// 只要调用tt.cancel()，就会收到信号
		case <-tt.ctx.Done():
			logrus.Infof("[tailfile] Stop the tailTask with collection path: %s", tt.path)
			return
		// 循环读数据
		case line, ok := <-tt.tObj.Lines:
			if !ok {
				logrus.Warningf("[tailfile] Reopen closed tail file with path: %s", tt.path)
				time.Sleep(time.Second)
				continue
			}
			// 如果是空行则跳过
			if len(strings.Trim(line.Text, "\r")) == 0 {
				logrus.Infof("[tailfile] Skip blank line")
				continue
			}
			// 利用channel将同步的代码改为异步的
			// 把读出来的一行日志包装成kafka里面的msg类型，放到通道中
			msg := &sarama.ProducerMessage{}
			// 每个日志收集任务各自的topic
			msg.Topic = tt.topic
			msg.Value = sarama.StringEncoder(line.Text)
			// 放到channel中
			kafka.ToMsgChan(msg)
		}
	}
}

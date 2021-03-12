package tailfile

import (
	"github.com/sirupsen/logrus"
	"log_agent/common"
)

// tailTaskMgr tailTask管理器
type tailTaskMgr struct {
	tailTaskMap         map[string]*tailTask          // 所有的tailTask任务
	collectionEntryList []common.CollectionEntry      // 所有配置项
	confChan            chan []common.CollectionEntry // 等待新配置的通道
}

var (
	ttMgr *tailTaskMgr
)

// Init 初始化一个全局的tailTask管理器，为每一个日志收集任务创建一个单独的tailTask
func Init(allConf []common.CollectionEntry) (err error) {
	ttMgr = &tailTaskMgr{
		tailTaskMap:         make(map[string]*tailTask, 20),
		collectionEntryList: allConf,
		// 初始化新配置的channel（无缓冲区的）
		confChan: make(chan []common.CollectionEntry),
	}
	// allConf里保存了若干个日志的收集项
	// 针对每一个日志收集项创建一个对应的tailTask
	for _, conf := range allConf {
		// 创建一个日志收集任务
		tt := newTailTask(conf.Path, conf.Topic)
		// 打开日志文件，准备读取
		err = tt.init()
		if err != nil {
			logrus.Errorf("[tailfile] Create tailTask for path: %s failed, err: %v", tt.path, err)
			continue
		}
		logrus.Infof("[tailfile] Create a tail task for path: %s success", tt.path)
		// 把创建的tailTask任务登记在册，方便后续管理
		ttMgr.tailTaskMap[tt.path] = tt
		// 启动一个后台的goroutine开始收集日志
		go tt.run()
	}
	// 启动一个后台的goroutine等待新配置的到来
	go ttMgr.watch()
	return
}

// watch 一直在等confChan中有值，有值就开始去管理之前的tailTask
// 管理分三种情况：
// 1、原来有就什么都不干
// 2、原来没有现在有就新建
// 3、原来有现在没有就停掉
func (ttMgr *tailTaskMgr) watch() {
	for {
		// 能取到值说明新的配置已到来
		newConf := <-ttMgr.confChan
		// 新配置来了之后应该管理之前启动的那些tailTask
		logrus.Infof("[tailfile] Get new conf from etcd success, conf: %v", newConf)
		for _, conf := range newConf {
			// 1、原来已经存在的任务无需理会
			if ttMgr.isExist(conf) {
				continue
			}
			// 2、原来没有的需要新创建一个tailTask任务
			tt := newTailTask(conf.Path, conf.Topic)
			err := tt.init()
			if err != nil {
				logrus.Errorf("[tailfile] Create tailTask for path: %s failed, err: %v", tt.path, err)
				continue
			}
			logrus.Infof("[tailfile] Create a tail task for path: %s success", tt.path)
			ttMgr.tailTaskMap[tt.path] = tt
			go tt.run()
		}
		// 3、原来有的现在没有的tailTask需要停掉
		// 找出tailTaskMap中存在，但是newConf中不存在的那些tailTask，把它们都停掉
		for key, task := range ttMgr.tailTaskMap {
			var found bool
			for _, conf := range newConf {
				if key == conf.Path {
					found = true
					break
				}
			}
			if !found {
				// 这个tailTask需要停掉
				logrus.Infof("[tailfile] Stop the task with collection path: %s", task.path)
				// 删除tailTask
				delete(ttMgr.tailTaskMap, key)
				task.cancel()
			}
		}
	}
}

// isExist 判断tailTaskMap中是否存在该收集项
func (ttMgr *tailTaskMgr) isExist(conf common.CollectionEntry) bool {
	_, ok := ttMgr.tailTaskMap[conf.Path]
	return ok
}

// SendNewConf 把新的配置放到管理器的confChan中
func SendNewConf(newConf []common.CollectionEntry) {
	ttMgr.confChan <- newConf
}

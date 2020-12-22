package apollo

import (
	"github.com/pkg/errors"
	"go.uber.org/atomic"
	"sync"
	"time"
)

// 长轮训 定时请求是否有namespace更新

type PollConfig struct {
	ConfigServerUrl string
	AppId string
	ClusterName string
	NamespaceNames []string
}

type PollTask struct {
	// namespaceName map 保存每个namespace的notificationId
	namespaceNames map[string]int
	Mutex sync.RWMutex
	HttpRequest HttpRequest
	Config PollConfig
	logger ApolloLogInterface
	start *atomic.Bool
	Quit chan struct{}
	Interval time.Duration
	//wg sync.WaitGroup
	handler notificationHandler // poll handler
	initHandler notificationHandler // 初始化获取配置的 handler
}

func NewPollConfig(pollConfig *PollConfig, handler notificationHandler, initHandler notificationHandler, logger ApolloLogInterface) *PollTask {
	pollTask := PollTask{
		Mutex:          sync.RWMutex{},
		start: atomic.NewBool(false),
		logger: logger,
	}
	pollTask.Config = PollConfig{
		ConfigServerUrl: pollConfig.ConfigServerUrl,
		AppId:           pollConfig.AppId,
		ClusterName:     pollConfig.ClusterName,
	}
	wg := sync.WaitGroup{}
	wg.Add(len(pollConfig.NamespaceNames))
	//pollTask.wg = wg
	pollTask.namespaceNames = InitnamespaceNames(pollConfig.NamespaceNames)
	pollTask.HttpRequest =  NewDefaultHttpRequset(logger)
	pollTask.Interval = poolInterval
	pollTask.handler = handler
	pollTask.initHandler = initHandler
	return &pollTask
}

func InitnamespaceNames(namespaceNames []string) map[string]int {
	namespaceNameMap := make(map[string]int)
	for _, value := range namespaceNames {
		namespaceNameMap[value] = initNotificationId
	}
	return namespaceNameMap
}


/**
pollTask 启动
*/
func (pollTask *PollTask) Start() error {
	if !pollTask.start.CAS(false, true) {
		return errors.New("请勿重复启动pollTask")
	}
	syncPoll := func() {
		NotificationsGet(pollTask, true)
	}

	// 先主动请求一次
	syncPoll()
	return nil
}

func (pollTask *PollTask) StartPoll() error {
	doPoll := func() {
		NotificationsGet(pollTask, false)
	}
	// 然后开启轮训
	go func() {
		timer := time.NewTimer(pollTask.Interval)
		for  {
			select {
			case _ = <-pollTask.Quit :
				pollTask.logger.Info("quit poll")
				return
			case _ = <- timer.C :
				pollTask.logger.Info("polling")
				doPoll()
				timer.Reset(pollTask.Interval)
			}
		}
	}()
	return nil
}

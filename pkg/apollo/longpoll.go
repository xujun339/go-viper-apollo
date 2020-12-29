package apollo

import (
	"context"
	"github.com/asaskevich/EventBus"
	"github.com/pkg/errors"
	"go.uber.org/atomic"
	"sync"
	"time"
)

var (
	// 第一次启动拉取的配置
	ApolloFirstPollTopic = "first-poll"

	// 后续长轮训
	ApolloLongPollTopic = "long-poll"

	// 第一次启动等待回调处理结束等待时间
	ApolloFirstPollWaitTimeout = 6 * time.Second
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
	// 长连接拉取间隔
	Interval time.Duration
	// eventBus
	bus EventBus.Bus
}

func NewPollConfig(pollConfig *PollConfig, logger ApolloLogInterface) *PollTask {
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
	pollTask.namespaceNames = InitnamespaceNames(pollConfig.NamespaceNames)
	pollTask.HttpRequest =  NewDefaultHttpRequset(logger)
	pollTask.Interval = poolInterval
	pollTask.bus = EventBus.New();
	return &pollTask
}

// 订阅启动的时候第一次轮训
func (pollTask *PollTask) SubscribeStart(fn initNotificationHandler) error {
	return pollTask.bus.SubscribeAsync(ApolloFirstPollTopic, fn, false)
}

// 订阅长轮训
func (pollTask *PollTask) SubscribeLongPoll(fn notificationHandler) error {
	return pollTask.bus.Subscribe(ApolloLongPollTopic, fn)
}

func InitnamespaceNames(namespaceNames []string) map[string]int {
	namespaceNameMap := make(map[string]int)
	for _, value := range namespaceNames {
		namespaceNameMap[value] = initNotificationId
	}
	return namespaceNameMap
}

/**
	初始化的时候拉取 , 并阻塞等待事件处理
 */
func (pollTask *PollTask) syncPoll(ctx context.Context) error {
	NotificationsGet(pollTask, true)
	ctx, _ = context.WithTimeout(ctx, ApolloFirstPollWaitTimeout)
	waitErr := pollTask.syncPollWait(ctx)
	return waitErr
}

func (pollTask *PollTask) syncPollWait(ctx context.Context) error {
	var waitRs = make(chan int)
	go func() {
		pollTask.bus.WaitAsync()
		waitRs <- 1
	}()
	select {
		case _ = <-ctx.Done() :
			return errors.New("start poll wait timeout")
		case <-waitRs:;
	}
	return nil
}

/**
	后续长轮训
 */
func (pollTask *PollTask) longPoll() error {
	return NotificationsGet(pollTask, false)
}



/**
pollTask 启动
*/
func (pollTask *PollTask) Start() error {
	if !pollTask.start.CAS(false, true) {
		return errors.New("请勿重复启动pollTask")
	}

	// 主动请求一次
	err := pollTask.syncPoll(context.Background())
	if err != nil {
		return errors.WithMessage(err, "start fail")
	}

	// 开始轮训
	pollTask.StartPoll()
	return nil
}

func (pollTask *PollTask) StartPoll() {
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
				pollTask.longPoll()
				timer.Reset(pollTask.Interval)
			}
		}
	}()
}

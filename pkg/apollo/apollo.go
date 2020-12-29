package apollo

import (
	"sync"
	"time"
)

// apollo 的工具包

var (
	initNotificationId = -1
	spilCode           = ","
	once               sync.Once
	AplloSrvInstance   *ApolloSrv
	poolInterval       = 2 * time.Second
	apolloConfig       ApolloConfig
)

type ApolloSrv struct {
	// namespaceName
	namespaceNames  []string
	mutex           sync.RWMutex
	configServerUrl string
	appId           string
	clusterName     string
	lastPullTime    time.Time
	logger          ApolloLogInterface
	pollTask        *PollTask
}

func (this *ApolloSrv) AddYMLConfig(namespace string) {
	this.namespaceNames = append(this.namespaceNames, namespace+"yml")
}

func (this *ApolloSrv) AddConfig(namespace string) {
	this.namespaceNames = append(this.namespaceNames, namespace)
}

func New(log ApolloLogInterface) *ApolloSrv {
	// if namespaceNames == nil || len(namespaceNames) <= 0 {
	// 	panic("namespaceNames 不能为空")
	// }
	once.Do(func() {
		v := new(ApolloSrv)
		v.namespaceNames = []string{}
		v.configServerUrl = apolloConfig.uri
		v.appId = apolloConfig.AppId
		v.clusterName = apolloConfig.Cluster
		v.mutex = sync.RWMutex{}
		v.logger = log
		pollConf := toPollTaskConf(v)
		v.pollTask = NewPollConfig(pollConf, log)
		AplloSrvInstance = v
	})

	return AplloSrvInstance
}

func toPollTaskConf(srv *ApolloSrv) *PollConfig {
	return &PollConfig{
		ConfigServerUrl: srv.configServerUrl,
		AppId:           srv.appId,
		ClusterName:     srv.clusterName,
		NamespaceNames:  srv.namespaceNames,
	}
}

func (srv *ApolloSrv) Start() error {
	
	return srv.pollTask.Start()
}

func (srv *ApolloSrv) StartPoll() {
	srv.pollTask.StartPoll()
}

func (srv *ApolloSrv) SubscribeStart(fn initNotificationHandler) error {
	return srv.pollTask.SubscribeStart(fn)
}

func (srv *ApolloSrv) SubscribeLongPoll(fn notificationHandler) error {
	return srv.pollTask.SubscribeLongPoll(fn)
}

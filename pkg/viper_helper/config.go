package viper_helper

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-sasuke/sasuke/pkg/apollo"
	"github.com/gin-sasuke/sasuke/pkg/file_helper"
	"github.com/gin-sasuke/sasuke/pkg/string_helper"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
)

type Config struct {
	fileName   string
	fileType   FileType
	namespace  string
	sourceType SourceType
	viper      *viper.Viper
}

func (that Config) GetViper() *viper.Viper {
	return that.viper
}

var (
	Configmap             = make(map[string]Config) // 配置map ，每个nameSpace对应一个配置
	namespaceNameInitChan = make(map[string]*apollo.WatchEvent)
	namespaceNamePollChan = make(map[string]chan apollo.WatchEvent)
	// 读取命令行的
	apolloConfigService = ""
	logger              Logg
)

/**
初始化apollo的地址，方便配合flag或者cobra在启动的命令行设置apollo地址
*/
func InitApolloUrl(apolloUrl string) {
	apolloConfigService = apolloUrl
}

func SupportConfigType(fileType string) bool {
	Supported := []string{"properties", "yml"}
	return string_helper.StringInSlice(fileType, Supported)
}

/**
接收初始化的回调
*/
func initHandel(watchEvents []*apollo.WatchEvent) error {
	err := _initHandel(watchEvents)
	if err != nil {
		logger.Error(err.Error())
	}
	return err
}

func _initHandel(watchEvents []*apollo.WatchEvent) error {
	for _, watchEventCp := range watchEvents {
		namespaceNameInitChan[watchEventCp.NamespaceName] = watchEventCp
	}

	for _, watchEvent := range watchEvents {
		// 初始化
		viperInstance := viper.New()
		viperInstance.AddRemoteProvider("consul", watchEvent.NamespaceName, watchEvent.NamespaceName)
		var fileType FileType
		var filePrex string
		if strings.HasSuffix(watchEvent.NamespaceName, ".yml") {
			viperInstance.SetConfigType("yml")
			filePrex = strings.TrimSuffix(watchEvent.NamespaceName, ".yml")
			fileType = YML
		} else {
			viperInstance.SetConfigType("json")
			fileType = JSON
			filePrex = watchEvent.NamespaceName
		}
		err := viperInstance.ReadRemoteConfig()
		if err != nil {
			return errors.Wrapf(err, "%s read in config fail", watchEvent.NamespaceName)
		}
		logger.Info("create config " + watchEvent.NamespaceName)
		Configmap[filePrex] = Config{
			fileName:   watchEvent.NamespaceName,
			fileType:   fileType,
			namespace:  filePrex,
			sourceType: REMOTE_APOLLO,
			viper:      viperInstance,
		}
		go func() {
			viperInstance.WatchRemoteConfigOnChannel()
		}()
	}

	return nil
}

/**
长轮训的事件watch
*/
func poolHandle(watchEvent apollo.WatchEvent) error {
	namespaceNamePollChan[watchEvent.NamespaceName] <- watchEvent
	return nil
}

/**
变量逃逸
*/
func InitLocalConfig(configPath string) error {
	// 解析目录文件
	isDir, err := file_helper.IsDirOrFile(configPath)
	if err != nil {
		return err
	}
	if !isDir {
		return errors.New("非文件目录")
	}
	walkErr := filepath.Walk(configPath, func(path string, info os.FileInfo, err error) error {
		if configPath != path && !info.IsDir() {
			filePrex, fileExt := file_helper.FileBaseAndExt(path)
			if _, ok := Configmap[filePrex]; ok {
				// 有重复定义抛异常
				return errors.New(fmt.Sprintf("%s 重复定义", filePrex))
			}
			// 截取fileExt
			if fileExt == "" {
				return errors.New(fmt.Sprintf("%s 未找到后缀", path))
			}
			fileExtSpil := fileExt[1:]
			if !SupportConfigType(fileExtSpil) {
				return errors.New(fmt.Sprintf("%s 不支持的格式，support properties, yml", path))
			}

			viperInstance := viper.New()
			viperInstance.AddConfigPath(configPath)
			viperInstance.SetConfigName(filePrex)
			viperInstance.SetConfigType(fileExtSpil)
			err := viperInstance.ReadInConfig()
			if err != nil {
				return errors.Wrapf(err, "%s read in config fail", path)
			}

			Configmap[filePrex] = Config{
				fileName:   filePrex + fileExt,
				fileType:   ToFileType(fileExtSpil),
				namespace:  filePrex,
				sourceType: LOCAL_FILE,
				viper:      viperInstance,
			}
		}
		return nil
	})

	if walkErr != nil {
		return errors.WithMessage(walkErr, "本地文件遍历")
	}

	// 如果和本地文件冲突的话，会进行覆盖
	return nil
}

func StartApollo(srv *apollo.ApolloSrv) error {
	//srv := apollo.New(Logg{})
	//srv.SubscribeStart(initHandel)
	//srv.SubscribeLongPoll(poolHandle)

	startErr := srv.Start()
	if startErr != nil {
		return errors.WithMessage(startErr, "apollo启动失败")
	}
	return nil
}

func RemoveReplicaSliceString(slc []string) []string {
	/*
	   slice(string类型)元素去重
	*/
	result := make([]string, 0)
	tempMap := make(map[string]bool, len(slc))
	for _, e := range slc {
		if tempMap[e] == false {
			tempMap[e] = true
			result = append(result, e)
		}
	}
	return result
}

package viper_helper

import (
	"fmt"
	"github.com/gin-sasuke/sasuke/pkg/apollo"
	"github.com/gin-sasuke/sasuke/pkg/file_helper"
	"github.com/pkg/errors"
	"github.com/spf13/viper"

	"os"
	"path/filepath"
	"strings"
	"time"
)

type Config struct {
	fileName string
	fileType FileType
	namespace string
	sourceType SourceType
	viper *viper.Viper
}

func(that Config) GetViper() *viper.Viper {
	return that.viper
}

var (
	Configmap map[string]Config = make(map[string]Config) // 配置map ，每个nameSpace对应一个配置
	namespaceNameInitChan = make(map[string]watchEventResp)
	namespaceNamePollChan = make(map[string]chan apollo.WatchEvent)
)

type watchEventResp struct {
	WatchEventChan chan apollo.WatchEvent
	Bytes []byte
}

type Logg struct {

}

func (l Logg) Debug(format string) {
	format = time.Now().Format("2006-01-02 15:04:05.000") + "    " + format
	fmt.Println(format)
}

func (l Logg) Info(format string) {
	format = time.Now().Format("2006-01-02 15:04:05.000") + "    " + format
	fmt.Println(format)
}

func (l Logg) Warn(format string) {
	format = time.Now().Format("2006-01-02 15:04:05.000") + "    " + format
	fmt.Println(format)
}

func (l Logg) Error(format string) {
	format = time.Now().Format("2006-01-02 15:04:05.000") + "    " + format
	fmt.Println(format)
}

/**
 	 变量逃逸
 */
func InitLocalConfig(configPath string) error {
	// 解析目录文件
	isDir,err := file_helper.IsDirOrFile(configPath)
	if err != nil {
		return err
	}
	if !isDir {
		return errors.New("非文件目录")
	}
	filepath.Walk(configPath, func(path string, info os.FileInfo, err error) error {
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

			viperInstance := viper.New()
			viperInstance.AddConfigPath(configPath)
			viperInstance.SetConfigName(filePrex)
			viperInstance.SetConfigType(fileExtSpil)
			err := viperInstance.ReadInConfig()
			if err != nil {
				return errors.Wrapf(err, "%s read in config fail", path)
			}
			Configmap[filePrex] = Config{
				fileName:   filePrex+fileExt,
				fileType:  ToFileType(fileExtSpil),
				namespace:  filePrex,
				sourceType: LOCAL_FILE,
				viper:      viperInstance,
			}
		}
		return nil
	})

	//for _, v := range Configmap {
	//	fmt.Println(v.viper.AllSettings())
	//}

	// 开始加载远程配置 (apollo)
	if cf,ok := Configmap["config"]; ok {
		enableApollo := cf.viper.GetBool("viper.remoteprovider.apollo.enable")
		if enableApollo {
			ConfigServerUrl := cf.viper.GetString("viper.remoteprovider.apollo.configService")
			AppId := cf.viper.GetString("viper.remoteprovider.apollo.appid")
			ClusterName := cf.viper.GetString("viper.remoteprovider.apollo.clusterName")
			namespaceNames := cf.viper.GetString("viper.remoteprovider.apollo.namespaceNames")
			namespaceNameSlice := RemoveReplicaSliceString(strings.Split(namespaceNames, ","))
			// 去重切分


			// 定义初始化配置的接受chan
			for _, namespaceName := range namespaceNameSlice {
				namespaceNamePollChan[namespaceName] = make(chan apollo.WatchEvent, 512)
			}

			initHandle := func(watchEvent apollo.WatchEvent) error {
				wteChan := make(chan apollo.WatchEvent, 1)
				wteChan <- watchEvent
				namespaceNameInitChan[watchEvent.NamespaceName] = watchEventResp{
					WatchEventChan: wteChan,
					Bytes: watchEvent.Bytes,
				}
				return nil
			}

			poolHandle := func(watchEvent apollo.WatchEvent) error {
				namespaceNamePollChan[watchEvent.NamespaceName] <- watchEvent
				return nil
			}

			srv := apollo.New(ConfigServerUrl, AppId, namespaceNameSlice, ClusterName, poolHandle, initHandle, Logg{})
			srv.Start()
			viper.RemoteConfig = ApolloRemote{}

			for _, namespaceName := range namespaceNameSlice {
				// 循环初始化配置
				select {
				case event := <-namespaceNameInitChan[namespaceName].WatchEventChan :
					namespaceNameInitChan[event.NamespaceName] = watchEventResp{
						WatchEventChan: nil,
						Bytes:          event.Bytes,
					}
					// 初始化
					viperInstance := viper.New()
					viperInstance.AddRemoteProvider("consul", event.NamespaceName, event.NamespaceName)
					var fileType FileType
					var filePrex string
					if strings.HasSuffix(namespaceName, ".yml") {
						viperInstance.SetConfigType("yml")
						filePrex = strings.TrimSuffix(namespaceName, ".yml")
						fileType = YML
					} else {
						viperInstance.SetConfigType("json")
						fileType = JSON
						filePrex = namespaceName
					}
					err := viperInstance.ReadRemoteConfig()
					if err != nil {
						return errors.Wrapf(err, "%s read in config fail", namespaceNames)
					}
					Configmap[filePrex] = Config{
						fileName:   namespaceName,
						fileType:  fileType,
						namespace:  filePrex,
						sourceType: REMOTE_APOLLO,
						viper:      viperInstance,
					}
					go func() {
						viperInstance.WatchRemoteConfigOnChannel()
					}()

				}
			}

			// 开始长轮训
			srv.StartPoll()
		}
	}
	// 如果和本地文件冲突的话，会进行覆盖

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


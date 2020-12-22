package apollo

import (
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"io/ioutil"
)

type Notification struct {
	NamespaceName string `json:"namespaceName"`
	NotificationId int `json:"notificationId"`
}
type NocacheResponse struct {
	AppId string `json:"appId"`
	Cluster string `json:"cluster"`
	NamespaceName string `json:"namespaceName"`
	Configurations interface{} `json:"configurations"`
	releaseKey string `json:"releaseKey"`
}


// 请求无缓存接口
// URL: {config_server_url}/configs/{appId}/{clusterName}/{namespaceName}?releaseKey={releaseKey}&ip={clientIp}
func NocacheGet(pollTask *PollTask, nameSpaceName string, notificationId int, isInit bool) error {
	requestURL := fmt.Sprintf("%s/configs/%s/%s/%s", pollTask.Config.ConfigServerUrl, pollTask.Config.AppId, pollTask.Config.ClusterName, nameSpaceName)
	response, error := pollTask.HttpRequest.Request(requestURL)
	if error != nil {
		return errors.WithMessage(error, "notifications 请求notifications失败")
	}
	// 判断http状态
	if response.StatusCode == 200 {
		body,err := ioutil.ReadAll(response.Body)
		if err != nil {
			return errors.WithMessage(err, "notifications 请求notifications失败")
		}
		nocacheResp := NocacheResponse{}
		if unmarshalErr := json.Unmarshal(body, &nocacheResp); unmarshalErr != nil {
			return errors.WithMessagef(unmarshalErr, "获取无缓存接口 返回格式错误：%s", string(body))
		}
		byteSlice,marshalErr := json.Marshal(nocacheResp.Configurations);
		if marshalErr != nil {
			return errors.WithMessagef(err, "获取无缓存接口 返回格式错误：%s", string(body))
		}
		watchEvent := WatchEvent{
			NamespaceName: nameSpaceName,
			Bytes: byteSlice,
		}
		if isInit {
			pollTask.initHandler(watchEvent)
		} else {
			pollTask.handler(watchEvent)
		}
		pollTask.namespaceNames[nameSpaceName] = notificationId
	}
	return nil
}

// 请求nameSpace是否有变化
// URL: {config_server_url}/notifications/v2?appId={appId}&cluster={clusterName}&notifications={notifications}
func NotificationsGet(pollTask *PollTask, isInit bool) error {
	requestURL := fmt.Sprintf("%s/notifications/v2?appId=%s&cluster=%s&notifications=", pollTask.Config.ConfigServerUrl, pollTask.Config.AppId, pollTask.Config.ClusterName)
	nameSpaceSlice := make([]Notification, 0)
	pollTask.Mutex.Lock()
	defer pollTask.Mutex.Unlock()
	for key, value := range pollTask.namespaceNames {
		nameSpaceSlice = append(nameSpaceSlice, Notification{
			NamespaceName:  key,
			NotificationId: value,
		})
	}
	nameSpaceSliceJson, err := json.Marshal(nameSpaceSlice)
	if err != nil {
		return errors.WithMessage(err, "notifications 请求设置本地notifications信息 json序列化失败")
	}
	requestURL = requestURL + string(nameSpaceSliceJson)
	response, error := pollTask.HttpRequest.Request(requestURL)
	if error != nil {
		pollTask.logger.Error(error.Error())
		return errors.WithMessage(error, "notifications 请求notifications失败")
	}
	if response.StatusCode == 304 {
		pollTask.logger.Info("notifications请求304")
	}
	// 判断http状态
	if response.StatusCode == 200 {
		body,err := ioutil.ReadAll(response.Body)
		if err != nil {
			return errors.WithMessage(err, "notifications 请求notifications失败")
		}
		var rtNameSpaceSlice []*Notification

		if unmarshalErr := json.Unmarshal(body, &rtNameSpaceSlice); unmarshalErr != nil {
			return errors.WithMessagef(unmarshalErr, "notifications 返回格式错误：%s", string(body))
		}

		// 处理配置变化的方法
		// 批量发出更新事件
		pollTask.logger.Info("获取最新配置")
		for _, value := range rtNameSpaceSlice {
			errGet := NocacheGet(pollTask, value.NamespaceName, value.NotificationId, isInit)
			if errGet != nil {
				pollTask.logger.Error(errGet.Error())
			}
		}

	}
	return nil
}




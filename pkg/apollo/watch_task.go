package apollo

import (
	"bytes"
	"fmt"
	"io/ioutil"
)

type WatchEvent struct {
	NamespaceName string
	Bytes []byte
}

// 注册的回调函数
type notificationHandler func(watchEvent WatchEvent) error
type initNotificationHandler func(watchEvent []*WatchEvent) error

func DefaultNotificationHandler (watchEvent WatchEvent) error {
	fmt.Println(watchEvent.NamespaceName)
	reader := bytes.NewReader(watchEvent.Bytes)
	readRs, _ := ioutil.ReadAll(reader)
	fmt.Println(string(readRs))
	return nil
}



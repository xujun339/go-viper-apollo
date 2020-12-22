package viper_helper

import (
	"bytes"
	"encoding/json"
	"github.com/spf13/viper"
	"io"
	"strings"
)

// viper的apollo拓展

type ApolloRemote struct {

}

func (a ApolloRemote) Get(rp viper.RemoteProvider) (io.Reader, error) {
	byteSlice := namespaceNameInitChan[rp.Path()].Bytes
	var r io.Reader
	configBytes, err := convertByte(byteSlice, rp.Path())
	if err != nil {
		return nil, err
	}
	r = bytes.NewReader(configBytes)
	return r, nil
}

func (a ApolloRemote) Watch(rp viper.RemoteProvider) (io.Reader, error) {
	return nil, nil
}

func (a ApolloRemote) WatchChannel(rp viper.RemoteProvider) (<-chan *viper.RemoteResponse, chan bool) {
	quit := make(chan bool)
	viperResponseCh := make(chan *viper.RemoteResponse)

	go func(vc chan<- *viper.RemoteResponse, quit <-chan bool) {
		for {
			select {
			case <-quit:
				return
			case event := <-namespaceNamePollChan[rp.Path()] :
				byteSlice := event.Bytes
				configBytes, err := convertByte(byteSlice, rp.Path())
				// check 配置信息是否有变化
				vc <- &viper.RemoteResponse{Value: configBytes, Error: err}
			}
		}
	}(viperResponseCh, quit)
	return viperResponseCh, quit
}

func deepSearch(m map[string]interface{}, path []string) map[string]interface{} {
	for _, k := range path {
		m2, ok := m[k]
		if !ok {
			// intermediate key does not exist
			// => create it and continue from there
			m3 := make(map[string]interface{})
			m[k] = m3
			m = m3
			continue
		}
		m3, ok := m2.(map[string]interface{})
		if !ok {
			// intermediate key is a value
			// => replace with a new map
			m3 = make(map[string]interface{})
			m[k] = m3
		}
		// continue search from here
		m = m3
	}
	return m
}

func convertByte(byteSlice []byte, namespace string) ([]byte, error) {
	var r []byte
	if len(strings.Split(namespace, ".")) == 2 {
		// 说明是yml
		var mapStyle map[string]interface{}
		json.Unmarshal(byteSlice, &mapStyle)
		r = []byte(mapStyle["content"].(string))
	} else {
		// 普通properities格式进行转换
		var mapStyle map[string]interface{}
		var convert = make(map[string]interface{})
		json.Unmarshal(byteSlice, &mapStyle)

		for key, value := range mapStyle {
			// recursively build nested maps
			path := strings.Split(key, ".")
			lastKey := strings.ToLower(path[len(path)-1])
			deepestMap := deepSearch(convert, path[0:len(path)-1])
			// set innermost value
			deepestMap[lastKey] = value
		}

		bb,err := json.Marshal(convert)
		if err != nil {
			return nil, err
		}
		r = bb
	}
	return r ,nil
}
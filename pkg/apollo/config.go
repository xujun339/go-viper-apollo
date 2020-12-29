package apollo

import (
	"errors"
	"log"
	"strings"

	"github.com/spf13/viper"
)

type ConfigFileFormat int32

const (
	Properties ConfigFileFormat = iota
	Xml
	Json
	Yml
	Yaml
	Txt
)

func (c ConfigFileFormat) String() string {
	switch c {
	case Properties:
		return "properties"
	case Xml:
		return "xml"
	case Json:
		return "json"
	case Yml:
		return "yml"
	case Yaml:
		return "yaml"
	case Txt:
		return "txt"
	default:
		return ""
	}
}

type ApolloConfig struct {
	AppId      string            `json:"AppId"`
	Cluster    string            `json:"Cluster"`
	Env        string            `json:"Env"`
	MetaServer map[string]string `json:"MetaServer"`
	uri        string
}

func (this *ApolloConfig) getUri() (string, error) {
	if v, ok := this.MetaServer[this.Env]; ok {
		return v, nil
	}
	return "", errors.New("not found uri")
}

func (this *ApolloConfig) verify() error {
	if this.AppId == "" {
		return errors.New("appid is required")
	}
	if this.Cluster == "" {
		this.Cluster = "default"
	}
	if this.Env == "" {
		return errors.New("env is required")
	}
	switch strings.ToUpper(this.Env) {
	case "DEV", "FAT", "UAT", "PRO":
		break
	default:
		return errors.New("env error")
	}
	if len(this.MetaServer) == 0 {
		return errors.New("metaServer is required")
	}
	var err error
	this.uri, err = this.getUri()
	return err
}

func InitViperConfig(path string, configType string, configName string) *ApolloConfig {
	apolloConfig := &ApolloConfig{}
	vip := viper.New()
	vip.AddConfigPath(path)
	vip.SetConfigName(configName)
	vip.SetConfigType(configType)
	if err := vip.Unmarshal(&apolloConfig); err != nil {
		log.Fatal(err)
	}
	if err := apolloConfig.verify(); err != nil {
		log.Fatal(err)
	}

	return apolloConfig
}

# viper + apollo 配置模块，支持配置拉取更新
## 设置配置文件路径，会扫描出配置文件，以文件名作为namespace
`比如application.yml namespace=application`

## 使用方式
``初始化配置viper_helper.InitLocalConfig("config")``
### config目录下的config.properties是配置的根配置，会去判断是否需要启动apollo 
``viper.remoteprovider.apollo.enable = true``
``获取配置viper_helper.Configmap["logging-level"].GetViper().AllSettings()``

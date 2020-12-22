package apollo

// 定义log的接口

type ApolloLogInterface interface {
	Debug(format string)
	Info(format string)
	Warn(format string)
	Error(format string)
}

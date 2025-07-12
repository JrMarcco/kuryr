package main

import (
	"github.com/JrMarcco/kuryr/internal/ioc"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"
	"go.uber.org/zap"
)

func main() {
	initViper()

	fx.New(
		// 初始化 zap.Logger
		ioc.LoggerFxOpt,

		// 初始化 grpc
		ioc.GrpcFxOpt,

		// 初始化 ioc.App
		ioc.AppFxOpt,

		fx.WithLogger(func(logger *zap.Logger) fxevent.Logger {
			return &fxevent.ZapLogger{Logger: logger}
		}),

		// 注册 zap logger lifecycle 确保日志缓冲区被刷新
		ioc.LoggerFxInvoke,
		// 注册 app lifecycle
		ioc.AppFxInvoke,
	).Run()
}

// initViper 初始化 viper
func initViper() {
	configFile := pflag.String("config", "etc/config.yaml", "配置文件路径")
	pflag.Parse()

	viper.SetConfigFile(*configFile)
	viper.SetConfigType("yaml")
	if err := viper.ReadInConfig(); err != nil {
		panic(err)
	}
}

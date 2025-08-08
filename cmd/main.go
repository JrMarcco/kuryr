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
		fx.WithLogger(func(logger *zap.Logger) fxevent.Logger {
			return &fxevent.ZapLogger{Logger: logger}
		}),

		// 初始化 zap.Logger
		ioc.LoggerFxOpt,
		// 初始化 id 生成器
		ioc.IdGeneratorFxOpt,
		// 初始化 go cache
		ioc.GoCacheFxOpt,
		// 初始化 redis
		ioc.RedisFxOpt,
		// 初始化 db
		ioc.DBFxOpt,
		// 初始化 etcd
		ioc.EtcdFxOpt,
		// 初始化 grpc registry
		ioc.RegistryFxOpt,
		// 初始化 repository
		ioc.RepoFxOpt,
		// 初始化 service
		ioc.ServiceFxOpt,
		// 初始化 grpc
		ioc.GrpcFxOpt,
		// 初始化 ioc.App
		ioc.AppFxOpt,

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

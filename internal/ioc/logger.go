package ioc

import (
	"context"

	"github.com/spf13/viper"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

var (
	LoggerFxOpt    = fx.Provide(InitLogger)
	LoggerFxInvoke = fx.Invoke(LoggerLifecycle)
)

func InitLogger() *zap.Logger {
	type config struct {
		Env string `mapstructure:"env"`
	}

	cfg := config{}
	if err := viper.Unmarshal(&cfg); err != nil {
		panic(err)
	}

	var zapLogger *zap.Logger
	var err error
	switch cfg.Env {
	case "prod":
		zapLogger, err = zap.NewProduction()
	default:
		zapLogger, err = zap.NewDevelopment()
	}
	if err != nil {
		panic(err)
	}
	return zapLogger
}

func LoggerLifecycle(lc fx.Lifecycle, logger *zap.Logger) {
	lc.Append(fx.Hook{
		// 程序停止时 flush buffer 防止日志丢失
		OnStop: func(ctx context.Context) error {
			_ = logger.Sync()
			return nil
		},
	})
}

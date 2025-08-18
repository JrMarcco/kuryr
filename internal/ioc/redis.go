package ioc

import (
	"github.com/JrMarcco/dlock"
	dredis "github.com/JrMarcco/dlock/redis"
	"github.com/redis/go-redis/v9"
	"github.com/spf13/viper"
	"go.uber.org/fx"
)

var RedisFxOpt = fx.Module(
	"redis",
	fx.Provide(
		InitRedis,
		InitDClient,
	),
)

func InitRedis() redis.Cmdable {
	type config struct {
		Addr     string `mapstructure:"addr"`
		Password string `mapstructure:"password"`
	}

	cfg := config{}
	if err := viper.UnmarshalKey("redis", &cfg); err != nil {
		panic(err)
	}

	return redis.NewClient(&redis.Options{
		Addr:     cfg.Addr,
		Password: cfg.Password,
	})
}

func InitDClient(rc redis.Cmdable) dlock.Dclient {
	return dredis.NewDClientBuilder(rc).Build()
}

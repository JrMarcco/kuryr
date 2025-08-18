package ioc

import (
	"time"

	"github.com/patrickmn/go-cache"
	"github.com/spf13/viper"
	"go.uber.org/fx"
)

var GoCacheFxOpt = fx.Module("go-cache", fx.Provide(InitGoCache))

func InitGoCache() *cache.Cache {
	type config struct {
		DefaultExpiration time.Duration `mapstructure:"default_expiration"`
		CleanupInterval   time.Duration `mapstructure:"cleanup_interval"`
	}

	cfg := config{}
	if err := viper.UnmarshalKey("go_cache", &cfg); err != nil {
		panic(err)
	}

	return cache.New(cfg.DefaultExpiration, cfg.CleanupInterval)
}

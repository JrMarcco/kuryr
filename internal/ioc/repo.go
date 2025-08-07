package ioc

import (
	"github.com/JrMarcco/kuryr/internal/repository"
	"github.com/JrMarcco/kuryr/internal/repository/cache"
	"github.com/JrMarcco/kuryr/internal/repository/cache/local"
	"github.com/JrMarcco/kuryr/internal/repository/cache/redis"
	"github.com/JrMarcco/kuryr/internal/repository/dao"
	"github.com/spf13/viper"
	"go.uber.org/fx"
	"gorm.io/gorm"
)

var RepoFxOpt = fx.Options(
	// dao
	fx.Provide(
		// biz info dao
		fx.Annotate(
			dao.NewDefaultBizInfoDao,
			fx.As(new(dao.BizInfoDao)),
		),

		// biz config dao
		fx.Annotate(
			dao.NewDefaultBizConfigDao,
			fx.As(new(dao.BizConfigDao)),
		),

		// provider dao
		fx.Annotate(
			InitProviderDao,
			fx.As(new(dao.ProviderDao)),
		),

		// channel template dao
		fx.Annotate(
			dao.NewDefaultChannelTplDao,
			fx.As(new(dao.ChannelTplDao)),
		),
	),

	// cache
	fx.Provide(
		// local cache
		fx.Annotate(
			local.NewLBizConfigCache,
			fx.As(new(cache.BizConfigCache)),
			fx.ResultTags(`name:"local_biz_config_cache"`),
		),

		// redis cache
		fx.Annotate(
			redis.NewRBizConfigCache,
			fx.As(new(cache.BizConfigCache)),
			fx.ResultTags(`name:"redis_biz_config_cache"`),
		),
	),

	// repo
	fx.Provide(
		// biz info repo
		fx.Annotate(
			repository.NewDefaultBizInfoRepo,
			fx.As(new(repository.BizInfoRepo)),
		),

		// biz config repo
		fx.Annotate(
			repository.NewDefaultBizConfigRepo,
			fx.As(new(repository.BizConfigRepo)),
			fx.ParamTags(``, `name:"local_biz_config_cache"`, `name:"redis_biz_config_cache"`, ``),
		),

		// provider repo
		fx.Annotate(
			repository.NewDefaultProviderRepo,
			fx.As(new(repository.ProviderRepo)),
		),

		// channel template repo
		fx.Annotate(
			repository.NewDefaultChannelTplRepo,
			fx.As(new(repository.ChannelTplRepo)),
		),
	),
)

func InitProviderDao(db *gorm.DB) *dao.DefaultProviderDao {
	var encryptKey string
	if err := viper.UnmarshalKey("provider.encrypt_key", &encryptKey); err != nil {
		panic(err)
	}
	return dao.NewDefaultProviderDao(db, encryptKey)
}

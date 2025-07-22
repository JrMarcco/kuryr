package ioc

import (
	"github.com/JrMarcco/kuryr/internal/repository"
	"github.com/JrMarcco/kuryr/internal/repository/cache"
	"github.com/JrMarcco/kuryr/internal/repository/cache/local"
	"github.com/JrMarcco/kuryr/internal/repository/cache/redis"
	"github.com/JrMarcco/kuryr/internal/repository/dao"
	"go.uber.org/fx"
)

var RepoFxOpt = fx.Options(
	// dao
	fx.Provide(
		// biz config dao
		fx.Annotate(
			dao.NewDefaultBizConfigDao,
			fx.As(new(dao.BizConfigDao)),
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
		// biz config repo
		fx.Annotate(
			repository.NewDefaultBizConfigRepo,
			fx.As(new(repository.BizConfigRepo)),
			fx.ParamTags(``, `name:"local_biz_config_cache"`, `name:"redis_biz_config_cache"`, ``),
		),
	),
)

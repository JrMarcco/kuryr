package ioc

import (
	"github.com/JrMarcco/kuryr/internal/service/bizconf"
	"github.com/JrMarcco/kuryr/internal/service/bizinfo"
	"github.com/JrMarcco/kuryr/internal/service/callback"
	"github.com/JrMarcco/kuryr/internal/service/provider"
	"go.uber.org/fx"
)

var ServiceFxOpt = fx.Options(
	fx.Provide(
		// biz info service
		fx.Annotate(
			bizinfo.NewDefaultService,
			fx.As(new(bizinfo.Service)),
		),

		// biz config service
		fx.Annotate(
			bizconf.NewDefaultService,
			fx.As(new(bizconf.Service)),
		),

		// provider service
		fx.Annotate(
			provider.NewDefaultService,
			fx.As(new(provider.Service)),
		),

		// callback service
		fx.Annotate(
			callback.NewDefaultService,
			fx.As(new(callback.Service)),
		),
	),
)

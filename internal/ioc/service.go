package ioc

import (
	"github.com/JrMarcco/kuryr/internal/service/bizconf"
	"github.com/JrMarcco/kuryr/internal/service/provider"
	"go.uber.org/fx"
)

var ServiceFxOpt = fx.Options(
	fx.Provide(
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
	),
)

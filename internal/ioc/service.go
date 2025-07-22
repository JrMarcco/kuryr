package ioc

import (
	"github.com/JrMarcco/kuryr/internal/service/bizconf"
	"go.uber.org/fx"
)

var ServiceFxOpt = fx.Options(
	fx.Provide(
		// biz config service
		fx.Annotate(
			bizconf.NewDefaultBizConfigService,
			fx.As(new(bizconf.BizConfigService)),
		),
	),
)

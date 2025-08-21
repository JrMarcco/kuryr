package ioc

import (
	"github.com/JrMarcco/kuryr/internal/pkg/secret"
	"github.com/JrMarcco/kuryr/internal/pkg/secret/base64"
	"github.com/JrMarcco/kuryr/internal/service/bizconf"
	"github.com/JrMarcco/kuryr/internal/service/bizinfo"
	"github.com/JrMarcco/kuryr/internal/service/callback"
	"github.com/JrMarcco/kuryr/internal/service/provider"
	"github.com/JrMarcco/kuryr/internal/service/sendstrategy"
	"go.uber.org/fx"
)

var ServiceFxOpt = fx.Module(
	"service",
	fx.Provide(
		// user service
		fx.Annotate(
			base64.NewGenerator,
			fx.As(new(secret.Generator)),
		),

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

		// default send strategy
		fx.Annotate(
			sendstrategy.NewDefaultSendStrategy,
			fx.As(new(sendstrategy.SendStrategy)),
			fx.ResultTags(`name:"default_send_strategy"`),
		),

		// immediate send strategy
		fx.Annotate(
			sendstrategy.NewImmediateSendStrategy,
			fx.As(new(sendstrategy.SendStrategy)),
			fx.ResultTags(`name:"immediate_send_strategy"`),
		),

		// send strategy dispatcher
		fx.Annotate(
			sendstrategy.NewDispatcher,
			fx.As(new(sendstrategy.SendStrategy)),
			fx.ParamTags(`name:"default_send_strategy"`, `name:"immediate_send_strategy"`),
			fx.ResultTags(`name:"send_strategy_dispatcher"`),
		),
	),
)

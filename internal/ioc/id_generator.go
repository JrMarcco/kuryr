package ioc

import (
	"github.com/JrMarcco/kuryr/internal/pkg/idgen"
	"github.com/JrMarcco/kuryr/internal/pkg/idgen/snowflake"
	"go.uber.org/fx"
)

var IdGeneratorFxOpt = fx.Module(
	"id-generator",
	fx.Provide(
		fx.Annotate(
			snowflake.NewGenerator,
			fx.As(new(idgen.Generator)),
		),
	),
)

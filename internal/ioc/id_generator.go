package ioc

import (
	"github.com/JrMarcco/kuryr/internal/pkg/idgen"
	"github.com/JrMarcco/kuryr/internal/pkg/idgen/snowflake"
	"go.uber.org/fx"
)

var IdGeneratorFxOpt = fx.Provide(
	fx.Annotate(
		snowflake.NewGenerator,
		fx.As(new(idgen.Generator)),
	),
)

package ioc

import (
	"sync"

	"github.com/JrMarcco/easy-kit/xsync"
	"github.com/JrMarcco/kuryr/internal/pkg/sharding"
	"github.com/spf13/viper"
	"go.uber.org/fx"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DBFxOpt = fx.Provide(
	InitBaseDB,
	InitShardingDB,
	fx.Annotate(
		InitCblShardingStrategy,
		fx.As(new(sharding.Strategy)),
		fx.ResultTags(`name:"cbl_sharding_strategy"`),
	),
)

var (
	mu   sync.Mutex
	once sync.Once
)

// InitBaseDB 初始化基础 db ( kuryr )
func InitBaseDB() *gorm.DB {
	type config struct {
		DSN string `mapstructure:"dsn"`
	}
	cfg := config{}
	if err := viper.UnmarshalKey("db.base", &cfg); err != nil {
		panic(err)
	}

	db, err := gorm.Open(postgres.Open(cfg.DSN), &gorm.Config{})
	if err != nil {
		panic(err)
	}
	return db
}

func InitShardingDB() *xsync.Map[string, *gorm.DB] {
	type config struct {
		DSN string `mapstructure:"dsn"`
	}

	type allConfig map[string]config
	allCfg := make(allConfig)
	if err := viper.UnmarshalKey("db.sharding", &allCfg); err != nil {
		panic(err)
	}

	mu.Lock()
	defer mu.Unlock()

	var dbs xsync.Map[string, *gorm.DB]
	once.Do(func() {
		for key, cfg := range allCfg {
			db, err := gorm.Open(postgres.Open(cfg.DSN), &gorm.Config{})
			if err != nil {
				panic(err)
			}
			dbs.Store(key, db)
		}
	})

	return &dbs
}

// InitCblShardingStrategy 初始化 callback log 分库分表策略
func InitCblShardingStrategy() *sharding.HashSharding {
	return sharding.NewHashSharding(
		"kuryr", "callback_log", 2, 4,
	)
}

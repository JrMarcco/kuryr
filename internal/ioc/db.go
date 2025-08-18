package ioc

import (
	"sync"
	"time"

	"github.com/JrMarcco/easy-kit/xsync"
	pkggorm "github.com/JrMarcco/kuryr/internal/pkg/gorm"
	"github.com/JrMarcco/kuryr/internal/pkg/sharding"
	"github.com/spf13/viper"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DBFxOpt = fx.Module(
	"db",
	fx.Provide(
		InitBaseDB,
		InitShardingDB,
		fx.Annotate(
			InitCallbackLogSharding,
			fx.As(new(sharding.Strategy)),
			fx.ResultTags(`name:"cbl_sharding_strategy"`),
		),
	),
)

var (
	mu   sync.Mutex
	once sync.Once
)

// InitBaseDB 初始化基础 db ( kuryr )
func InitBaseDB(zLogger *zap.Logger) *gorm.DB {
	type baseConfig struct {
		DSN string `mapstructure:"dsn"`
	}

	type config struct {
		LogLevel                  string     `mapstructure:"log_level"`
		SlowThreshold             int        `mapstructure:"slow_threshold"`
		IgnoreRecordNotFoundError bool       `mapstructure:"ignore_record_not_found_error"`
		Base                      baseConfig `mapstructure:"base"`
	}
	cfg := config{}
	if err := viper.UnmarshalKey("db", &cfg); err != nil {
		panic(err)
	}

	var level logger.LogLevel
	switch cfg.LogLevel {
	case "silent":
		level = logger.Silent
	case "error":
		level = logger.Error
	case "warn":
		level = logger.Warn
	case "info":
		level = logger.Info
	default:
		panic("invalid logger level")
	}

	db, err := gorm.Open(postgres.Open(cfg.Base.DSN), &gorm.Config{
		Logger: pkggorm.NewZapLogger(
			zLogger,
			pkggorm.WithLogLevel(level),
			pkggorm.WithSlowThreshold(time.Duration(cfg.SlowThreshold)*time.Millisecond),
			pkggorm.WithIgnoreRecordNotFoundError(cfg.IgnoreRecordNotFoundError),
		),
	})
	if err != nil {
		panic(err)
	}
	return db
}

func InitShardingDB(zLogger *zap.Logger) *xsync.Map[string, *gorm.DB] {
	type shardingConfig struct {
		Name string `mapstructure:"name"`
		DSN  string `mapstructure:"dsn"`
	}

	type config struct {
		LogLevel                  string           `mapstructure:"log_level"`
		SlowThreshold             int              `mapstructure:"slow_threshold"`
		IgnoreRecordNotFoundError bool             `mapstructure:"ignore_record_not_found_error"`
		Sharding                  []shardingConfig `mapstructure:"sharding"`
	}

	cfg := config{}
	if err := viper.UnmarshalKey("db", &cfg); err != nil {
		panic(err)
	}

	var level logger.LogLevel
	switch cfg.LogLevel {
	case "silent":
		level = logger.Silent
	case "error":
		level = logger.Error
	case "warn":
		level = logger.Warn
	case "info":
		level = logger.Info
	default:
		panic("invalid log level")
	}

	mu.Lock()
	defer mu.Unlock()

	var dbs xsync.Map[string, *gorm.DB]
	once.Do(func() {
		for _, s := range cfg.Sharding {
			db, err := gorm.Open(postgres.Open(s.DSN), &gorm.Config{
				Logger: pkggorm.NewZapLogger(
					zLogger,
					pkggorm.WithLogLevel(level),
					pkggorm.WithSlowThreshold(time.Duration(cfg.SlowThreshold)*time.Millisecond),
					pkggorm.WithIgnoreRecordNotFoundError(cfg.IgnoreRecordNotFoundError),
				),
			})
			if err != nil {
				panic(err)
			}
			dbs.Store(s.Name, db)
		}
	})

	return &dbs
}

// InitCallbackLogSharding 初始化 callback log 分库分表策略。
func InitCallbackLogSharding() *sharding.BalancedSharding {
	type config struct {
		DBPrefix        string `mapstructure:"db_prefix"`
		TablePrefix     string `mapstructure:"table_prefix"`
		DBShardCount    uint64 `mapstructure:"db_shard_count"`
		TableShardCount uint64 `mapstructure:"table_shard_count"`
		BroadcastMode   string `mapstructure:"broadcast_mode"`
	}

	cfg := config{}
	if err := viper.UnmarshalKey("sharding.callback_log", &cfg); err != nil {
		panic(err)
	}

	base := sharding.NewHashSharding(
		cfg.DBPrefix,
		cfg.TablePrefix,
		cfg.DBShardCount,
		cfg.TableShardCount,
	)

	return sharding.NewBalancedSharding(base, sharding.BroadcastMode(cfg.BroadcastMode))
}

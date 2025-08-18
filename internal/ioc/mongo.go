package ioc

import (
	"context"
	"time"

	"github.com/spf13/viper"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.mongodb.org/mongo-driver/v2/mongo/readpref"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

var MongoFxOpt = fx.Module("mongo", fx.Provide(InitMongoClient))

func InitMongoClient(lc fx.Lifecycle, logger *zap.Logger) *mongo.Client {
	type config struct {
		Uri                    string        `mapstructure:"uri"`
		AppName                string        `mapstructure:"app_name"`
		Database               string        `mapstructure:"database"`
		MaxPoolSize            uint64        `mapstructure:"max_pool_size"`
		MaxConnIdleTime        time.Duration `mapstructure:"max_conn_idle_time"`
		ConnectTimeout         time.Duration `mapstructure:"connect_timeout"`
		StartupTimeout         time.Duration `mapstructure:"startup_timeout"`
		ShutdownTimeout        time.Duration `mapstructure:"shutdown_timeout"`
		ServerSelectionTimeout time.Duration `mapstructure:"server_selection_timeout"`
	}

	cfg := config{}
	if err := viper.UnmarshalKey("mongo", &cfg); err != nil {
		panic(err)
	}

	clientOptions := options.Client().
		ApplyURI(cfg.Uri).
		SetAppName(cfg.AppName).
		SetMaxPoolSize(cfg.MaxPoolSize).
		SetMaxConnIdleTime(cfg.MaxConnIdleTime).
		SetConnectTimeout(cfg.ConnectTimeout).
		SetServerSelectionTimeout(cfg.ServerSelectionTimeout)

	client, err := mongo.Connect(clientOptions)
	if err != nil {
		panic(err)
	}

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			startupCtx, cancel := context.WithTimeout(ctx, cfg.StartupTimeout)
			defer cancel()

			err := client.Ping(startupCtx, nil)
			if err != nil {
				return err
			}

			pingCtx, pingCancel := context.WithTimeout(ctx, cfg.ServerSelectionTimeout)
			defer pingCancel()

			if err := client.Ping(pingCtx, readpref.Primary()); err != nil {
				return err
			}

			logger.Info("[kuryr] successfully connected to mongo", zap.String("uri", cfg.Uri), zap.String("db", cfg.Database))
			return nil
		},
		OnStop: func(ctx context.Context) error {
			stopCtx, cancel := context.WithTimeout(ctx, cfg.ShutdownTimeout)
			defer cancel()

			err := client.Disconnect(stopCtx)
			if err != nil {
				logger.Error("[kuryr] failed to disconnect from mongo", zap.Error(err))
				return err
			}

			logger.Info("[kuryr] successfully disconnected from mongo")
			return nil
		},
	})

	return client
}

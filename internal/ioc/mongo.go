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
		Uri     string `mapstructure:"uri"`
		AppName string `mapstructure:"app_name"`

		AuthSource string `mapstructure:"auth_source"`
		Username   string `mapstructure:"username"`
		Password   string `mapstructure:"password"`

		MaxPoolSize            uint64 `mapstructure:"max_pool_size"`
		MaxConnIdleTime        int    `mapstructure:"max_conn_idle_time"`
		ConnectTimeout         int    `mapstructure:"connect_timeout"`
		ServerSelectionTimeout int    `mapstructure:"server_selection_timeout"`

		StartupTimeout  int `mapstructure:"startup_timeout"`
		ShutdownTimeout int `mapstructure:"shutdown_timeout"`
	}

	cfg := config{}
	if err := viper.UnmarshalKey("mongo", &cfg); err != nil {
		panic(err)
	}

	clientOptions := options.Client().
		ApplyURI(cfg.Uri).
		SetAppName(cfg.AppName).
		SetMaxPoolSize(cfg.MaxPoolSize).
		SetMaxConnIdleTime(time.Duration(cfg.MaxConnIdleTime) * time.Millisecond).
		SetConnectTimeout(time.Duration(cfg.ConnectTimeout) * time.Millisecond).
		SetServerSelectionTimeout(time.Duration(cfg.ServerSelectionTimeout) * time.Millisecond)

	if cfg.Username != "" {
		cred := options.Credential{Username: cfg.Username, Password: cfg.Password}
		if cfg.AuthSource != "" {
			cred.AuthSource = cfg.AuthSource
		}
		clientOptions.SetAuth(cred)
	}

	client, err := mongo.Connect(clientOptions)
	if err != nil {
		panic(err)
	}

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			startupCtx, cancel := context.WithTimeout(ctx, time.Duration(cfg.StartupTimeout)*time.Millisecond)
			defer cancel()

			err := client.Ping(startupCtx, nil)
			if err != nil {
				return err
			}

			pingCtx, pingCancel := context.WithTimeout(ctx, time.Duration(cfg.ServerSelectionTimeout)*time.Millisecond)
			defer pingCancel()

			if err := client.Ping(pingCtx, readpref.Primary()); err != nil {
				return err
			}

			logger.Info("[kuryr] successfully connected to mongo", zap.String("uri", cfg.Uri))
			return nil
		},
		OnStop: func(ctx context.Context) error {
			stopCtx, cancel := context.WithTimeout(ctx, time.Duration(cfg.ShutdownTimeout)*time.Millisecond)
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

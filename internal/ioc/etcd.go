package ioc

import (
	"context"
	"crypto/tls"
	"fmt"
	"time"

	"github.com/spf13/viper"
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

var EtcdFxOpt = fx.Provide(InitEtcdClient)

func InitEtcdClient(logger *zap.Logger, lc fx.Lifecycle) *clientv3.Client {
	type tlsConfig struct {
		Enabled            bool   `mapstructure:"enabled"`
		CertFile           string `mapstructure:"cert_file"`
		KeyFile            string `mapstructure:"key_file"`
		CAFile             string `mapstructure:"ca_file"`
		InsecureSkipVerify bool   `mapstructure:"insecure_skip_verify"`
	}

	type config struct {
		Endpoints []string `mapstructure:"endpoints"`

		Username    string    `mapstructure:"username"`
		Password    string    `mapstructure:"password"`
		DialTimeout int       `mapstructure:"dial_timeout"`
		TLS         tlsConfig `mapstructure:"tls"`
	}
	cfg := config{}
	if err := viper.UnmarshalKey("etcd", &cfg); err != nil {
		panic(err)
	}

	clientCfg := clientv3.Config{
		Endpoints:   cfg.Endpoints,
		Username:    cfg.Username,
		Password:    cfg.Password,
		DialTimeout: time.Duration(cfg.DialTimeout) * time.Millisecond,
	}

	// 配置 tls
	if cfg.TLS.Enabled {
		tlsCfg := &tls.Config{InsecureSkipVerify: cfg.TLS.InsecureSkipVerify}

		if cfg.TLS.CertFile != "" && cfg.TLS.KeyFile != "" {
			cert, err := tls.LoadX509KeyPair(cfg.TLS.CertFile, cfg.TLS.KeyFile)
			if err != nil {
				panic(err)
			}
			tlsCfg.Certificates = []tls.Certificate{cert}
		}
		clientCfg.TLS = tlsCfg
	}

	client, err := clientv3.New(clientCfg)
	if err != nil {
		panic(err)
	}

	// 测试连接
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(cfg.DialTimeout)*time.Millisecond)
	defer cancel()

	_, err = client.Status(ctx, cfg.Endpoints[0])
	if err != nil {
		logger.Error("[kuryr] failed to connect to etcd", zap.Error(err))
		_ = client.Close()
		panic(fmt.Errorf("failed to connect to etcd: %w", err))
	}

	logger.Info("[kuryr] successfully connected to etcd")

	// 注册生命周期 hook，确保客户端正确关闭
	lc.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			logger.Info("[kuryr] etcd client closed")
			return client.Close()
		},
	})
	return client
}

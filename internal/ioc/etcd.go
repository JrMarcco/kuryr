package ioc

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"
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
		tlsCfg := &tls.Config{
			InsecureSkipVerify: cfg.TLS.InsecureSkipVerify,
			MinVersion:         tls.VersionTLS12,
			CipherSuites: []uint16{
				// 支持现代密码套件，包括 Ed25519
				tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
				tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
				tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305,
				tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,
				tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
				tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
			},
		}

		if cfg.TLS.CertFile != "" && cfg.TLS.KeyFile != "" {
			cert, err := tls.LoadX509KeyPair(cfg.TLS.CertFile, cfg.TLS.KeyFile)
			if err != nil {
				logger.Error("[kuryr] failed to load Ed25519 client certificate",
					zap.String("cert_file", cfg.TLS.CertFile),
					zap.String("key_file", cfg.TLS.KeyFile),
					zap.Error(err),
				)
				panic(fmt.Errorf("failed to load client certificate: %w", err))
			}
			tlsCfg.Certificates = []tls.Certificate{cert}

			// 检查证书的公钥算法
			if len(cert.Certificate) > 0 {
				parsedCert, err := x509.ParseCertificate(cert.Certificate[0])
				if err == nil {
					logger.Info("[kuryr] client certificate loaded successfully",
						zap.String("public_key_algorithm", parsedCert.PublicKeyAlgorithm.String()),
						zap.String("signature_algorithm", parsedCert.SignatureAlgorithm.String()))
				}
			}
		}

		// 加载CA证书
		if cfg.TLS.CAFile != "" {
			caCert, err := os.ReadFile(cfg.TLS.CAFile)
			if err != nil {
				logger.Error("[kuryr] failed to load CA certificate",
					zap.String("ca_file", cfg.TLS.CAFile),
					zap.Error(err),
				)
				panic(fmt.Errorf("failed to load CA certificate: %w", err))
			}

			caCertPool := x509.NewCertPool()
			if !caCertPool.AppendCertsFromPEM(caCert) {
				logger.Error("[kuryr] failed to parse CA certificate")
				panic(fmt.Errorf("failed to parse CA certificate"))
			}
			tlsCfg.RootCAs = caCertPool
			logger.Info("[kuryr] CA certificate loaded successfully")
		}

		clientCfg.TLS = tlsCfg
		logger.Info("[kuryr] TLS configuration enabled with Ed25519 support")
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

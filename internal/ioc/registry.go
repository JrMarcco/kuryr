package ioc

import (
	"github.com/JrMarcco/easy-grpc/registry"
	"github.com/JrMarcco/easy-grpc/registry/etcd"
	"github.com/spf13/viper"
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.uber.org/fx"
)

var RegistryFxOpt = fx.Provide(
	fx.Annotate(
		InitRegistry,
		fx.As(new(registry.Registry)),
	),
)

func InitRegistry(etcdClient *clientv3.Client) *etcd.Registry {
	type config struct {
		KeyPrefix string `mapstructure:"key_prefix"`
		LeaseTTL  int    `mapstructure:"lease_ttl"`
	}

	cfg := config{}
	if err := viper.UnmarshalKey("registry", &cfg); err != nil {
		panic(err)
	}

	r, err := etcd.NewBuilder(etcdClient).
		LeaseTTL(cfg.LeaseTTL).
		Build()
	if err != nil {
		panic(err)
	}
	return r
}

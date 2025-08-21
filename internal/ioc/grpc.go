package ioc

import (
	"time"

	"github.com/JrMarcco/easy-grpc/client"
	"github.com/JrMarcco/easy-grpc/client/br"
	"github.com/JrMarcco/easy-grpc/client/rr"
	"github.com/JrMarcco/easy-grpc/registry"
	clientv1 "github.com/JrMarcco/kuryr-api/api/go/client/v1"
	configv1 "github.com/JrMarcco/kuryr-api/api/go/config/v1"
	notificationv1 "github.com/JrMarcco/kuryr-api/api/go/notification/v1"
	providerv1 "github.com/JrMarcco/kuryr-api/api/go/provider/v1"
	"github.com/JrMarcco/kuryr/internal/api"
	"github.com/spf13/viper"
	"go.uber.org/fx"
	"google.golang.org/grpc"
	"google.golang.org/grpc/balancer"
	"google.golang.org/grpc/balancer/base"
	"google.golang.org/grpc/keepalive"
)

var GrpcFxOpt = fx.Module(
	"grpc",
	fx.Provide(
		InitGrpc,
		InitCallbackGrpcClients,
		api.NewBizInfoServer,
		api.NewBizConfigServer,
		api.NewProviderServer,
		api.NewNotificationServer,
	),
)

func InitGrpc(
	bizConfigServer *api.BizConfigServer,
	providerServer *api.ProviderServer,
	notificationServer *api.NotificationServer,
) *grpc.Server {
	grpcServer := grpc.NewServer()

	configv1.RegisterBizConfigServiceServer(grpcServer, bizConfigServer)
	providerv1.RegisterProviderServiceServer(grpcServer, providerServer)
	notificationv1.RegisterNotificationServiceServer(grpcServer, notificationServer)
	return grpcServer
}

// InitCallbackGrpcClients 初始化回调通知的 grpc 客户端。
func InitCallbackGrpcClients(r registry.Registry) *client.Manager[clientv1.CallbackServiceClient] {
	type keepaliveConfig struct {
		Time                int  `mapstructure:"time"`    // keep alive 请求间隔时间，单位：毫秒
		Timeout             int  `mapstructure:"timeout"` // keep alive 请求超时时间，单位：毫秒
		PermitWithoutStream bool `mapstructure:"permit_without_stream"`
	}

	type config struct {
		Name      string          `mapstructure:"name"`    // 负载均衡 resolver 名称
		Timeout   int             `mapstructure:"timeout"` // grpc 请求超时时间，单位：毫秒
		Keepalive keepaliveConfig `mapstructure:"keepalive"`
	}

	cfg := config{}
	if err := viper.UnmarshalKey("grpc.client", &cfg); err != nil {
		panic(err)
	}

	bb := base.NewBalancerBuilder(
		cfg.Name,
		br.NewRwWeightBalancerBuilder(),
		base.Config{
			HealthCheck: true,
		},
	)

	// 注册负载均衡
	balancer.Register(bb)

	return client.NewManagerBuilder[clientv1.CallbackServiceClient](
		rr.NewResolverBuilder(r, time.Duration(cfg.Timeout)*time.Millisecond),
		bb,
		func(conn *grpc.ClientConn) clientv1.CallbackServiceClient {
			return clientv1.NewCallbackServiceClient(conn)
		},
	).KeepAlive(keepalive.ClientParameters{
		Time:                time.Duration(cfg.Keepalive.Time) * time.Millisecond,
		Timeout:             time.Duration(cfg.Keepalive.Timeout) * time.Millisecond,
		PermitWithoutStream: cfg.Keepalive.PermitWithoutStream,
	}).Insecure().Build()
}

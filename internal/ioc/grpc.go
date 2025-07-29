package ioc

import (
	configv1 "github.com/JrMarcco/kuryr-api/api/config/v1"
	notificationv1 "github.com/JrMarcco/kuryr-api/api/notification/v1"
	providerv1 "github.com/JrMarcco/kuryr-api/api/provider/v1"
	"github.com/JrMarcco/kuryr/internal/api"
	"go.uber.org/fx"
	"google.golang.org/grpc"
)

var GrpcFxOpt = fx.Provide(
	InitGrpc,
	api.NewBizConfigServer,
	api.NewProviderServer,
	api.NewNotificationServer,
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

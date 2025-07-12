package ioc

import (
	notificationv1 "github.com/JrMarcco/kuryr-api/api/notification/v1"
	"github.com/JrMarcco/kuryr/internal/api"
	"go.uber.org/fx"
	"google.golang.org/grpc"
)

var GrpcFxOpt = fx.Provide(
	InitGrpc,
	api.NewNotificationServer,
)

func InitGrpc(notificationServer *api.NotificationServer) *grpc.Server {
	grpcServer := grpc.NewServer()

	notificationv1.RegisterNotificationServiceServer(grpcServer, notificationServer)
	return grpcServer
}

package api

import (
	"context"

	notificationv1 "github.com/JrMarcco/kuryr-api/api/notification/v1"
)

var _ notificationv1.NotificationServiceServer = (*NotificationServer)(nil)

type NotificationServer struct {
}

func (s *NotificationServer) Send(ctx context.Context, request *notificationv1.SendRequest) (*notificationv1.SendResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (s *NotificationServer) AsyncSend(ctx context.Context, request *notificationv1.AsyncSendRequest) (*notificationv1.AsyncSendResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (s *NotificationServer) BatchSend(ctx context.Context, request *notificationv1.BatchSendRequest) (*notificationv1.BatchSendResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (s *NotificationServer) AsyncBatchSend(ctx context.Context, request *notificationv1.AsyncBatchSendRequest) (*notificationv1.AsyncBatchSendResponse, error) {
	//TODO implement me
	panic("implement me")
}

func NewNotificationServer() *NotificationServer {
	return &NotificationServer{}
}

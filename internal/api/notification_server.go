package api

import notificationv1 "github.com/JrMarcco/kuryr-api/api/notification/v1"

type NotificationServer struct {
	notificationv1.UnimplementedNotificationServiceServer
}

func NewNotificationServer() *NotificationServer {
	return &NotificationServer{}
}

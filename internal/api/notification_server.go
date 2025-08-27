package api

import (
	"context"

	notificationv1 "github.com/JrMarcco/kuryr-api/api/go/notification/v1"
)

var _ notificationv1.NotificationServiceServer = (*NotificationServer)(nil)

// NotificationServer 消息发送服务。
//
// 消息发送调用链：
//
//	├── grpc client -> NotificationServer.Send / NotificationServer.AsyncSend / NotificationServer.BatchSend / NotificationServer.AsyncBatchSend
//	└── NotificationServer 根据 Notification 的 SendStrategy 选择不同的策略发送消息 ( ImmediateSendStrategy 或 DefaultSendStrategy )
//	    ├── DefaultSendStrategy 默认策略 ( 延迟发送策略 )
//	   	│   └── 创建记录并入库，等待异步发送。
//	    └── ImmediateSendStrategy 立即发送策略
//	    	├── 创建记录并入库。
//	    	├── ImmediateSendStrategy -> NotificationSender -> ChannelSender。
//	    	├── ChannelSender 选择供应商，此时是真正的消息下发。
//	    	└── 变更状态，返回结果。
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

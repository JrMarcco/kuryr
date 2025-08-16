package ports

import (
	"context"

	"github.com/JrMarcco/kuryr/internal/domain"
)

//go:generate mockgen -source=notification_sender.go -destination=./mock/notification_sender.mock.go -package=portsmock -typed NotificationSender

// Sender 消息发送接口，负责消息的实际发送逻辑。
// Sender 依照根据渠道调用不同的 channel.ChannelSender 来发送消息。
//
// notification.SendService 负责处理消息发送前的准备工作 ( Notification 记录入库，配额管理等 )，并调用 Sender 发送消息。
//
// 注意：
//
//	这里使用了 pool.TaskPool 来管理发送任务，避免阻塞主线程。
type NotificationSender interface {
	Send(ctx context.Context, n domain.Notification) (domain.SendResp, error)
	BatchSend(ctx context.Context, ns []domain.Notification) (domain.BatchSendResp, error)
}

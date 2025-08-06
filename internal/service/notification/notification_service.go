package notification

import (
	"context"

	"github.com/JrMarcco/kuryr/internal/domain"
)

// SendService 消息发送服务接口，负责处理消息发送前的准备工作。
// 例如：
//
//	Notification 入库
//	配额管理
//	...
type SendService interface {
	Send(ctx context.Context, n domain.Notification) (domain.SendResp, error)
	AsyncSend(ctx context.Context, n domain.Notification) (domain.Notification, error)

	BatchSend(ctx context.Context, ns []domain.Notification) (domain.BatchSendResp, error)
	BatchAsyncSend(ctx context.Context, ns []domain.Notification) (domain.BatchSendResp, error)
}

package sender

import (
	"context"

	"github.com/JrMarcco/easy-kit/pool"
	"github.com/JrMarcco/kuryr/internal/domain"
	"github.com/JrMarcco/kuryr/internal/repository"
	"github.com/JrMarcco/kuryr/internal/service/ports"
	"go.uber.org/zap"
)

//go:generate mockgen -source=sender.go -destination=./mock/sender.mock.go -package=sendermock -typed Sender

// Sender 消息发送接口，负责消息的实际发送逻辑。
// Sender 依照根据渠道调用不同的 channel.ChannelSender 来发送消息。
//
// notification.SendService 负责处理消息发送前的准备工作 ( Notification 记录入库，配额管理等 )，并调用 Sender 发送消息。
//
// 注意：
//
//	这里使用了 pool.TaskPool 来管理发送任务，避免阻塞主线程。
type Sender interface {
	Send(ctx context.Context, n domain.Notification) (domain.SendResp, error)
	BatchSend(ctx context.Context, ns []domain.Notification) (domain.BatchSendResp, error)
}

var _ Sender = (*DefaultSender)(nil)

type DefaultSender struct {
	bizConfigRepo    repository.BizConfigRepo
	callbackLogRepo  repository.CallbackLogRepo
	notificationRepo repository.NotificationRepo

	channelSender ports.ChannelSender

	taskPool *pool.TaskPool

	logger *zap.Logger
}

func (s *DefaultSender) Send(ctx context.Context, n domain.Notification) (domain.SendResp, error) {
	res := domain.SendResult{
		NotificationId: n.Id,
	}

	_, err := s.channelSender.Send(ctx, n)
	if err != nil {
		s.logger.Error("failed to send notification", zap.Error(err))

		res.SendStatus = domain.SendStatusFailure

		// TODO: 回退 quota
		// 标记发送失败
		n.SendStatus = domain.SendStatusFailure
		err = s.notificationRepo.MarkFailure(ctx, n)
	} else {
		res.SendStatus = domain.SendStatusSuccess
		n.SendStatus = domain.SendStatusSuccess
		err = s.notificationRepo.MarkSuccess(ctx, n)
	}

	if err != nil {
		// 更新发送状态失败
		return domain.SendResp{}, err
	}

	// TODO: 处理回调，写入一条 callback_log 等待异步任务执行

	return domain.SendResp{
		Result: res,
	}, nil
}

func (s *DefaultSender) BatchSend(ctx context.Context, ns []domain.Notification) (domain.BatchSendResp, error) {
	return domain.BatchSendResp{}, nil
}

func NewDefaultSender(
	bizConfigRepo repository.BizConfigRepo,
	callbackLogRepo repository.CallbackLogRepo,
	notificationRepo repository.NotificationRepo,
	channelSender ports.ChannelSender,
	taskPool *pool.TaskPool,
	logger *zap.Logger,
) *DefaultSender {
	return &DefaultSender{
		bizConfigRepo:    bizConfigRepo,
		callbackLogRepo:  callbackLogRepo,
		notificationRepo: notificationRepo,
		channelSender:    channelSender,
		taskPool:         taskPool,
		logger:           logger,
	}
}

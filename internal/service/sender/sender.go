package sender

import (
	"context"

	"github.com/JrMarcco/easy-kit/pool"
	"github.com/JrMarcco/kuryr/internal/domain"
	"github.com/JrMarcco/kuryr/internal/repository"
	"github.com/JrMarcco/kuryr/internal/service/ports"
	"go.uber.org/zap"
)

var _ ports.NotificationSender = (*DefaultSender)(nil)

type DefaultSender struct {
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

	callbackLog := domain.CallbackLog{
		Notification: domain.Notification{
			Id:         n.Id,
			SendStatus: n.SendStatus,
		},
		BizId:        n.BizId,
		BizKey:       n.BizKey,
		RetriedTimes: 0,
		NextRetryAt:  0,
		Status:       domain.CallbackLogStatusPrepare,
	}
	err = s.callbackLogRepo.Save(ctx, callbackLog)
	if err != nil {
		// 回调保存失败记录日志，不影响发送结果。
		s.logger.Error("failed to save callback log for notification [ %d ]", zap.String("notification_id", n.Id), zap.Error(err))
	}

	return domain.SendResp{
		Result: res,
	}, nil
}

func (s *DefaultSender) BatchSend(ctx context.Context, ns []domain.Notification) (domain.BatchSendResp, error) {
	return domain.BatchSendResp{}, nil
}

func NewDefaultSender(
	callbackLogRepo repository.CallbackLogRepo,
	notificationRepo repository.NotificationRepo,
	channelSender ports.ChannelSender,
	taskPool *pool.TaskPool,
	logger *zap.Logger,
) *DefaultSender {
	return &DefaultSender{
		callbackLogRepo:  callbackLogRepo,
		notificationRepo: notificationRepo,
		channelSender:    channelSender,
		taskPool:         taskPool,
		logger:           logger,
	}
}

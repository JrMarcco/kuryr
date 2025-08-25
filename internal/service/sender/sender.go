package sender

import (
	"context"
	"fmt"
	"sync"

	"github.com/JrMarcco/easy-kit/pool"
	"github.com/JrMarcco/kuryr/internal/domain"
	"github.com/JrMarcco/kuryr/internal/repository"
	"github.com/JrMarcco/kuryr/internal/service/ports"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

var _ ports.NotificationSender = (*DefaultSender)(nil)

type DefaultSender struct {
	callbackLogRepo  repository.CallbackLogRepo
	notificationRepo repository.NotificationRepo

	channelSender ports.ChannelSender

	taskPool pool.TaskPool

	logger *zap.Logger
}

func (s *DefaultSender) Send(ctx context.Context, n domain.Notification) (domain.SendResp, error) {
	res := domain.SendResult{
		NotificationId: n.Id,
	}

	_, err := s.channelSender.Send(ctx, n)
	if err != nil {
		s.logger.Error("[kuryr] failed to send notification", zap.Error(err))

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

	// 写入 callback log 记录。
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
		s.logger.Error("[kuryr] failed to save callback log for notification [ %d ]", zap.String("notification_id", n.Id), zap.Error(err))
	}

	return domain.SendResp{
		Result: res,
	}, nil
}

func (s *DefaultSender) BatchSend(ctx context.Context, ns []domain.Notification) (domain.BatchSendResp, error) {
	if len(ns) == 0 {
		return domain.BatchSendResp{}, nil
	}

	var successMu, failureMu sync.Mutex
	var successResults, failureResults []domain.SendResult

	var eg errgroup.Group
	for i := range ns {
		n := ns[i]

		eg.Go(func() error {
			err := s.taskPool.Submit(ctx, pool.TaskFunc(func(ctx context.Context) error {
				_, err := s.channelSender.Send(ctx, n)
				if err != nil {
					res := domain.SendResult{
						NotificationId: n.Id,
						SendStatus:     domain.SendStatusFailure,
					}

					failureMu.Lock()
					failureResults = append(failureResults, res)
					failureMu.Unlock()

					return nil
				}

				res := domain.SendResult{
					NotificationId: n.Id,
					SendStatus:     domain.SendStatusSuccess,
				}

				successMu.Lock()
				successResults = append(successResults, res)
				successMu.Unlock()

				return nil
			}))

			if err != nil {
				s.logger.Warn("[kuryr] failed to submit task to task pool", zap.Error(err), zap.String("notification_id", n.Id))
			}
			return err
		})
	}

	if err := eg.Wait(); err != nil {
		s.logger.Warn("[kuryr] failed to send notifications", zap.Error(err))
		return domain.BatchSendResp{}, fmt.Errorf("[kuryr] failed to send notifications: %w", err)
	}

	allId := make([]string, 0, len(successResults)+len(failureResults))
	for _, res := range successResults {
		allId = append(allId, res.NotificationId)
	}

	for _, res := range failureResults {
		allId = append(allId, res.NotificationId)
	}

	// TODO: 写入 callback log 记录

	// 合并结果并返回。
	return domain.BatchSendResp{
		Results: append(successResults, failureResults...),
	}, nil
}

func NewDefaultSender(
	callbackLogRepo repository.CallbackLogRepo,
	notificationRepo repository.NotificationRepo,
	channelSender ports.ChannelSender,
	taskPool pool.TaskPool,
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

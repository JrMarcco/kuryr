package sendstrategy

import (
	"context"
	"fmt"

	"github.com/JrMarcco/kuryr/internal/domain"
	"github.com/JrMarcco/kuryr/internal/errs"
	"github.com/JrMarcco/kuryr/internal/repository"
	"go.uber.org/zap"
)

var _ SendStrategy = (*DefaultSendStrategy)(nil)

type DefaultSendStrategy struct {
	bizConfigRepo    repository.BizConfigRepo
	notificationRepo repository.NotificationRepo

	logger *zap.Logger
}

func (s *DefaultSendStrategy) Send(ctx context.Context, n domain.Notification) (domain.SendResp, error) {
	n.SetSendTime()

	saved, err := s.save(ctx, n)
	if err != nil {
		return domain.SendResp{}, fmt.Errorf("[kuryr] failed to save notification: %w", err)
	}

	return domain.SendResp{
		Result: domain.SendResult{
			NotificationId: saved.Id,
			SendStatus:     saved.SendStatus,
		},
	}, nil
}

func (s *DefaultSendStrategy) save(ctx context.Context, n domain.Notification) (domain.Notification, error) {
	if s.needCallback(ctx, n) {
		return s.notificationRepo.SaveWithCallback(ctx, n)
	}
	return s.notificationRepo.Save(ctx, n)
}

func (s *DefaultSendStrategy) BatchSend(ctx context.Context, ns []domain.Notification) (domain.BatchSendResp, error) {
	if len(ns) == 0 {
		return domain.BatchSendResp{}, fmt.Errorf("%w: notifications cannot be empty", errs.ErrInvalidParam)
	}

	savedNs, err := s.batchSave(ctx, ns)
	if err != nil {
		return domain.BatchSendResp{}, fmt.Errorf("[kuryr] failed to save notifications: %w", err)
	}

	res := make([]domain.SendResult, 0, len(savedNs))
	for _, saved := range savedNs {
		res = append(res, domain.SendResult{
			NotificationId: saved.Id,
			SendStatus:     saved.SendStatus,
		})
	}
	return domain.BatchSendResp{Results: res}, nil
}

func (s *DefaultSendStrategy) batchSave(ctx context.Context, ns []domain.Notification) ([]domain.Notification, error) {
	const first = 0
	if s.needCallback(ctx, ns[first]) {
		return s.notificationRepo.BatchSaveWithCallback(ctx, ns)
	}
	return s.notificationRepo.BatchSave(ctx, ns)
}

func (s *DefaultSendStrategy) needCallback(ctx context.Context, n domain.Notification) bool {
	bizConfig, err := s.bizConfigRepo.FindByBizId(ctx, n.BizId)
	if err != nil {
		s.logger.Error("[kuryr] failed to find biz config", zap.Uint64("biz_id", n.BizId), zap.Error(err))
		return false
	}

	return bizConfig.CallbackConfig != nil
}

func NewDefaultSendStrategy(
	bizConfigRepo repository.BizConfigRepo,
	notificationRepo repository.NotificationRepo,
	logger *zap.Logger,
) *DefaultSendStrategy {
	return &DefaultSendStrategy{
		bizConfigRepo:    bizConfigRepo,
		notificationRepo: notificationRepo,
		logger:           logger,
	}
}

package sendstrategy

import (
	"context"
	"fmt"

	"github.com/JrMarcco/kuryr/internal/domain"
	"github.com/JrMarcco/kuryr/internal/errs"
	"github.com/JrMarcco/kuryr/internal/repository"
	"github.com/JrMarcco/kuryr/internal/service/ports"
)

var _ SendStrategy = (*ImmediateSendStrategy)(nil)

type ImmediateSendStrategy struct {
	sender           ports.NotificationSender
	notificationRepo repository.NotificationRepo
}

func (s *ImmediateSendStrategy) Send(ctx context.Context, n domain.Notification) (domain.SendResp, error) {
	n.SetSendTime()

	saved, err := s.notificationRepo.Save(ctx, n)
	if err != nil {
		return domain.SendResp{}, err
	}

	return s.sender.Send(ctx, saved)
}

// BatchSend 批量发送消息。
// 注意：
//
//	消息的发送策略必须相同。
func (s *ImmediateSendStrategy) BatchSend(ctx context.Context, ns []domain.Notification) (domain.BatchSendResp, error) {
	if len(ns) == 0 {
		return domain.BatchSendResp{}, fmt.Errorf("%w: notifications cannot be empty", errs.ErrInvalidParam)
	}

	for i := range ns {
		ns[i].SetSendTime()
	}

	saved, err := s.notificationRepo.BatchSave(ctx, ns)
	if err != nil {
		return domain.BatchSendResp{}, fmt.Errorf("[kuryr] failed to save notifications: %w", err)
	}

	return s.sender.BatchSend(ctx, saved)
}

func NewImmediateSendStrategy(
	sender ports.NotificationSender,
	notificationRepo repository.NotificationRepo,
) *ImmediateSendStrategy {
	return &ImmediateSendStrategy{
		sender:           sender,
		notificationRepo: notificationRepo,
	}
}

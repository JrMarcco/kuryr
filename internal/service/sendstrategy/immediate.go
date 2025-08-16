package sendstrategy

import (
	"context"

	"github.com/JrMarcco/kuryr/internal/domain"
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

	// 立即发送，设置状态为 sending
	n.SendStatus = domain.SendStatusSending
	created, err := s.notificationRepo.Create(ctx, n)
	if err != nil {
		return domain.SendResp{}, err
	}

	return s.sender.Send(ctx, created)
}

func (s *ImmediateSendStrategy) BatchSend(ctx context.Context, ns []domain.Notification) (domain.BatchSendResp, error) {
	// TODO: implement me
	panic("implement me")
}

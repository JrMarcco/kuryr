package sender

import (
	"context"

	"github.com/JrMarcco/kuryr/internal/domain"
)

//go:generate mockgen -source=sender.go -destination=./mock/sender.mock.go -package=sendermock -typed Sender

// Sender 消息发送接口，负责消息的实机发送逻辑。
type Sender interface {
	Send(ctx context.Context, n domain.Notification) (domain.SendResp, error)
	BatchSend(ctx context.Context, ns []domain.Notification) (domain.BatchSendResp, error)
}

var _ Sender = (*DefaultSender)(nil)

type DefaultSender struct {
}

func (s *DefaultSender) Send(ctx context.Context, n domain.Notification) (domain.SendResp, error) {
	return domain.SendResp{}, nil
}

func (s *DefaultSender) BatchSend(ctx context.Context, ns []domain.Notification) (domain.BatchSendResp, error) {
	return domain.BatchSendResp{}, nil
}

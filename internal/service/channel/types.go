package channel

import (
	"context"

	"github.com/JrMarcco/kuryr/internal/domain"
	"github.com/JrMarcco/kuryr/internal/errs"
)

// ChannelSender 渠道发送器接口，负责将消息发送给渠道。
type ChannelSender interface {
	Send(ctx context.Context, n domain.Notification) (domain.SendResp, error)
}

var _ ChannelSender = (*Dispatcher)(nil)

// Dispatcher 渠道分发器，负责将消息分发给不同的渠道。
type Dispatcher struct {
	senders map[domain.Channel]ChannelSender
}

func (d *Dispatcher) Send(ctx context.Context, n domain.Notification) (domain.SendResp, error) {
	if sender, ok := d.senders[n.Channel]; ok {
		return sender.Send(ctx, n)
	}

	return domain.SendResp{}, errs.ErrInvalidChannel
}

func NewDispatcher(senders map[domain.Channel]ChannelSender) *Dispatcher {
	return &Dispatcher{
		senders: senders,
	}
}

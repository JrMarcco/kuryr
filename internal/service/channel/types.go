package channel

import (
	"context"
	"fmt"

	"github.com/JrMarcco/kuryr/internal/domain"
	"github.com/JrMarcco/kuryr/internal/errs"
	"github.com/JrMarcco/kuryr/internal/service/ports"
	"github.com/JrMarcco/kuryr/internal/service/provider"
)

var _ ports.ChannelSender = (*DefaultChannelSender)(nil)

// DefaultChannelSender 默认渠道发送器，负责将消息发送给渠道。
type DefaultChannelSender struct {
	SelectorBuilder provider.SelectorBuilder
}

func (cs *DefaultChannelSender) Send(ctx context.Context, n domain.Notification) (domain.SendResp, error) {
	selector, err := cs.SelectorBuilder.Build()
	if err != nil {
		return domain.SendResp{}, fmt.Errorf("%w: %w", errs.ErrFailedToSendNotification, err)
	}

	for {
		p, selectErr := selector.Next(ctx, n)
		if selectErr != nil {
			// 选择供应商异常直接退出发送
			return domain.SendResp{}, fmt.Errorf("%w: %w", errs.ErrFailedToSendNotification, selectErr)
		}

		resp, sendErr := p.Send(ctx, n)
		if sendErr != nil {
			// 发送异常继续循环，获取下一个供应商来执行发送请求。
			continue
		}
		return resp, nil
	}
}

var _ ports.ChannelSender = (*Dispatcher)(nil)

// Dispatcher 渠道分发器，负责将消息分发给不同的渠道。
type Dispatcher struct {
	senders map[domain.Channel]ports.ChannelSender
}

func (d *Dispatcher) Send(ctx context.Context, n domain.Notification) (domain.SendResp, error) {
	if sender, ok := d.senders[n.Channel]; ok {
		return sender.Send(ctx, n)
	}

	return domain.SendResp{}, errs.ErrInvalidChannel
}

func NewDispatcher(senders map[domain.Channel]ports.ChannelSender) *Dispatcher {
	return &Dispatcher{
		senders: senders,
	}
}

package sendstrategy

import (
	"context"

	"github.com/JrMarcco/kuryr/internal/domain"
)

//go:generate mockgen -source=./types.go -destination=./mock/send_strategy.mock.go -package=sendstrategymock -typed SendStrategy

// SendStrategy 发送策略接口，负责根据发送策略配置发送消息。
//
// 固定两种实现：
// ├── DefaultSendStrategy 默认策略，使用异步发送。
// └── ImmediateSendStrategy 立即发送策略，使用同步发送。
type SendStrategy interface {
	Send(ctx context.Context, n domain.Notification) (domain.SendResp, error)
	BatchSend(ctx context.Context, ns []domain.Notification) (domain.BatchSendResp, error)
}

var _ SendStrategy = (*Dispatcher)(nil)

// Dispatcher 发送策略分发器，根据配置选择不同的策略发送消息。
type Dispatcher struct {
	defaultStrategy   SendStrategy
	immediateStrategy SendStrategy
}

func (d *Dispatcher) Send(ctx context.Context, n domain.Notification) (domain.SendResp, error) {
	return d.chooseStrategy(n).Send(ctx, n)
}

func (d *Dispatcher) BatchSend(ctx context.Context, ns []domain.Notification) (domain.BatchSendResp, error) {
	return d.chooseStrategy(ns[0]).BatchSend(ctx, ns)
}

func (d *Dispatcher) chooseStrategy(n domain.Notification) SendStrategy {
	if n.StrategyConfig.StrategyType == domain.SendStrategyImmediate {
		return d.immediateStrategy
	}
	return d.defaultStrategy
}

func NewDispatcher(defaultStrategy SendStrategy, immediateStrategy SendStrategy) *Dispatcher {
	return &Dispatcher{
		defaultStrategy:   defaultStrategy,
		immediateStrategy: immediateStrategy,
	}
}

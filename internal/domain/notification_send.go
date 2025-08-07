package domain

import (
	"fmt"
	"time"

	"github.com/JrMarcco/kuryr/internal/errs"
)

// SendStrategyType 发送策略类型
type SendStrategyType string

const (
	SendStrategyImmediate SendStrategyType = "immediate" // 立即发送
	SendStrategyDelayed   SendStrategyType = "delayed"   // 延迟发送
	SendStrategyScheduled SendStrategyType = "scheduled" // 定时发送
	SendStrategyWindow    SendStrategyType = "window"    // 窗口发送
	SendStrategyDeadline  SendStrategyType = "deadline"  // 截止时间发送
)

// SendStrategyConfig 发送策略配置领域对象。
type SendStrategyConfig struct {
	StrategyType SendStrategyType `json:"strategy_type"` // 发送策略类型
	Delay        time.Duration    `json:"delay"`         // 延迟发送时间（延迟策略参数）
	ScheduledAt  time.Time        `json:"scheduled_at"`  // 计划发送时间（定时发送策略参数）
	StartAt      time.Time        `json:"start_at"`      // 实际发送开始时间（窗口发送策略参数）
	EndAt        time.Time        `json:"end_at"`        // 实际发送结束时间（窗口发送策略参数）
	Deadline     time.Time        `json:"deadline"`      // 截止时间（截止策略参数）
}

func (c SendStrategyConfig) Validate() error {
	switch c.StrategyType {
	case SendStrategyImmediate:
		return nil

	case SendStrategyDelayed:
		if c.Delay <= 0 {
			return fmt.Errorf("%w: delay must be greater than 0", errs.ErrInvalidParam)
		}

	case SendStrategyScheduled:
		if c.ScheduledAt.IsZero() || c.ScheduledAt.Before(time.Now()) {
			return fmt.Errorf("%w: scheduled_at must be in the future", errs.ErrInvalidParam)
		}

	case SendStrategyWindow:
		if c.StartAt.IsZero() || c.StartAt.After(c.EndAt) {
			return fmt.Errorf("%w: start_at must be before end_at", errs.ErrInvalidParam)
		}

	case SendStrategyDeadline:
		if c.Deadline.IsZero() || c.Deadline.Before(time.Now()) {
			return fmt.Errorf("%w: deadline must be in the future", errs.ErrInvalidParam)
		}
	}

	return nil
}

// SendTimeWindow 获取发送时间窗口（最早和最晚发送时间）。
func (c SendStrategyConfig) SendTimeWindow() (startAt, endAt time.Time) {
	switch c.StrategyType {
	case SendStrategyImmediate:
		now := time.Now()
		const defaultEndDuration = 30 * time.Minute
		return now, now.Add(defaultEndDuration)

	case SendStrategyDelayed:
		now := time.Now()
		return now, now.Add(c.Delay)

	case SendStrategyDeadline:
		now := time.Now()
		return now, c.Deadline

	case SendStrategyScheduled:
		// 允许 10s 误差
		const scheduledTimeTolerance = 10 * time.Second
		return c.ScheduledAt.Add(-scheduledTimeTolerance), c.ScheduledAt.Add(scheduledTimeTolerance)

	case SendStrategyWindow:
		return c.StartAt, c.EndAt

	default:
		now := time.Now()
		return now, now
	}
}

// SendResult 消息发送结果领域对象
type SendResult struct {
	NotificationId uint64
	SendStatus     SendStatus
}

// SendResp 消息请求响应领域对象
type SendResp struct {
	Result SendResult
}

// BatchSendResp 消息批量发送请求响应领域对象
type BatchSendResp struct {
	Results []SendResult
}

// BatchAsyncSendResp 批量异步发送请求响应领域对象
type BatchAsyncSendResp struct {
	NotificationIds []uint64
}

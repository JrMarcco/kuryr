package domain

import (
	"fmt"
	"time"

	"github.com/JrMarcco/kuryr/internal/errs"
)

type SendStatus string

const (
	SendStatusPrepare SendStatus = "prepare"
	SendStatusPending SendStatus = "pending"
	SendStatusSending SendStatus = "sending"
	SendStatusSuccess SendStatus = "success"
	SendStatusFailure SendStatus = "failure"
	SendStatusCancel  SendStatus = "cancel"
)

// Template 消息关联模板信息领域对象。
// 包含模板 id、版本、参数。
type Template struct {
	Id      uint64            `json:"id"`      // 模板 id
	Version uint64            `json:"version"` // 模板版本
	Params  map[string]string `json:"params"`  // 模板参数
}

// Notification 通知消息领域对象。
type Notification struct {
	Id             uint64             `json:"id"`              // 消息 id
	BizId          uint64             `json:"biz_id"`          // 业务 id
	BizKey         string             `json:"biz_key"`         // 业务 key
	Receivers      []string           `json:"receivers"`       // 接收者
	Channel        Channel            `json:"channel"`         // 渠道
	Template       Template           `json:"template"`        // 模板
	SendStatus     SendStatus         `json:"send_status"`     // 发送状态
	ScheduledStrat time.Time          `json:"scheduled_strat"` // 计划发送开始时间
	ScheduledEnd   time.Time          `json:"scheduled_end"`   // 计划发送结束时间
	Version        int32              `json:"version"`         // 版本号
	StrategyConfig SendStrategyConfig `json:"strategy_config"` // 发送策略配置
}

func (n *Notification) Validate() error {
	if n.BizId == 0 {
		return fmt.Errorf("%w: biz id cannot be zero", errs.ErrInvalidParam)
	}

	if n.BizKey == "" {
		return fmt.Errorf("%w: biz key cannot be empty", errs.ErrInvalidParam)
	}

	if len(n.Receivers) == 0 {
		return fmt.Errorf("%w: receivers cannot be empty", errs.ErrInvalidParam)
	}

	if !n.Channel.IsValid() {
		return fmt.Errorf("%w: invalid channel: %d", errs.ErrInvalidParam, n.Channel)
	}

	if n.Template.Id == 0 {
		return fmt.Errorf("%w: template id cannot be zero", errs.ErrInvalidParam)
	}
	if n.Template.Version == 0 {
		return fmt.Errorf("%w: template version cannot be zero", errs.ErrInvalidParam)
	}

	if len(n.Template.Params) == 0 {
		return fmt.Errorf("%w: template params cannot be empty", errs.ErrInvalidParam)
	}

	if err := n.StrategyConfig.Validate(); err != nil {
		return err
	}

	return nil
}

func (n *Notification) SetSendTime() {
	startAt, endAt := n.StrategyConfig.SendTimeWindow()
	n.ScheduledStrat = startAt
	n.ScheduledEnd = endAt
}

func (n *Notification) IsImmediate() bool {
	return n.StrategyConfig.StrategyType == SendStrategyImmediate
}

// ReplaceAsyncImmediate 将立即发送的通知替换为截止时间发送。
func (n *Notification) ReplaceAsyncImmediate() {
	if n.IsImmediate() {
		n.StrategyConfig.Deadline = time.Now().Add(time.Minute)
		n.StrategyConfig.StrategyType = SendStrategyDeadline
	}
}

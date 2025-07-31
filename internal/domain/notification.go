package domain

import "time"

type SendStatus string

const (
	SendStatusPrepare SendStatus = "prepare"
	SendStatusPending SendStatus = "pending"
	SendStatusSending SendStatus = "sending"
	SendStatusSuccess SendStatus = "success"
	SendStatusFailure SendStatus = "failure"
	SendStatusCancel  SendStatus = "cancel"
)

type Template struct {
	Id      uint64            `json:"id"`
	Version uint64            `json:"version"`
	Params  map[string]string `json:"params"`
}

// Notification 通知消息领域对象
// TODO
type Notification struct {
	Id             uint64     `json:"id"`
	BizId          uint64     `json:"biz_id"`
	BizKey         string     `json:"biz_key"`
	Receivers      []string   `json:"receivers"`
	Channel        Channel    `json:"channel"`
	Template       Template   `json:"template"`
	SendStatus     SendStatus `json:"send_status"`
	ScheduledStrat time.Time  `json:"scheduled_strat"`
	ScheduledEnd   time.Time  `json:"scheduled_end"`
}

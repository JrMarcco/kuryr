package domain

// CallbackLogStatus 回调日志状态。
type CallbackLogStatus string

const (
	CallbackLogStatusPrepare CallbackLogStatus = "prepare"
	CallbackLogStatusPending CallbackLogStatus = "pending"
	CallbackLogStatusSuccess CallbackLogStatus = "success"
	CallbackLogStatusFailure CallbackLogStatus = "failure"
)

// CallbackLog 回调日志领域对象。
//
// 当消息发送策略为立即发送时，不会有回调日志。
type CallbackLog struct {
	Notification Notification `json:"notification"`

	Id           uint64            `json:"id"`
	BizId        uint64            `json:"biz_id"`
	BizKey       string            `json:"biz_key"`
	RetriedTimes int32             `json:"retried_times"`
	NextRetryAt  int64             `json:"next_retry_at"`
	Status       CallbackLogStatus `json:"status"` // 回调状态
}

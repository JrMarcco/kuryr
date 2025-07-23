package domain

import "github.com/JrMarcco/kuryr/internal/pkg/retry"

// BizConfig 业务方配置领域对象。
type BizConfig struct {
	Id             uint64          `json:"id"`
	ChannelConfig  *ChannelConfig  `json:"channel_config"` // 渠道配置
	QuotaConfig    *QuotaConfig    `json:"quota_config"`   // 配额配置
	CallbackConfig *CallbackConfig `json:"callback_config"`
	RateLimit      int             `json:"rate_limit"`
	CreatedAt      int64           `json:"created_at"`
	UpdatedAt      int64           `json:"updated_at"`
}

type ChannelItem struct {
	Channel  string `json:"channel"`
	Priority int    `json:"priority"`
	Enabled  bool   `json:"enabled"`
}

type ChannelConfig struct {
	Channels          []ChannelItem `json:"channels"`
	RetryPolicyConfig *retry.Config `json:"retry_policy_config"`
}

type Quota struct {
	SMS   int32 `json:"sms"`
	Email int32 `json:"email"`
}

type QuotaConfig struct {
	Daily   *Quota `json:"daily"`
	Monthly *Quota `json:"monthly"`
}

type CallbackConfig struct {
	ServiceName       string        `json:"service_name"`
	RetryPolicyConfig *retry.Config `json:"retry_policy_config"`
}

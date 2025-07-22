package domain

import "github.com/JrMarcco/kuryr/internal/pkg/retry"

// BizConfig 业务方配置领域对象。
type BizConfig struct {
	Id             uint64
	OwnerId        uint64         // 所有者 id ( biz_info.id )
	ChannelConfig  *ChannelConfig // 渠道配置
	QuotaConfig    *QuotaConfig   // 配额配置
	CallbackConfig *CallbackConfig
	RateLimit      int
	CreatedAt      int64
	UpdatedAt      int64
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

type QuotaDetail struct {
	SMS   int32 `json:"sms"`
	Email int32 `json:"email"`
}

type QuotaConfig struct {
	DailyQuota   *QuotaDetail `json:"daily_quota"`
	MonthlyQuota *QuotaDetail `json:"monthly_quota"`
}

type CallbackConfig struct {
	ServiceName       string        `json:"service_name"`
	RetryPolicyConfig *retry.Config `json:"retry_policy_config"`
}

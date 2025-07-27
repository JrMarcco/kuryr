package domain

import (
	"fmt"

	"github.com/JrMarcco/kuryr/internal/errs"
)

// Channel 通知渠道
type Channel string

const (
	ChannelSms   Channel = "sms"
	ChannelEmail Channel = "email"
)

func (c Channel) String() string {
	return string(c)
}

func (c Channel) IsValid() bool {
	return c == ChannelSms || c == ChannelEmail
}

func (c Channel) IsSms() bool {
	return c == ChannelSms
}

func (c Channel) IsEmail() bool {
	return c == ChannelEmail
}

// Provider 供应商信息
type Provider struct {
	Id           uint64  `json:"id"`
	ProviderName string  `json:"provider_name"` // 供应商名称
	Channel      Channel `json:"channel"`       // 渠道

	Endpoint string `json:"endpoint"`  // 接口地址
	RegionId string `json:"region_id"` // 区域 ID

	AppId     string `json:"app_id"`     // 应用 ID
	ApiKey    string `json:"api_key"`    // 接口密钥
	ApiSecret string `json:"api_secret"` // 接口密钥

	Weight     int `json:"weight"`      // 权重
	QpsLimit   int `json:"qps_limit"`   // 每秒请求限制
	DailyLimit int `json:"daily_limit"` // 每日请求限制

	AuditCallbackUrl string `json:"audit_callback_url"` // 审核回调地址

	ActiveStatus ActiveStatus `json:"active_status"` // 状态
}

func (p *Provider) IsValid() error {
	if p.ProviderName == "" {
		return fmt.Errorf("%w: provider name can not be empty", errs.ErrInvalidParam)
	}

	if !p.Channel.IsValid() {
		return fmt.Errorf("%w: invalid channel: %s", errs.ErrInvalidParam, p.Channel)
	}

	if p.Endpoint == "" {
		return fmt.Errorf("%w: provider endpoint can not be empty", errs.ErrInvalidParam)
	}

	if p.AppId == "" {
		return fmt.Errorf("%w: provider app id can not be empty", errs.ErrInvalidParam)
	}
	if p.ApiKey == "" {
		return fmt.Errorf("%w: provider api key can not be empty", errs.ErrInvalidParam)
	}
	if p.ApiSecret == "" {
		return fmt.Errorf("%w: provider api secret can not be empty", errs.ErrInvalidParam)
	}

	if p.Weight < 0 {
		return fmt.Errorf("%w: provider weight can not be negative", errs.ErrInvalidParam)
	}
	if p.QpsLimit <= 0 {
		return fmt.Errorf("%w: provider qps limit can not be negative or zero", errs.ErrInvalidParam)
	}
	if p.DailyLimit <= 0 {
		return fmt.Errorf("%w: provider daily limit can not be negative or zero", errs.ErrInvalidParam)
	}

	return nil
}

package domain

import (
	"fmt"

	"github.com/JrMarcco/kuryr/internal/errs"
	"github.com/JrMarcco/kuryr/internal/pkg/retry"
)

// BizType 业务类型。
type BizType string

const (
	BizTypeIndividual   BizType = "individual"
	BizTypeOrganization BizType = "organization"
)

func (bt BizType) IsValid() bool {
	switch bt {
	case BizTypeIndividual, BizTypeOrganization:
		return true
	default:
		return false
	}
}

func (bt BizType) IsIndividual() bool {
	return bt == BizTypeIndividual
}

func (bt BizType) IsOrganization() bool {
	return bt == BizTypeOrganization
}

// BizInfo 业务信息领域对象。
type BizInfo struct {
	Id     uint64 `json:"id"`
	BizKey string `json:"biz_key"`

	BizName   string  `json:"biz_name"`
	BizType   BizType `json:"biz_type"`
	BizSecret string  `json:"biz_secret"`

	Contact      string `json:"contact"`
	ContactEmail string `json:"contact_email"`

	CreatorId uint64 `json:"creator_id"`

	CreatedAt int64 `json:"created_at"`
	UpdatedAt int64 `json:"updated_at"`
}

func (bi *BizInfo) Validate() error {
	if bi.BizKey == "" {
		return fmt.Errorf("%w: biz key cannot be empty", errs.ErrInvalidParam)
	}

	if bi.BizName == "" {
		return fmt.Errorf("%w: biz name cannot be empty", errs.ErrInvalidParam)
	}

	if !bi.BizType.IsValid() {
		return fmt.Errorf("%w: invalid biz type [ %s ]", errs.ErrInvalidParam, bi.BizType)
	}

	if bi.Contact == "" {
		return fmt.Errorf("%w: contact cannot be empty", errs.ErrInvalidParam)
	}

	if bi.ContactEmail == "" {
		return fmt.Errorf("%w: contact email cannot be empty", errs.ErrInvalidParam)
	}

	return nil
}

// OwnerType 拥有者类型
type OwnerType string

const (
	OwnerTypeIndividual   OwnerType = "individual"   // 个人
	OwnerTypeOrganization OwnerType = "organization" // 组织
)

func (s OwnerType) IsValid() bool {
	switch s {
	case OwnerTypeIndividual, OwnerTypeOrganization:
		return true
	}
	return false
}

// BizConfig 业务方配置领域对象。
type BizConfig struct {
	Id             uint64          `json:"id"`
	BizId          uint64          `json:"biz_id"`
	OwnerType      OwnerType       `json:"owner_type"`
	ChannelConfig  *ChannelConfig  `json:"channel_config"` // 渠道配置
	QuotaConfig    *QuotaConfig    `json:"quota_config"`   // 配额配置
	CallbackConfig *CallbackConfig `json:"callback_config"`
	RateLimit      int32           `json:"rate_limit"`
}

type ChannelItem struct {
	Channel  Channel `json:"channel"`
	Priority int32   `json:"priority"`
	Enabled  bool    `json:"enabled"`
}

type ChannelConfig struct {
	Channels          []ChannelItem `json:"channels"`
	RetryPolicyConfig *retry.Config `json:"retry_policy_config"`
}

type Quota struct {
	Sms   int32 `json:"sms"`
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

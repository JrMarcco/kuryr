package domain

import (
	"fmt"

	"github.com/JrMarcco/kuryr/internal/errs"
)

type NotificationType uint8

const (
	NotificationTypeVerifyCode = 0
)

func (n NotificationType) Uint8() uint8 {
	return uint8(n)
}

func (n NotificationType) IsValid() bool {
	switch n {
	case NotificationTypeVerifyCode:
		return true
	}
	return false
}

type ChannelTemplate struct {
	Id        uint64    `json:"id"`
	OwnerId   uint64    `json:"owner_id"`   // 拥有者 id，即 biz_id
	OwnerType OwnerType `json:"owner_type"` // 拥有者类型，即 biz_type

	TplName string `json:"tpl_name"` // 模板名
	TplDesc string `json:"tpl_desc"` // 模板描述

	Channel             Channel          `json:"channel"`               // 渠道类型
	NotificationType    NotificationType `json:"notification_type"`     // 通知类型
	ActivatedTplVersion uint64           `json:"activated_tpl_version"` // 激活版本 id

	CreatedAt int64 `json:"created_at"`
	UpdatedAt int64 `json:"updated_at"`

	Versions []ChannelTemplateVersion `json:"versions"`
}

func (t ChannelTemplate) Validate() error {
	if t.OwnerId == 0 {
		return fmt.Errorf("%w: owner id can not be zero", errs.ErrInvalidParam)
	}
	if !t.OwnerType.IsValid() {
		return fmt.Errorf("%w: invalid owner type: %s", errs.ErrInvalidParam, t.OwnerType)
	}

	if t.TplName == "" {
		return fmt.Errorf("%w: template name can not be empty", errs.ErrInvalidParam)
	}
	if t.TplDesc == "" {
		return fmt.Errorf("%w: template desc can not be empty", errs.ErrInvalidParam)
	}

	if !t.Channel.IsValid() {
		return fmt.Errorf("%w: invalid channel: %s", errs.ErrInvalidParam, t.Channel)
	}
	if !t.NotificationType.IsValid() {
		return fmt.Errorf("%w: invalid msg type: %s", errs.ErrInvalidParam, t.NotificationType)
	}
	return nil
}

func (t ChannelTemplate) Published() bool {
	return t.ActivatedTplVersion > 0
}

func (t ChannelTemplate) ActivatedVersion() *ChannelTemplateVersion {
	if t.ActivatedTplVersion == 0 {
		return nil
	}

	for i := range t.Versions {
		if t.Versions[i].Id == t.ActivatedTplVersion {
			return &t.Versions[i]
		}
	}
	return nil
}

// ChannelTemplateVersion 渠道模板版本信息
type ChannelTemplateVersion struct {
	Id              uint64      `json:"id"`
	TplId           uint64      `json:"tpl_id"`           // 模板 id
	VersionName     string      `json:"version_name"`     // 版本名
	Signature       string      `json:"signature"`        // 签名
	Content         string      `json:"content"`          // 模板内容
	ApplyRemark     string      `json:"apply_remark"`     // 申请说明
	AuditId         uint64      `json:"audit_id"`         // 审批记录 id
	AuditorId       uint64      `json:"auditor_id"`       // 审批人 id
	AuditTime       int64       `json:"audit_time"`       // 审批时间
	AuditStatus     AuditStatus `json:"audit_status"`     // 审批状态
	RejectionReason string      `json:"rejection_reason"` // 拒绝原因
	LastReviewAt    int64       `json:"last_review_at"`   // 上次提交审批时间
	CreatedAt       int64       `json:"created_at"`
	UpdatedAt       int64       `json:"updated_at"`

	Providers []ChannelTemplateProvider `json:"providers"` // 关联供应商
}

type ChannelTemplateProvider struct {
	Id              uint64      `json:"id"`
	TplId           uint64      `json:"tpl_id"`           // 模板 id
	TplVersionId    uint64      `json:"tpl_version_id"`   // 模板版本 id
	ProviderId      uint64      `json:"provider_id"`      // 供应商 id
	ProviderName    string      `json:"provider_name"`    //
	ProviderChannel Channel     `json:"provider_channel"` // 供应商渠道类型
	ProviderTplId   string      `json:"provider_tpl_id"`  // 供应商侧模板 id
	AuditRequestId  string      `json:"audit_request_id"` // 审批请求 id
	AuditStatus     AuditStatus `json:"audit_status"`     // 审批状态
	RejectionReason string      `json:"rejection_reason"` // 拒绝原因
	LastReviewAt    int64       `json:"last_review_at"`   // 上次提交审批时间
	CreatedAt       int64       `json:"created_at"`
	UpdatedAt       int64       `json:"updated_at"`
}

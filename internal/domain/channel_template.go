package domain

import (
	"fmt"

	"github.com/JrMarcco/kuryr/internal/errs"
)

type NotificationType int32

const (
	NotificationTypeVerifyCode = 1
	NotificationTypeNotice     = 2
)

func (n NotificationType) IsValid() bool {
	switch n {
	case NotificationTypeVerifyCode, NotificationTypeNotice:
		return true
	}
	return false
}

// ChannelTemplate 渠道模板领域对象
type ChannelTemplate struct {
	Id        uint64    `json:"id"`
	OwnerId   uint64    `json:"owner_id"`   // 拥有者 id，即 biz_id
	OwnerType OwnerType `json:"owner_type"` // 拥有者类型，即 biz_type

	TplName string `json:"tpl_name"` // 模板名
	TplDesc string `json:"tpl_desc"` // 模板描述

	Channel            Channel          `json:"channel"`              // 渠道类型
	NotificationType   NotificationType `json:"notification_type"`    // 通知类型
	ActivatedVersionId uint64           `json:"activated_version_id"` // 激活版本 id

	CreatedAt int64 `json:"created_at"`
	UpdatedAt int64 `json:"updated_at"`

	Versions []ChannelTemplateVersion `json:"versions"`
}

func (t ChannelTemplate) Validate() error {
	if t.OwnerId == 0 {
		return fmt.Errorf("%w: owner id cannot be zero", errs.ErrInvalidParam)
	}
	if !t.OwnerType.IsValid() {
		return fmt.Errorf("%w: invalid owner type: %s", errs.ErrInvalidParam, t.OwnerType)
	}

	if t.TplName == "" {
		return fmt.Errorf("%w: template name cannot be empty", errs.ErrInvalidParam)
	}
	if t.TplDesc == "" {
		return fmt.Errorf("%w: template desc cannot be empty", errs.ErrInvalidParam)
	}

	if !t.Channel.IsValid() {
		return fmt.Errorf("%w: invalid channel: %d", errs.ErrInvalidParam, t.Channel)
	}
	if !t.NotificationType.IsValid() {
		return fmt.Errorf("%w: invalid msg type: %d", errs.ErrInvalidParam, t.NotificationType)
	}
	return nil
}

// GetActivatedVersion 当前启用版本
func (t ChannelTemplate) GetActivatedVersion() (ChannelTemplateVersion, error) {
	if t.ActivatedVersionId == 0 {
		return ChannelTemplateVersion{}, fmt.Errorf("%w: channel template id = %d", errs.ErrNoActivatedTplVersion, t.Id)
	}

	for i := range t.Versions {
		if t.Versions[i].Id == t.ActivatedVersionId {
			if t.Versions[i].AuditStatus != AuditStatusApproved {
				return ChannelTemplateVersion{}, fmt.Errorf("%w: channel template id = %d, version id = %d", errs.ErrNotApprovedTplVersion, t.Id, t.ActivatedVersionId)
			}

			return t.Versions[i], nil
		}
	}
	return ChannelTemplateVersion{}, fmt.Errorf("%w: channel template id = %d", errs.ErrNoActivatedTplVersion, t.Id)
}

// GetVersion 根据 id 获取版本信息
func (t ChannelTemplate) GetVersion(versionId uint64) *ChannelTemplateVersion {
	for i := range t.Versions {
		if t.Versions[i].Id == versionId {
			return &t.Versions[i]
		}
	}
	return nil
}

// HasApprovedVersion 检查是否存在已审核通过的版本
func (t ChannelTemplate) HasApprovedVersion() bool {
	for i := range t.Versions {
		if t.Versions[i].AuditStatus == AuditStatusApproved {
			return true
		}
	}
	return false
}

// GetProvider 获取特定版本关联的供应商
func (t ChannelTemplate) GetProvider(versionId, providerId uint64) *ChannelTemplateProvider {
	version := t.GetVersion(versionId)
	if version == nil {
		return nil
	}

	for i := range version.Providers {
		if version.Providers[i].Id == providerId {
			return &version.Providers[i]
		}
	}
	return nil
}

// ChannelTemplateVersion 渠道模板版本信息领域对象
type ChannelTemplateVersion struct {
	Id    uint64 `json:"id"`
	TplId uint64 `json:"tpl_id"` // 模板 id

	VersionName string `json:"version_name"` // 版本名
	Signature   string `json:"signature"`    // 签名
	Content     string `json:"content"`      // 模板内容

	ApplyRemark string `json:"apply_remark"` // 申请说明

	AuditId         uint64      `json:"audit_id"`         // 审批记录 id
	AuditorId       uint64      `json:"auditor_id"`       // 审批人 id
	AuditTime       int64       `json:"audit_time"`       // 审批时间
	AuditStatus     AuditStatus `json:"audit_status"`     // 审批状态
	RejectionReason string      `json:"rejection_reason"` // 拒绝原因
	LastReviewAt    int64       `json:"last_review_at"`   // 上次提交审批时间

	CreatedAt int64 `json:"created_at"`
	UpdatedAt int64 `json:"updated_at"`

	Providers []ChannelTemplateProvider `json:"providers"` // 关联供应商
}

func (tv ChannelTemplateVersion) Validate() error {
	if tv.TplId == 0 {
		return fmt.Errorf("%w: template id cannot be zero", errs.ErrInvalidParam)
	}
	if tv.VersionName == "" {
		return fmt.Errorf("%w: version name cannot be empty", errs.ErrInvalidParam)
	}
	if tv.Signature == "" {
		return fmt.Errorf("%w: signature cannot be empty", errs.ErrInvalidParam)
	}
	if tv.Content == "" {
		return fmt.Errorf("%w: content cannot be empty", errs.ErrInvalidParam)
	}
	if tv.ApplyRemark == "" {
		return fmt.Errorf("%w: apply remark cannot be empty", errs.ErrInvalidParam)
	}
	return nil
}

// ChannelTemplateProvider 渠道模板供应商领域对象
type ChannelTemplateProvider struct {
	Id           uint64 `json:"id"`
	TplId        uint64 `json:"tpl_id"`         // 模板 id
	TplVersionId uint64 `json:"tpl_version_id"` // 模板版本 id

	ProviderId      uint64  `json:"provider_id"`      // 供应商 id
	ProviderName    string  `json:"provider_name"`    // 供应商名称
	ProviderTplId   string  `json:"provider_tpl_id"`  // 供应商侧模板 id
	ProviderChannel Channel `json:"provider_channel"` // 供应商渠道类型

	AuditRequestId  string      `json:"audit_request_id"` // 审批请求 id
	AuditStatus     AuditStatus `json:"audit_status"`     // 审批状态
	RejectionReason string      `json:"rejection_reason"` // 拒绝原因
	LastReviewAt    int64       `json:"last_review_at"`   // 上次提交审批时间

	CreatedAt int64 `json:"created_at"`
	UpdatedAt int64 `json:"updated_at"`
}

func (tp ChannelTemplateProvider) Validate() error {
	if tp.TplId == 0 {
		return fmt.Errorf("%w: template id cannot be zero", errs.ErrInvalidParam)
	}
	if tp.TplVersionId == 0 {
		return fmt.Errorf("%w: template version id cannot be zero", errs.ErrInvalidParam)
	}
	if tp.ProviderId == 0 {
		return fmt.Errorf("%w: provider id cannot be zero", errs.ErrInvalidParam)
	}
	if tp.ProviderName == "" {
		return fmt.Errorf("%w: provider name cannot be empty", errs.ErrInvalidParam)
	}
	if !tp.ProviderChannel.IsValid() {
		return fmt.Errorf("%w: invalid provider channel: %d", errs.ErrInvalidParam, tp.ProviderChannel)
	}
	return nil
}

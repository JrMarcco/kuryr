package dao

import (
	"context"

	"gorm.io/gorm"
)

// ChannelTemplate 渠道模板信息表
type ChannelTemplate struct {
	Id        uint64 `gorm:"column:id"`
	OwnerId   uint64 `gorm:"column:owner_id"`
	OwnerType string `gorm:"column:owner_type"`

	TplName string `gorm:"column:tpl_name"`
	TplDesc string `gorm:"column:tpl_desc"`

	Channel          int32  `gorm:"column:channel"`
	NotificationType int32  `gorm:"column:notification_type"`
	ActivatedVersion uint64 `gorm:"column:activated_version"`

	CreatedAt int64 `gorm:"column:created_at"`
	UpdatedAt int64 `gorm:"column:updated_at"`
}

func (ChannelTemplate) TableName() string {
	return "channel_template"
}

// ChannelTemplateVersion 渠道模板版本信息表
type ChannelTemplateVersion struct {
	Id    uint64 `gorm:"column:id"`
	TplId uint64 `gorm:"column:tpl_id"`

	VersionName string `gorm:"column:version_name"`
	Signature   string `gorm:"column:signature"`
	Content     string `gorm:"column:content"`

	ApplyRemark     string `gorm:"column:apply_remark"`
	AuditId         uint64 `gorm:"column:audit_id"`
	AuditorId       uint64 `gorm:"column:auditor_id"`
	AuditTime       int64  `gorm:"column:audit_time"`
	AuditStatus     string `gorm:"column:audit_status"`
	RejectionReason string `gorm:"column:rejection_reason"`
	LastReviewAt    int64  `gorm:"column:last_review_at"`

	CreatedAt int64 `gorm:"column:created_at"`
	UpdatedAt int64 `gorm:"column:updated_at"`
}

func (ChannelTemplateVersion) TableName() string {
	return "channel_template_version"
}

// ChannelTemplateProvider 渠道模板供应商信息表
type ChannelTemplateProvider struct {
	Id           uint64 `gorm:"column:id"`
	TplId        uint64 `gorm:"column:tpl_id"`
	TplVersionId uint64 `gorm:"column:tpl_version_id"`

	ProviderId      uint64 `gorm:"column:provider_id"`
	ProviderName    string `gorm:"column:provider_name"`
	ProviderChannel int32  `gorm:"column:provider_channel"`
	ProviderTplId   string `gorm:"column:provider_tpl_id"`

	AuditRequestId  string `gorm:"column:audit_request_id"`
	AuditStatus     string `gorm:"column:audit_status"`
	RejectionReason string `gorm:"column:rejection_reason"`
	LastReviewAt    int64  `gorm:"column:last_review_at"`

	CreatedAt int64 `gorm:"column:created_at"`
	UpdatedAt int64 `gorm:"column:updated_at"`
}

func (ChannelTemplateProvider) TableName() string {
	return "channel_template_provider"
}

type ChannelTplDao interface {
	FindById(ctx context.Context, id uint64) (ChannelTemplate, error)
	FindVersionsByIds(ctx context.Context, ids []uint64) ([]ChannelTemplateVersion, error)
	FindProviderByVersionIds(ctx context.Context, versionIds []uint64) ([]ChannelTemplateProvider, error)
}

var _ ChannelTplDao = (*DefaultChannelTplDao)(nil)

type DefaultChannelTplDao struct {
	db *gorm.DB
}

func (d *DefaultChannelTplDao) FindById(ctx context.Context, id uint64) (ChannelTemplate, error) {
	var tpl ChannelTemplate
	err := d.db.WithContext(ctx).Model(&ChannelTemplate{}).
		Where("id = ?", id).
		First(&tpl).Error
	return tpl, err
}

func (d *DefaultChannelTplDao) FindVersionsByIds(ctx context.Context, ids []uint64) ([]ChannelTemplateVersion, error) {
	if len(ids) == 0 {
		return []ChannelTemplateVersion{}, nil
	}

	var versions []ChannelTemplateVersion
	err := d.db.WithContext(ctx).Model(&ChannelTemplateVersion{}).
		Where("tpl_id in (?)", ids).
		Find(&versions).Error
	return versions, err
}

func (d *DefaultChannelTplDao) FindProviderByVersionIds(ctx context.Context, versionIds []uint64) ([]ChannelTemplateProvider, error) {
	if len(versionIds) == 0 {
		return []ChannelTemplateProvider{}, nil
	}

	var providers []ChannelTemplateProvider
	err := d.db.WithContext(ctx).Model(&ChannelTemplateProvider{}).
		Where("tpl_version_id in (?)", versionIds).
		Find(&providers).Error
	return providers, err
}

func NewDefaultChannelTplDao(db *gorm.DB) *DefaultChannelTplDao {
	return &DefaultChannelTplDao{
		db: db,
	}
}

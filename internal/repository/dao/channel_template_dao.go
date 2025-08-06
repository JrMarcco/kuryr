package dao

import (
	"context"
	"time"

	"gorm.io/gorm"
)

// ChannelTemplate 渠道模板信息表
type ChannelTemplate struct {
	Id        uint64 `gorm:"column:id"`
	OwnerId   uint64 `gorm:"column:owner_id"`
	OwnerType string `gorm:"column:owner_type"`

	TplName string `gorm:"column:tpl_name"`
	TplDesc string `gorm:"column:tpl_desc"`

	Channel            int32  `gorm:"column:channel"`
	NotificationType   int32  `gorm:"column:notification_type"`
	ActivatedVersionId uint64 `gorm:"column:activated_version_id"`

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

	ApplyRemark string `gorm:"column:apply_remark"`

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
	ProviderTplId   string `gorm:"column:provider_tpl_id"`
	ProviderChannel int32  `gorm:"column:provider_channel"`

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
	SaveTemplate(ctx context.Context, template ChannelTemplate) error
	SaveVersion(ctx context.Context, version ChannelTemplateVersion) error
	SaveProviders(ctx context.Context, providers []ChannelTemplateProvider) error

	// ActivateVersion 激活版本
	ActivateVersion(ctx context.Context, templateId uint64, versionId uint64) error

	FindById(ctx context.Context, id uint64) (ChannelTemplate, error)
	FindVersionById(ctx context.Context, id uint64) (ChannelTemplateVersion, error)
	FindVersionsByIds(ctx context.Context, ids []uint64) ([]ChannelTemplateVersion, error)
	FindProviderByTplId(ctx context.Context, tplId uint64) ([]ChannelTemplateProvider, error)
	FindProviderByVersionIds(ctx context.Context, versionIds []uint64) ([]ChannelTemplateProvider, error)

	ListByOwner(ctx context.Context, ownerId uint64) ([]ChannelTemplate, error)
}

var _ ChannelTplDao = (*DefaultChannelTplDao)(nil)

type DefaultChannelTplDao struct {
	db *gorm.DB
}

func (d *DefaultChannelTplDao) SaveTemplate(ctx context.Context, template ChannelTemplate) error {
	now := time.Now().UnixMilli()
	template.CreatedAt = now
	template.UpdatedAt = now

	return d.db.WithContext(ctx).Model(&ChannelTemplate{}).Create(&template).Error
}

func (d *DefaultChannelTplDao) SaveVersion(ctx context.Context, version ChannelTemplateVersion) error {
	now := time.Now().UnixMilli()
	version.CreatedAt = now
	version.UpdatedAt = now

	return d.db.WithContext(ctx).Model(&ChannelTemplateVersion{}).Create(&version).Error
}

func (d *DefaultChannelTplDao) SaveProviders(ctx context.Context, providers []ChannelTemplateProvider) error {
	now := time.Now().UnixMilli()
	for i := range providers {
		providers[i].CreatedAt = now
		providers[i].UpdatedAt = now
	}

	return d.db.WithContext(ctx).Model(&ChannelTemplateProvider{}).Create(&providers).Error
}

func (d *DefaultChannelTplDao) ActivateVersion(ctx context.Context, templateId uint64, versionId uint64) error {
	return d.db.WithContext(ctx).Model(&ChannelTemplate{}).
		Where("id = ?", templateId).
		Updates(map[string]any{
			"activated_version_id": versionId,
			"updated_at":           time.Now().UnixMilli(),
		}).Error
}

func (d *DefaultChannelTplDao) FindById(ctx context.Context, id uint64) (ChannelTemplate, error) {
	var tpl ChannelTemplate
	err := d.db.WithContext(ctx).Model(&ChannelTemplate{}).
		Where("id = ?", id).
		First(&tpl).Error
	return tpl, err
}

func (d *DefaultChannelTplDao) FindVersionById(ctx context.Context, id uint64) (ChannelTemplateVersion, error) {
	var version ChannelTemplateVersion
	err := d.db.WithContext(ctx).Model(&ChannelTemplateVersion{}).
		Where("id = ?", id).
		First(&version).Error
	return version, err
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

func (d *DefaultChannelTplDao) FindProviderByTplId(ctx context.Context, tplId uint64) ([]ChannelTemplateProvider, error) {
	if tplId == 0 {
		return []ChannelTemplateProvider{}, nil
	}

	var providers []ChannelTemplateProvider
	err := d.db.WithContext(ctx).Model(&ChannelTemplateProvider{}).
		Where("tpl_id = ?", tplId).
		Find(&providers).Error
	return providers, err
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

func (d *DefaultChannelTplDao) ListByOwner(ctx context.Context, ownerId uint64) ([]ChannelTemplate, error) {
	var templates []ChannelTemplate
	err := d.db.WithContext(ctx).Model(&ChannelTemplate{}).
		Where("owner_id = ?", ownerId).
		Find(&templates).Error
	return templates, err
}

func NewDefaultChannelTplDao(db *gorm.DB) *DefaultChannelTplDao {
	return &DefaultChannelTplDao{
		db: db,
	}
}

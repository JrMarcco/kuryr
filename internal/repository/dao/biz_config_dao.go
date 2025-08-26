package dao

import (
	"context"
	"time"

	"github.com/JrMarcco/kuryr/internal/domain"
	pkgsql "github.com/JrMarcco/kuryr/internal/pkg/sql"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type BizConfig struct {
	Id             uint64                                   `gorm:"column:id"`
	BizId          uint64                                   `gorm:"column:biz_id"`
	OwnerType      string                                   `gorm:"column:owner_type"`
	ChannelConfig  pkgsql.JsonColumn[domain.ChannelConfig]  `gorm:"column:channel_config;type:JSON"`
	QuotaConfig    pkgsql.JsonColumn[domain.QuotaConfig]    `gorm:"column:quota_config;type:JSON"`
	CallbackConfig pkgsql.JsonColumn[domain.CallbackConfig] `gorm:"column:callback_config;type:JSON"`
	RateLimit      int32                                    `gorm:"column:rate_limit"`
	CreatedAt      int64                                    `gorm:"column:created_at"`
	UpdatedAt      int64                                    `gorm:"column:updated_at"`
}

func (BizConfig) TableName() string {
	return "biz_config"
}

type BizConfigDao interface {
	Save(ctx context.Context, bizConfig BizConfig) (BizConfig, error)
	Update(ctx context.Context, bizConfig BizConfig) (BizConfig, error)

	Delete(ctx context.Context, id uint64) error
	DeleteInTx(ctx context.Context, tx *gorm.DB, id uint64) error

	FindById(ctx context.Context, id uint64) (BizConfig, error)
}

var _ BizConfigDao = (*DefaultBizConfigDao)(nil)

type DefaultBizConfigDao struct {
	db *gorm.DB
}

func (d *DefaultBizConfigDao) Save(ctx context.Context, bizConfig BizConfig) (BizConfig, error) {
	now := time.Now().UnixMilli()
	bizConfig.CreatedAt = now
	bizConfig.UpdatedAt = now

	err := d.db.WithContext(ctx).Model(&BizConfig{}).
		Clauses(clause.Returning{}).
		Create(&bizConfig).
		Scan(&bizConfig).Error
	return bizConfig, err
}

func (d *DefaultBizConfigDao) Update(ctx context.Context, bizConfig BizConfig) (BizConfig, error) {
	now := time.Now().UnixMilli()
	bizConfig.UpdatedAt = now

	values := map[string]any{
		"updated_at": now,
	}

	if bizConfig.ChannelConfig.Valid {
		values["channel_config"] = bizConfig.ChannelConfig
	}

	if bizConfig.QuotaConfig.Valid {
		values["quota_config"] = bizConfig.QuotaConfig
	}

	if bizConfig.CallbackConfig.Valid {
		values["callback_config"] = bizConfig.CallbackConfig
	}

	if bizConfig.RateLimit != 0 {
		values["rate_limit"] = bizConfig.RateLimit
	}

	err := d.db.WithContext(ctx).Model(&BizConfig{}).
		Clauses(clause.Returning{}). // 这里一定要返回更新后的全量字段，否则会导致缓存更新失败
		Where("id = ?", bizConfig.Id).
		Updates(values).
		Scan(&bizConfig).Error
	return bizConfig, err
}

func (d *DefaultBizConfigDao) Delete(ctx context.Context, id uint64) error {
	return d.db.WithContext(ctx).Model(&BizConfig{}).
		Where("id = ?", id).
		Delete(&BizConfig{}).Error
}

func (d *DefaultBizConfigDao) DeleteInTx(ctx context.Context, tx *gorm.DB, id uint64) error {
	return tx.WithContext(ctx).Model(&BizConfig{}).
		Where("id = ?", id).
		Delete(&BizConfig{}).Error
}

func (d *DefaultBizConfigDao) FindById(ctx context.Context, id uint64) (BizConfig, error) {
	var bizConfig BizConfig
	err := d.db.WithContext(ctx).Model(&BizConfig{}).
		Where("id = ?", id).
		First(&bizConfig).Error
	return bizConfig, err
}

func NewDefaultBizConfigDao(db *gorm.DB) *DefaultBizConfigDao {
	return &DefaultBizConfigDao{
		db: db,
	}
}

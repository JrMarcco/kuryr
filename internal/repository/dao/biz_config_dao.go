package dao

import (
	"context"
	"time"

	"github.com/JrMarcco/kuryr/internal/domain"
	"github.com/JrMarcco/kuryr/internal/pkg/xsql"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type BizConfig struct {
	Id             uint64                                 `gorm:"column:id"`
	OwnerType      string                                 `gorm:"column:owner_type"`
	ChannelConfig  xsql.JsonColumn[domain.ChannelConfig]  `gorm:"column:channel_config;type:JSON"`
	QuotaConfig    xsql.JsonColumn[domain.QuotaConfig]    `gorm:"column:quota_config;type:JSON"`
	CallbackConfig xsql.JsonColumn[domain.CallbackConfig] `gorm:"column:callback_config;type:JSON"`
	RateLimit      int32                                  `gorm:"column:rate_limit"`
	CreatedAt      int64                                  `gorm:"column:created_at"`
	UpdatedAt      int64                                  `gorm:"column:updated_at"`
}

func (BizConfig) TableName() string {
	return "biz_config"
}

type BizConfigDao interface {
	SaveOrUpdate(ctx context.Context, bizConfig BizConfig) (BizConfig, error)

	Delete(ctx context.Context, id uint64) error
	DeleteInTx(ctx context.Context, tx *gorm.DB, id uint64) error

	FindById(ctx context.Context, id uint64) (BizConfig, error)
}

var _ BizConfigDao = (*DefaultBizConfigDao)(nil)

type DefaultBizConfigDao struct {
	db *gorm.DB
}

func (d *DefaultBizConfigDao) SaveOrUpdate(ctx context.Context, bizConfig BizConfig) (BizConfig, error) {
	now := time.Now().UnixMilli()
	bizConfig.CreatedAt = now
	bizConfig.UpdatedAt = now

	// 使用 upsert 语句，根据 id 判断冲突
	res := d.db.WithContext(ctx).Model(&BizConfig{}).Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "id"}},
		DoUpdates: clause.Assignments(map[string]any{
			"channel_config":  bizConfig.ChannelConfig,
			"quota_config":    bizConfig.QuotaConfig,
			"callback_config": bizConfig.CallbackConfig,
			"rate_limit":      bizConfig.RateLimit,
			"updated_at":      now,
		}),
	}).Create(&bizConfig)
	if res.Error != nil {
		return BizConfig{}, res.Error
	}
	return bizConfig, nil
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

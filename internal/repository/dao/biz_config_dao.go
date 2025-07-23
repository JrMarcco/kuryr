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
	Id             uint64
	ChannelConfig  xsql.JsonColumn[domain.ChannelConfig]  `gorm:"type:json;serializer:json"`
	QuotaConfig    xsql.JsonColumn[domain.QuotaConfig]    `gorm:"type:json;serializer:json"`
	CallbackConfig xsql.JsonColumn[domain.CallbackConfig] `gorm:"type:json;serializer:json"`
	RateLimit      int
	CreatedAt      int64
	UpdatedAt      int64
}

func (bc BizConfig) TableName() string {
	return "biz_config"
}

type BizConfigDao interface {
	SaveOrUpdate(ctx context.Context, bizConfig BizConfig) (BizConfig, error)
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

func NewDefaultBizConfigDao(db *gorm.DB) *DefaultBizConfigDao {
	return &DefaultBizConfigDao{
		db: db,
	}
}

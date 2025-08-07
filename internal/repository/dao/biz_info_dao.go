package dao

import (
	"context"
	"time"

	pkggorm "github.com/JrMarcco/kuryr/internal/pkg/gorm"
	"github.com/JrMarcco/kuryr/internal/search"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type BizInfo struct {
	Id           uint64 `gorm:"column:id"`
	BizType      string `gorm:"column:biz_type"`
	BizKey       string `gorm:"column:biz_key"`
	BizSecret    string `gorm:"column:biz_secret"`
	BizName      string `gorm:"column:biz_name"`
	Contact      string `gorm:"column:contact"`
	ContactEmail string `gorm:"column:contact_email"`
	CreatorId    uint64 `gorm:"column:creator_id"`
	CreatedAt    int64  `gorm:"column:created_at"`
	UpdatedAt    int64  `gorm:"column:updated_at"`
}

func (BizInfo) TableName() string {
	return "biz_info"
}

type BizInfoDao interface {
	Save(ctx context.Context, bizInfo BizInfo) error
	Delete(ctx context.Context, id uint64) error
	DeleteInTx(ctx context.Context, tx *gorm.DB, id uint64) error

	Search(ctx context.Context, criteria search.BizSearchCriteria, param *pkggorm.PaginationParam) (*pkggorm.PaginationResult[BizInfo], error)
	FindById(ctx context.Context, id uint64) (BizInfo, error)
}

var _ BizInfoDao = (*DefaultBizInfoDao)(nil)

type DefaultBizInfoDao struct {
	db *gorm.DB
}

func (d *DefaultBizInfoDao) Save(ctx context.Context, bizInfo BizInfo) error {
	now := time.Now().UnixMilli()
	bizInfo.CreatedAt = now
	bizInfo.UpdatedAt = now

	// 这里使用 upsert
	err := d.db.WithContext(ctx).Model(&BizInfo{}).
		Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: "id"}},
			DoUpdates: clause.Assignments(map[string]any{
				"biz_type":   bizInfo.BizType,
				"biz_key":    bizInfo.BizKey,
				"biz_name":   bizInfo.BizName,
				"updated_at": now,
			}),
		}).Create(&bizInfo).Error
	return err
}

func (d *DefaultBizInfoDao) Delete(ctx context.Context, id uint64) error {
	return d.db.WithContext(ctx).Model(&BizInfo{}).
		Where("id = ?", id).
		Delete(&BizInfo{}).Error
}

func (d *DefaultBizInfoDao) DeleteInTx(ctx context.Context, tx *gorm.DB, id uint64) error {
	return tx.WithContext(ctx).Model(&BizInfo{}).
		Where("id = ?", id).
		Delete(&BizInfo{}).Error
}

func (d *DefaultBizInfoDao) Search(ctx context.Context, criteria search.BizSearchCriteria, param *pkggorm.PaginationParam) (*pkggorm.PaginationResult[BizInfo], error) {
	var records []BizInfo

	query := d.db.WithContext(ctx).Model(&BizInfo{})
	if criteria.BizName != "" {
		query = query.Where("biz_name LIKE ?", pkggorm.BuildLikePattern(criteria.BizName))
	}
	return pkggorm.Pagination(query, param, records)
}

func (d *DefaultBizInfoDao) FindById(ctx context.Context, id uint64) (BizInfo, error) {
	var bizInfo BizInfo
	err := d.db.WithContext(ctx).Model(&BizInfo{}).
		Where("id = ?", id).
		First(&bizInfo).Error
	return bizInfo, err
}

func NewDefaultBizInfoDao(db *gorm.DB) *DefaultBizInfoDao {
	return &DefaultBizInfoDao{
		db: db,
	}
}

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
	Id     uint64 `gorm:"column:id"`
	BizKey string `gorm:"column:biz_key"`

	BizName   string `gorm:"column:biz_name"`
	BizType   string `gorm:"column:biz_type"`
	BizSecret string `gorm:"column:biz_secret"`

	Contact      string `gorm:"column:contact"`
	ContactEmail string `gorm:"column:contact_email"`

	CreatorId uint64 `gorm:"column:creator_id"`

	CreatedAt int64 `gorm:"column:created_at"`
	UpdatedAt int64 `gorm:"column:updated_at"`

	IsDeleted bool  `gorm:"column:is_deleted"`
	DeletedAt int64 `gorm:"column:deleted_at"`
}

func (BizInfo) TableName() string {
	return "biz_info"
}

type BizInfoDao interface {
	Save(ctx context.Context, bizInfo BizInfo) (BizInfo, error)
	Update(ctx context.Context, bizInfo BizInfo) (BizInfo, error)
	DeleteInTx(ctx context.Context, tx *gorm.DB, id uint64) error

	Search(ctx context.Context, criteria search.BizSearchCriteria, param *pkggorm.PaginationParam) (*pkggorm.PaginationResult[BizInfo], error)
	FindById(ctx context.Context, id uint64) (BizInfo, error)
}

var _ BizInfoDao = (*DefaultBizInfoDao)(nil)

type DefaultBizInfoDao struct {
	db *gorm.DB
}

func (d *DefaultBizInfoDao) Save(ctx context.Context, bizInfo BizInfo) (BizInfo, error) {
	now := time.Now().UnixMilli()
	bizInfo.CreatedAt = now
	bizInfo.UpdatedAt = now

	err := d.db.WithContext(ctx).Model(&BizInfo{}).
		Clauses(clause.Returning{}).
		Create(&bizInfo).
		Scan(&bizInfo).Error
	return bizInfo, err
}

func (d *DefaultBizInfoDao) Update(ctx context.Context, bizInfo BizInfo) (BizInfo, error) {
	now := time.Now().UnixMilli()

	values := map[string]any{
		"updated_at": now,
	}

	if bizInfo.BizName != "" {
		values["biz_name"] = bizInfo.BizName
	}
	if bizInfo.Contact != "" {
		values["contact"] = bizInfo.Contact
	}
	if bizInfo.ContactEmail != "" {
		values["contact_email"] = bizInfo.ContactEmail
	}

	var res BizInfo
	err := d.db.WithContext(ctx).Model(&BizInfo{}).
		Clauses(clause.Returning{}). // Postgresql 才支持这个语法
		Where("id = ?", bizInfo.Id).
		Updates(values).
		Scan(&res).
		Error
	return res, err
}

// DeleteInTx 逻辑删除。
func (d *DefaultBizInfoDao) DeleteInTx(ctx context.Context, tx *gorm.DB, id uint64) error {
	now := time.Now().UnixMilli()

	return tx.WithContext(ctx).Model(&BizInfo{}).
		Where("id = ?", id).
		Updates(map[string]any{
			"is_deleted": true,
			"deleted_at": now,
		}).Error
}

func (d *DefaultBizInfoDao) Search(ctx context.Context, criteria search.BizSearchCriteria, param *pkggorm.PaginationParam) (*pkggorm.PaginationResult[BizInfo], error) {
	var records []BizInfo

	query := d.db.WithContext(ctx).Model(&BizInfo{}).Where("is_deleted = ?", false)

	if criteria.BizId != 0 {
		query.Where("biz_id = ?", criteria.BizId)
	}
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

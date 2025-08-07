package repository

import (
	"context"
	"fmt"
	"strings"

	"github.com/JrMarcco/easy-kit/slice"
	"github.com/JrMarcco/kuryr/internal/domain"
	"github.com/JrMarcco/kuryr/internal/errs"
	pkggorm "github.com/JrMarcco/kuryr/internal/pkg/gorm"
	"github.com/JrMarcco/kuryr/internal/repository/dao"
	"github.com/JrMarcco/kuryr/internal/search"
	"gorm.io/gorm"
)

type BizInfoRepo interface {
	Save(ctx context.Context, bizInfo domain.BizInfo) error
	Delete(ctx context.Context, id uint64) error
	DeleteInTx(ctx context.Context, tx *gorm.DB, id uint64) error

	Search(ctx context.Context, criteria search.BizSearchCriteria, param *pkggorm.PaginationParam) (*pkggorm.PaginationResult[domain.BizInfo], error)
	FindById(ctx context.Context, id uint64) (domain.BizInfo, error)
}

var _ BizInfoRepo = (*DefaultBizInfoRepo)(nil)

type DefaultBizInfoRepo struct {
	dao dao.BizInfoDao
}

func (r *DefaultBizInfoRepo) Save(ctx context.Context, bizInfo domain.BizInfo) error {
	err := r.dao.Save(ctx, r.toEntity(bizInfo))
	if err != nil {
		if pkggorm.IsUniqueConstraintError(err) {
			if strings.Contains(err.Error(), "biz_key") {
				return fmt.Errorf("%w: biz key [ %s ] already exists", errs.ErrInvalidParam, bizInfo.BizKey)
			}
		}
		return err
	}
	return nil
}

func (r *DefaultBizInfoRepo) Delete(ctx context.Context, id uint64) error {
	return r.dao.Delete(ctx, id)
}

func (r *DefaultBizInfoRepo) DeleteInTx(ctx context.Context, tx *gorm.DB, id uint64) error {
	return r.dao.DeleteInTx(ctx, tx, id)
}

func (r *DefaultBizInfoRepo) Search(
	ctx context.Context, criteria search.BizSearchCriteria, param *pkggorm.PaginationParam,
) (*pkggorm.PaginationResult[domain.BizInfo], error) {
	res, err := r.dao.Search(ctx, criteria, param)
	if err != nil {
		return nil, err
	}

	if res.Total == 0 {
		return pkggorm.NewPaginationResult([]domain.BizInfo{}, 0), nil
	}

	records := slice.Map(res.Records, func(idx int, src dao.BizInfo) domain.BizInfo {
		return r.toDomain(src)
	})

	return pkggorm.NewPaginationResult(records, res.Total), nil
}

func (r *DefaultBizInfoRepo) FindById(ctx context.Context, id uint64) (domain.BizInfo, error) {
	entity, err := r.dao.FindById(ctx, id)
	if err != nil {
		return domain.BizInfo{}, err
	}
	return r.toDomain(entity), nil
}

func (r *DefaultBizInfoRepo) toDomain(entity dao.BizInfo) domain.BizInfo {
	return domain.BizInfo{
		Id:           entity.Id,
		BizType:      domain.BizType(entity.BizType),
		BizKey:       entity.BizKey,
		BizSecret:    entity.BizSecret[:3] + "****" + entity.BizSecret[len(entity.BizSecret)-3:],
		BizName:      entity.BizName,
		Contact:      entity.Contact,
		ContactEmail: entity.ContactEmail,
		CreatedAt:    entity.CreatedAt,
		UpdatedAt:    entity.UpdatedAt,
	}
}

func (r *DefaultBizInfoRepo) toEntity(bi domain.BizInfo) dao.BizInfo {
	return dao.BizInfo{
		Id:           bi.Id,
		BizType:      string(bi.BizType),
		BizKey:       bi.BizKey,
		BizSecret:    bi.BizSecret,
		BizName:      bi.BizName,
		Contact:      bi.Contact,
		ContactEmail: bi.ContactEmail,
		CreatedAt:    bi.CreatedAt,
		UpdatedAt:    bi.UpdatedAt,
	}
}

func NewDefaultBizInfoRepo(dao dao.BizInfoDao) *DefaultBizInfoRepo {
	return &DefaultBizInfoRepo{
		dao: dao,
	}
}

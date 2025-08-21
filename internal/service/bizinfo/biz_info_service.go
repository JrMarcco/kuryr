package bizinfo

import (
	"context"
	"fmt"

	"github.com/JrMarcco/kuryr/internal/domain"
	"github.com/JrMarcco/kuryr/internal/errs"
	pkggorm "github.com/JrMarcco/kuryr/internal/pkg/gorm"
	"github.com/JrMarcco/kuryr/internal/pkg/secret"
	"github.com/JrMarcco/kuryr/internal/repository"
	"github.com/JrMarcco/kuryr/internal/search"
	"gorm.io/gorm"
)

type Service interface {
	Save(ctx context.Context, bizInfo domain.BizInfo) (domain.BizInfo, error)
	Update(ctx context.Context, bizInfo domain.BizInfo) (domain.BizInfo, error)
	Delete(ctx context.Context, id uint64) error

	Search(ctx context.Context, criteria search.BizSearchCriteria, param *pkggorm.PaginationParam) (*pkggorm.PaginationResult[domain.BizInfo], error)
	FindById(ctx context.Context, id uint64) (domain.BizInfo, error)
}

var _ Service = (*DefaultService)(nil)

type DefaultService struct {
	db *gorm.DB // db 数据库连接，用于开启事务

	generator secret.Generator

	bizInfoRepo   repository.BizInfoRepo
	bizConfigRepo repository.BizConfigRepo
}

func (s *DefaultService) Save(ctx context.Context, bizInfo domain.BizInfo) (domain.BizInfo, error) {
	bizSecret, err := s.generator.Generate(32)
	if err != nil {
		return domain.BizInfo{}, err
	}
	bizInfo.BizSecret = bizSecret

	return s.bizInfoRepo.Save(ctx, bizInfo)
}

func (s *DefaultService) Update(ctx context.Context, bizInfo domain.BizInfo) (domain.BizInfo, error) {
	return s.bizInfoRepo.Update(ctx, bizInfo)
}

func (s *DefaultService) Delete(ctx context.Context, id uint64) error {
	bizInfo, err := s.bizInfoRepo.FindById(ctx, id)
	if err != nil {
		return err
	}

	if !s.canDelete(ctx, bizInfo) {
		return fmt.Errorf("%w: biz info is not allowed to delete", errs.ErrInvalidStatus)
	}

	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var txErr error

		// 先删除配置。
		txErr = s.bizConfigRepo.DeleteInTx(ctx, tx, id)
		if txErr != nil {
			return txErr
		}

		txErr = s.bizInfoRepo.DeleteInTx(ctx, tx, id)
		return txErr
	})
}

// canDelete 判断当前业务是否允许删除。
func (s *DefaultService) canDelete(ctx context.Context, bizInfo domain.BizInfo) bool {
	// TODO: implement me
	panic("implement me")
}

func (s *DefaultService) Search(ctx context.Context, criteria search.BizSearchCriteria, param *pkggorm.PaginationParam) (*pkggorm.PaginationResult[domain.BizInfo], error) {
	return s.bizInfoRepo.Search(ctx, criteria, param)
}

func (s *DefaultService) FindById(ctx context.Context, id uint64) (domain.BizInfo, error) {
	return s.bizInfoRepo.FindById(ctx, id)
}

func NewDefaultService(
	db *gorm.DB, generator secret.Generator, bizInfoRepo repository.BizInfoRepo, bizConfigRepo repository.BizConfigRepo,
) *DefaultService {
	return &DefaultService{
		db:            db,
		generator:     generator,
		bizInfoRepo:   bizInfoRepo,
		bizConfigRepo: bizConfigRepo,
	}
}

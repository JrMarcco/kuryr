package bizinfo

import (
	"context"

	"github.com/JrMarcco/kuryr/internal/domain"
	pkggorm "github.com/JrMarcco/kuryr/internal/pkg/gorm"
	"github.com/JrMarcco/kuryr/internal/repository"
	"github.com/JrMarcco/kuryr/internal/search"
	"gorm.io/gorm"
)

type Service interface {
	Save(ctx context.Context, bizInfo domain.BizInfo) error
	Delete(ctx context.Context, id uint64) error

	Search(ctx context.Context, criteria search.BizSearchCriteria, param *pkggorm.PaginationParam) (*pkggorm.PaginationResult[domain.BizInfo], error)
	FindById(ctx context.Context, id uint64) (domain.BizInfo, error)
}

var _ Service = (*DefaultService)(nil)

type DefaultService struct {
	db *gorm.DB // db 数据库连接，用于开启事务

	infoRepo   repository.BizInfoRepo
	configRepo repository.BizConfigRepo
}

func (s *DefaultService) Save(ctx context.Context, bizInfo domain.BizInfo) error {
	return s.infoRepo.Save(ctx, bizInfo)
}

func (s *DefaultService) Delete(ctx context.Context, id uint64) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var txErr error

		// 先删除配置
		txErr = s.configRepo.DeleteInTx(ctx, tx, id)
		if txErr != nil {
			return txErr
		}

		// 在删除
		txErr = s.infoRepo.DeleteInTx(ctx, tx, id)
		return txErr
	})
}

func (s *DefaultService) Search(ctx context.Context, criteria search.BizSearchCriteria, param *pkggorm.PaginationParam) (*pkggorm.PaginationResult[domain.BizInfo], error) {
	return s.infoRepo.Search(ctx, criteria, param)
}

func (s *DefaultService) FindById(ctx context.Context, id uint64) (domain.BizInfo, error) {
	return s.infoRepo.FindById(ctx, id)
}

func NewDefaultService(
	db *gorm.DB, infoRepo repository.BizInfoRepo, configRepo repository.BizConfigRepo,
) *DefaultService {
	return &DefaultService{
		db:         db,
		infoRepo:   infoRepo,
		configRepo: configRepo,
	}
}

package bizconf

import (
	"context"
	"fmt"

	"github.com/JrMarcco/kuryr/internal/domain"
	"github.com/JrMarcco/kuryr/internal/errs"
	"github.com/JrMarcco/kuryr/internal/repository"
)

//go:generate mockgen -source=./biz_config_service.go -destination=./mock/biz_config_service.mock.go -package=bizconfmock -typed Service

type Service interface {
	Save(ctx context.Context, bizConfig domain.BizConfig) (domain.BizConfig, error)
	Update(ctx context.Context, bizConfig domain.BizConfig) (domain.BizConfig, error)
	FindByBizId(ctx context.Context, id uint64) (domain.BizConfig, error)
}

var _ Service = (*DefaultService)(nil)

type DefaultService struct {
	bizInfoRepo   repository.BizInfoRepo
	bizConfigRepo repository.BizConfigRepo
}

func (s *DefaultService) Save(ctx context.Context, bizConfig domain.BizConfig) (domain.BizConfig, error) {
	if bizConfig.BizId == 0 {
		return domain.BizConfig{}, fmt.Errorf("%w: invalidate biz id [ %d ]", errs.ErrInvalidParam, bizConfig.Id)
	}

	bizInfo, err := s.bizInfoRepo.FindById(ctx, bizConfig.BizId)
	if err != nil {
		return domain.BizConfig{}, fmt.Errorf("%w: failed to find biz info by id [ %d ]", errs.ErrInvalidParam, bizConfig.Id)
	}

	bizConfig.OwnerType = domain.OwnerType(bizInfo.BizType)
	return s.bizConfigRepo.Save(ctx, bizConfig)
}

func (s *DefaultService) Update(ctx context.Context, bizConfig domain.BizConfig) (domain.BizConfig, error) {
	if bizConfig.Id == 0 {
		return domain.BizConfig{}, fmt.Errorf("%w: invalidate biz config id [ %d ]", errs.ErrInvalidParam, bizConfig.Id)
	}

	return s.bizConfigRepo.Update(ctx, bizConfig)
}

func (s *DefaultService) FindByBizId(ctx context.Context, bizId uint64) (domain.BizConfig, error) {
	if bizId == 0 {
		return domain.BizConfig{}, fmt.Errorf("%w: invalidate biz id [ %d ]", errs.ErrInvalidParam, bizId)
	}
	return s.bizConfigRepo.FindByBizId(ctx, bizId)
}

func NewDefaultService(bizInfoRepo repository.BizInfoRepo, bizConfigRepo repository.BizConfigRepo) *DefaultService {
	return &DefaultService{
		bizInfoRepo:   bizInfoRepo,
		bizConfigRepo: bizConfigRepo,
	}
}

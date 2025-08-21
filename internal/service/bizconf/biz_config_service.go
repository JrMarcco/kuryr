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
	Delete(ctx context.Context, id uint64) error
	FindById(ctx context.Context, id uint64) (domain.BizConfig, error)
}

var _ Service = (*DefaultService)(nil)

type DefaultService struct {
	bizInfoRepo   repository.BizInfoRepo
	bizConfigRepo repository.BizConfigRepo
}

func (s *DefaultService) Save(ctx context.Context, bizConfig domain.BizConfig) (domain.BizConfig, error) {
	if bizConfig.Id == 0 {
		return domain.BizConfig{}, fmt.Errorf("%w: invalidate biz config id [ %d ]", errs.ErrInvalidParam, bizConfig.Id)
	}

	bizInfo, err := s.bizInfoRepo.FindById(ctx, bizConfig.Id)
	if err != nil {
		return domain.BizConfig{}, fmt.Errorf("%w: failed to find biz info by id [ %d ]", errs.ErrInvalidParam, bizConfig.Id)
	}

	bizConfig.OwnerType = domain.OwnerType(bizInfo.BizType)
	return s.bizConfigRepo.Save(ctx, bizConfig)
}

func (s *DefaultService) Delete(ctx context.Context, id uint64) error {
	if id == 0 {
		return fmt.Errorf("%w: invalidate biz config id [ %d ]", errs.ErrInvalidParam, id)
	}
	return s.bizConfigRepo.Delete(ctx, id)
}

func (s *DefaultService) FindById(ctx context.Context, id uint64) (domain.BizConfig, error) {
	if id == 0 {
		return domain.BizConfig{}, fmt.Errorf("%w: invalidate biz config id [ %d ]", errs.ErrInvalidParam, id)
	}
	return s.bizConfigRepo.FindById(ctx, id)
}

func NewDefaultService(bizInfoRepo repository.BizInfoRepo, bizConfigRepo repository.BizConfigRepo) *DefaultService {
	return &DefaultService{
		bizInfoRepo:   bizInfoRepo,
		bizConfigRepo: bizConfigRepo,
	}
}

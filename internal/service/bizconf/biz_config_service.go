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
	Save(ctx context.Context, bizConfig domain.BizConfig) error
	Delete(ctx context.Context, id uint64) error
	FindById(ctx context.Context, id uint64) (domain.BizConfig, error)
}

var _ Service = (*DefaultService)(nil)

type DefaultService struct {
	repo repository.BizConfigRepo
}

func (s *DefaultService) Save(ctx context.Context, bizConfig domain.BizConfig) error {
	if bizConfig.Id == 0 {
		return fmt.Errorf("%w: invalidate biz config id [ %d ]", errs.ErrInvalidParam, bizConfig.Id)
	}
	return s.repo.Save(ctx, bizConfig)
}

func (s *DefaultService) Delete(ctx context.Context, id uint64) error {
	if id == 0 {
		return fmt.Errorf("%w: invalidate biz config id [ %d ]", errs.ErrInvalidParam, id)
	}
	return s.repo.Delete(ctx, id)
}

func (s *DefaultService) FindById(ctx context.Context, id uint64) (domain.BizConfig, error) {
	if id == 0 {
		return domain.BizConfig{}, fmt.Errorf("%w: invalidate biz config id [ %d ]", errs.ErrInvalidParam, id)
	}
	return s.repo.FindById(ctx, id)
}

func NewDefaultService(repo repository.BizConfigRepo) *DefaultService {
	return &DefaultService{
		repo: repo,
	}
}

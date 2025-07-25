package bizconf

import (
	"context"
	"errors"

	"github.com/JrMarcco/kuryr/internal/domain"
	"github.com/JrMarcco/kuryr/internal/repository"
)

//go:generate mockgen -source=./biz_config_service.go -destination=./mock/biz_config_service.mock.go -package=bizconfmock -typed BizConfigService

var ErrInvalidBizId = errors.New("[kuryr] biz id (owner_id) not set")

type BizConfigService interface {
	Save(ctx context.Context, bizConfig domain.BizConfig) error
	Delete(ctx context.Context, id uint64) error
	GetById(ctx context.Context, id uint64) (domain.BizConfig, error)
}

var _ BizConfigService = (*DefaultBizConfigService)(nil)

type DefaultBizConfigService struct {
	repo repository.BizConfigRepo
}

func (s *DefaultBizConfigService) Save(ctx context.Context, bizConfig domain.BizConfig) error {
	if bizConfig.Id == 0 {
		return ErrInvalidBizId
	}
	return s.repo.Save(ctx, bizConfig)
}

func (s *DefaultBizConfigService) Delete(ctx context.Context, id uint64) error {
	return s.repo.Delete(ctx, id)
}

func (s *DefaultBizConfigService) GetById(ctx context.Context, id uint64) (domain.BizConfig, error) {
	return s.repo.GetById(ctx, id)
}

func NewDefaultBizConfigService(repo repository.BizConfigRepo) *DefaultBizConfigService {
	return &DefaultBizConfigService{
		repo: repo,
	}
}

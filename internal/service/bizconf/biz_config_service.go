package bizconf

import (
	"context"
	"errors"

	"github.com/JrMarcco/kuryr/internal/domain"
	"github.com/JrMarcco/kuryr/internal/repository"
)

//go:generate mockgen -source=./biz_config_service.go -destination=./mock/biz_config_service.mock.go -package=bizconfmock -typed BizConfigService

var ErrInvalidBizId = errors.New("[kuryr] biz id (owner_id) not set")

type Service interface {
	Save(ctx context.Context, bizConfig domain.BizConfig) error
	Delete(ctx context.Context, id uint64) error
	GetById(ctx context.Context, id uint64) (domain.BizConfig, error)
}

var _ Service = (*DefaultService)(nil)

type DefaultService struct {
	repo repository.BizConfigRepo
}

func (s *DefaultService) Save(ctx context.Context, bizConfig domain.BizConfig) error {
	if bizConfig.Id == 0 {
		return ErrInvalidBizId
	}
	return s.repo.Save(ctx, bizConfig)
}

func (s *DefaultService) Delete(ctx context.Context, id uint64) error {
	return s.repo.Delete(ctx, id)
}

func (s *DefaultService) GetById(ctx context.Context, id uint64) (domain.BizConfig, error) {
	return s.repo.GetById(ctx, id)
}

func NewDefaultService(repo repository.BizConfigRepo) *DefaultService {
	return &DefaultService{
		repo: repo,
	}
}

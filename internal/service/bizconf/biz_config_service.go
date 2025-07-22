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
}

var _ BizConfigService = (*DefaultBizConfigService)(nil)

type DefaultBizConfigService struct {
	repo repository.BizConfigRepo
}

func (s *DefaultBizConfigService) Save(ctx context.Context, bizConfig domain.BizConfig) error {
	if bizConfig.OwnerId <= 0 {
		return ErrInvalidBizId
	}
	return s.repo.Save(ctx, bizConfig)
}

func NewDefaultBizConfigService(repo repository.BizConfigRepo) *DefaultBizConfigService {
	return &DefaultBizConfigService{
		repo: repo,
	}
}

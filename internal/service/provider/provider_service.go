package provider

import (
	"context"
	"fmt"

	"github.com/JrMarcco/kuryr/internal/domain"
	"github.com/JrMarcco/kuryr/internal/errs"
	pkggorm "github.com/JrMarcco/kuryr/internal/pkg/gorm"
	"github.com/JrMarcco/kuryr/internal/repository"
	"github.com/JrMarcco/kuryr/internal/search"
)

type Service interface {
	Save(ctx context.Context, provider domain.Provider) error
	Delete(ctx context.Context, id uint64) error
	Update(ctx context.Context, provider domain.Provider) error

	Search(ctx context.Context, criteria search.ProviderCriteria, param *pkggorm.PaginationParam) (*pkggorm.PaginationResult[domain.Provider], error)
	FindById(ctx context.Context, id uint64) (domain.Provider, error)
	FindByChannel(ctx context.Context, channel domain.Channel) ([]domain.Provider, error)
}

var _ Service = (*DefaultService)(nil)

type DefaultService struct {
	repo repository.ProviderRepo
}

func (s *DefaultService) Save(ctx context.Context, provider domain.Provider) error {
	if err := provider.Validate(); err != nil {
		return err
	}
	// 默认状态设置为“未启用”
	provider.ActiveStatus = domain.ActiveStatusInactive
	return s.repo.Save(ctx, provider)
}

func (s *DefaultService) Delete(ctx context.Context, id uint64) error {
	if id == 0 {
		return fmt.Errorf("%w: invalid provider id [ %d ]", errs.ErrInvalidParam, id)
	}

	provider, err := s.FindById(ctx, id)
	if err != nil {
		return fmt.Errorf("%w: cannot find provider", errs.ErrRecordNotFound)
	}

	if err := s.canDelete(provider); err != nil {
		return err
	}
	return s.repo.Delete(ctx, id)
}

// TODO
// canDelete 判断 provider 是否允许删除
func (s *DefaultService) canDelete(provider domain.Provider) error {
	if provider.ActiveStatus == domain.ActiveStatusActive {
		return fmt.Errorf("%w: provider is active, can not delete", errs.ErrInvalidStatus)
	}
	return nil
}

func (s *DefaultService) Update(ctx context.Context, provider domain.Provider) error {
	if err := provider.Validate(); err != nil {
		return err
	}
	return s.repo.Update(ctx, provider)
}

func (s *DefaultService) Search(ctx context.Context, criteria search.ProviderCriteria, param *pkggorm.PaginationParam) (*pkggorm.PaginationResult[domain.Provider], error) {
	return s.repo.Search(ctx, criteria, param)
}

func (s *DefaultService) FindById(ctx context.Context, id uint64) (domain.Provider, error) {
	if id == 0 {
		return domain.Provider{}, fmt.Errorf("%w: invalid provider id [ %d ]", errs.ErrInvalidParam, id)
	}
	return s.repo.FindById(ctx, id)
}

func (s *DefaultService) FindByChannel(ctx context.Context, channel domain.Channel) ([]domain.Provider, error) {
	if !channel.IsValid() {
		return nil, fmt.Errorf("%w: invalid provider channel [ %s ]", errs.ErrInvalidParam, string(channel))
	}
	return s.repo.FindByChannel(ctx, string(channel))
}

func NewDefaultService(repo repository.ProviderRepo) *DefaultService {
	return &DefaultService{
		repo: repo,
	}
}

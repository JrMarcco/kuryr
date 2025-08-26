package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/JrMarcco/easy-kit/slice"
	"github.com/JrMarcco/kuryr/internal/domain"
	"github.com/JrMarcco/kuryr/internal/errs"
	"github.com/JrMarcco/kuryr/internal/repository/dao"
	"gorm.io/gorm"
)

type ProviderRepo interface {
	Save(ctx context.Context, provider domain.Provider) (domain.Provider, error)
	Delete(ctx context.Context, id uint64) error
	Update(ctx context.Context, provider domain.Provider) (domain.Provider, error)

	List(ctx context.Context) ([]domain.Provider, error)
	FindById(ctx context.Context, id uint64) (domain.Provider, error)
	FindByChannel(ctx context.Context, channel string) ([]domain.Provider, error)
}

var _ ProviderRepo = (*DefaultProviderRepo)(nil)

type DefaultProviderRepo struct {
	dao dao.ProviderDao
}

func (r *DefaultProviderRepo) Save(ctx context.Context, provider domain.Provider) (domain.Provider, error) {
	entity, err := r.dao.Save(ctx, r.toEntity(provider))
	if err != nil {
		return domain.Provider{}, err
	}
	return r.toDomain(entity), nil
}

func (r *DefaultProviderRepo) Delete(ctx context.Context, id uint64) error {
	return r.dao.Delete(ctx, id)
}

func (r *DefaultProviderRepo) Update(ctx context.Context, provider domain.Provider) (domain.Provider, error) {
	entity, err := r.dao.Update(ctx, r.toEntity(provider))
	if err != nil {
		return domain.Provider{}, err
	}
	return r.toDomain(entity), nil
}

func (r *DefaultProviderRepo) List(ctx context.Context) ([]domain.Provider, error) {
	entities, err := r.dao.List(ctx)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("%w: cannot find provider", errs.ErrRecordNotFound)
		}
		return nil, err
	}
	return slice.Map(entities, func(_ int, src dao.Provider) domain.Provider {
		return r.toDomain(src)
	}), nil
}

func (r *DefaultProviderRepo) FindById(ctx context.Context, id uint64) (domain.Provider, error) {
	entity, err := r.dao.FindById(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return domain.Provider{}, fmt.Errorf("%w: cannot find provider", errs.ErrRecordNotFound)
		}
		return domain.Provider{}, err
	}
	return r.toDomain(entity), nil
}

func (r *DefaultProviderRepo) FindByChannel(ctx context.Context, channel string) ([]domain.Provider, error) {
	entities, err := r.dao.FindByChannel(ctx, channel)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("%w: cannot find provider", errs.ErrRecordNotFound)
		}
		return nil, err
	}
	return slice.Map(entities, func(_ int, src dao.Provider) domain.Provider {
		return r.toDomain(src)
	}), nil
}

func (r *DefaultProviderRepo) toEntity(provider domain.Provider) dao.Provider {
	return dao.Provider{
		Id:               provider.Id,
		ProviderName:     provider.ProviderName,
		Channel:          int32(provider.Channel),
		Endpoint:         provider.Endpoint,
		AppId:            provider.AppId,
		ApiKey:           provider.ApiKey,
		ApiSecret:        provider.ApiSecret,
		Weight:           provider.Weight,
		QpsLimit:         provider.QpsLimit,
		DailyLimit:       provider.DailyLimit,
		AuditCallbackUrl: provider.AuditCallbackUrl,
		ActiveStatus:     string(provider.ActiveStatus),
	}
}

func (r *DefaultProviderRepo) toDomain(entity dao.Provider) domain.Provider {
	return domain.Provider{
		Id:               entity.Id,
		ProviderName:     entity.ProviderName,
		Channel:          domain.Channel(entity.Channel),
		Endpoint:         entity.Endpoint,
		AppId:            entity.AppId,
		ApiKey:           entity.ApiKey,
		ApiSecret:        entity.ApiSecret,
		Weight:           entity.Weight,
		QpsLimit:         entity.QpsLimit,
		DailyLimit:       entity.DailyLimit,
		AuditCallbackUrl: entity.AuditCallbackUrl,
		ActiveStatus:     domain.ActiveStatus(entity.ActiveStatus),
	}
}

func NewDefaultProviderRepo(dao dao.ProviderDao) *DefaultProviderRepo {
	return &DefaultProviderRepo{
		dao: dao,
	}
}

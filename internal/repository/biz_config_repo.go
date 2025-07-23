package repository

import (
	"context"

	"github.com/JrMarcco/kuryr/internal/domain"
	"github.com/JrMarcco/kuryr/internal/pkg/xsql"
	"github.com/JrMarcco/kuryr/internal/repository/cache"
	"github.com/JrMarcco/kuryr/internal/repository/dao"
	"go.uber.org/zap"
)

type BizConfigRepo interface {
	Save(ctx context.Context, bizConfig domain.BizConfig) error
}

var _ BizConfigRepo = (*DefaultBizConfigRepo)(nil)

type DefaultBizConfigRepo struct {
	dao        dao.BizConfigDao
	localCache cache.BizConfigCache
	redisCache cache.BizConfigCache

	logger *zap.Logger
}

func (r *DefaultBizConfigRepo) Save(ctx context.Context, bizConfig domain.BizConfig) error {
	entity, err := r.dao.SaveOrUpdate(ctx, r.toEntity(bizConfig))
	if err != nil {
		return err
	}

	d := r.toDomain(entity)
	err = r.redisCache.Set(ctx, d)
	if err != nil {
		r.logger.Error("[biz config] failed to set biz config to redis cache", zap.Error(err))
	}

	// 注意：
	// 	如果通过 配置中心 / MQ 来同步本地缓存
	// 	这里则需要 更新配置中心 / 发送 MQ 消息
	err = r.localCache.Set(ctx, d)
	if err != nil {
		r.logger.Error("[biz config] failed to set biz config to local cache", zap.Error(err))
	}
	return nil
}

func (r *DefaultBizConfigRepo) toEntity(bizConfig domain.BizConfig) dao.BizConfig {
	entity := dao.BizConfig{
		Id:        bizConfig.Id,
		RateLimit: bizConfig.RateLimit,
		CreatedAt: bizConfig.CreatedAt,
		UpdatedAt: bizConfig.UpdatedAt,
	}

	if bizConfig.ChannelConfig != nil {
		entity.ChannelConfig = xsql.JsonColumn[domain.ChannelConfig]{
			Val:   *bizConfig.ChannelConfig,
			Valid: true,
		}
	}

	if bizConfig.QuotaConfig != nil {
		entity.QuotaConfig = xsql.JsonColumn[domain.QuotaConfig]{
			Val:   *bizConfig.QuotaConfig,
			Valid: true,
		}
	}

	if bizConfig.CallbackConfig != nil {
		entity.CallbackConfig = xsql.JsonColumn[domain.CallbackConfig]{
			Val:   *bizConfig.CallbackConfig,
			Valid: true,
		}
	}

	return entity
}

func (r *DefaultBizConfigRepo) toDomain(entity dao.BizConfig) domain.BizConfig {
	bizConfig := domain.BizConfig{
		Id:        entity.Id,
		RateLimit: entity.RateLimit,
		CreatedAt: entity.CreatedAt,
		UpdatedAt: entity.UpdatedAt,
	}

	if entity.ChannelConfig.Valid {
		bizConfig.ChannelConfig = &entity.ChannelConfig.Val
	}

	if entity.QuotaConfig.Valid {
		bizConfig.QuotaConfig = &entity.QuotaConfig.Val
	}

	if entity.CallbackConfig.Valid {
		bizConfig.CallbackConfig = &entity.CallbackConfig.Val
	}

	return bizConfig
}

func NewDefaultBizConfigRepo(
	dao dao.BizConfigDao, localCache cache.BizConfigCache, redisCache cache.BizConfigCache, logger *zap.Logger,
) *DefaultBizConfigRepo {
	return &DefaultBizConfigRepo{
		dao:        dao,
		localCache: localCache,
		redisCache: redisCache,
		logger:     logger,
	}
}

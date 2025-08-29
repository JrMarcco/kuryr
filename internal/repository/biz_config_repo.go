package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/JrMarcco/kuryr/internal/domain"
	"github.com/JrMarcco/kuryr/internal/errs"
	pkgsql "github.com/JrMarcco/kuryr/internal/pkg/sql"
	"github.com/JrMarcco/kuryr/internal/repository/cache"
	"github.com/JrMarcco/kuryr/internal/repository/dao"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type BizConfigRepo interface {
	Save(ctx context.Context, bizConfig domain.BizConfig) (domain.BizConfig, error)
	Update(ctx context.Context, bizConfig domain.BizConfig) (domain.BizConfig, error)
	FindByBizId(ctx context.Context, bizId uint64) (domain.BizConfig, error)

	DeleteInTx(ctx context.Context, tx *gorm.DB, id uint64) error
}

var _ BizConfigRepo = (*DefaultBizConfigRepo)(nil)

type DefaultBizConfigRepo struct {
	dao        dao.BizConfigDao
	localCache cache.BizConfigCache
	redisCache cache.BizConfigCache

	logger *zap.Logger
}

func (r *DefaultBizConfigRepo) Save(ctx context.Context, bizConfig domain.BizConfig) (domain.BizConfig, error) {
	entity, err := r.dao.Save(ctx, r.toEntity(bizConfig))
	if err != nil {
		return domain.BizConfig{}, err
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
	return d, nil
}

func (r *DefaultBizConfigRepo) Update(ctx context.Context, bizConfig domain.BizConfig) (domain.BizConfig, error) {
	entity, err := r.dao.Update(ctx, r.toEntity(bizConfig))
	if err != nil {
		return domain.BizConfig{}, err
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

	return d, nil
}

func (r *DefaultBizConfigRepo) FindByBizId(ctx context.Context, id uint64) (domain.BizConfig, error) {
	// 从本地缓存获取
	bizConfig, err := r.localCache.Get(ctx, id)
	if err == nil {
		return bizConfig, nil
	}

	// 从 redis 获取
	bizConfig, err = r.redisCache.Get(ctx, id)
	if err == nil {
		// 设置本地缓存
		err = r.localCache.Set(ctx, bizConfig)
		if err != nil {
			r.logger.Error("[biz config] failed to set biz config to local cache", zap.Error(err))
		}
		return bizConfig, nil
	}

	// TODO: 如果这里触发熔断、降级可以直接返回

	// 从 db 获取并设置 redis 缓存 + 本地缓存
	entity, err := r.dao.FindByBizId(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return domain.BizConfig{}, fmt.Errorf("%w: cannot find biz config, id = %d", errs.ErrRecordNotFound, id)
		}
		return domain.BizConfig{}, err
	}

	bizConfig = r.toDomain(entity)
	// 设置 redis 缓存
	err = r.redisCache.Set(ctx, bizConfig)
	if err != nil {
		r.logger.Error("[biz config] failed to set biz config to redis cache", zap.Error(err))
	}
	// 设置本地缓存
	err = r.localCache.Set(ctx, bizConfig)
	if err != nil {
		r.logger.Error("[biz config] failed to set biz config to local cache", zap.Error(err))
	}
	return bizConfig, nil
}

func (r *DefaultBizConfigRepo) DeleteInTx(ctx context.Context, tx *gorm.DB, id uint64) error {
	// 删数据库
	if err := r.dao.DeleteInTx(ctx, tx, id); err != nil {
		return err
	}

	r.clearCache(ctx, id)
	return nil
}

func (r *DefaultBizConfigRepo) clearCache(ctx context.Context, id uint64) {
	// 删 redis 缓存
	if err := r.redisCache.Del(ctx, id); err != nil {
		r.logger.Error("[biz config] failed to del biz config from redis cache", zap.Error(err))
	}
	// 删本地缓存
	if err := r.localCache.Del(ctx, id); err != nil {
		r.logger.Error("[biz config] failed to del biz config from local cache", zap.Error(err))
	}
}

func (r *DefaultBizConfigRepo) toEntity(bizConfig domain.BizConfig) dao.BizConfig {
	entity := dao.BizConfig{
		Id:        bizConfig.Id,
		BizId:     bizConfig.BizId,
		OwnerType: string(bizConfig.OwnerType),
		RateLimit: bizConfig.RateLimit,
	}

	if bizConfig.ChannelConfig != nil {
		entity.ChannelConfig = pkgsql.JsonColumn[domain.ChannelConfig]{
			Val:   *bizConfig.ChannelConfig,
			Valid: true,
		}
	}

	if bizConfig.QuotaConfig != nil {
		entity.QuotaConfig = pkgsql.JsonColumn[domain.QuotaConfig]{
			Val:   *bizConfig.QuotaConfig,
			Valid: true,
		}
	}

	if bizConfig.CallbackConfig != nil {
		entity.CallbackConfig = pkgsql.JsonColumn[domain.CallbackConfig]{
			Val:   *bizConfig.CallbackConfig,
			Valid: true,
		}
	}

	return entity
}

func (r *DefaultBizConfigRepo) toDomain(entity dao.BizConfig) domain.BizConfig {
	bizConfig := domain.BizConfig{
		Id:        entity.Id,
		BizId:     entity.BizId,
		OwnerType: domain.OwnerType(entity.OwnerType),
		RateLimit: entity.RateLimit,
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

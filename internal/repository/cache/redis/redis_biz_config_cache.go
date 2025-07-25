package redis

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/JrMarcco/kuryr/internal/domain"
	"github.com/JrMarcco/kuryr/internal/repository/cache"
	"github.com/redis/go-redis/v9"
)

var _ cache.BizConfigCache = (*RBizConfigCache)(nil)

type RBizConfigCache struct {
	rc redis.Cmdable
}

func (c *RBizConfigCache) Set(ctx context.Context, bizConfig domain.BizConfig) error {
	data, err := json.Marshal(bizConfig)
	if err != nil {
		return fmt.Errorf("[kuryr] failed to marshal biz config to json: %w", err)
	}

	key := cache.BizConfigCacheKey(bizConfig.Id)
	err = c.rc.Set(ctx, key, data, cache.BizConfigDefaultLocalExp).Err()
	if err != nil {
		return fmt.Errorf("[kuryr] failed to set biz config to redis: %w", err)
	}
	return nil
}

func (c *RBizConfigCache) Get(ctx context.Context, id uint64) (domain.BizConfig, error) {
	key := cache.BizConfigCacheKey(id)
	str, err := c.rc.Get(ctx, key).Result()
	if err != nil {
		return domain.BizConfig{}, fmt.Errorf("[kuryr] failed to get biz config from redis: %w", err)
	}
	var bizConfig domain.BizConfig
	err = json.Unmarshal([]byte(str), &bizConfig)
	if err != nil {
		return domain.BizConfig{}, fmt.Errorf("[kuryr] failed to unmarshal biz config from redis: %w", err)
	}
	return bizConfig, nil
}

func (c *RBizConfigCache) Del(ctx context.Context, id uint64) error {
	return c.rc.Del(ctx, cache.BizConfigCacheKey(id)).Err()
}

func NewRBizConfigCache(rc redis.Cmdable) *RBizConfigCache {
	return &RBizConfigCache{
		rc: rc,
	}
}

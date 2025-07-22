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

func (r *RBizConfigCache) Set(ctx context.Context, bizConfig domain.BizConfig) error {
	key := cache.BizConfigCacheKey(bizConfig.Id)

	data, err := json.Marshal(bizConfig)
	if err != nil {
		return fmt.Errorf("[kuryr] failed to marshal biz config to json: %w", err)
	}

	err = r.rc.Set(ctx, key, data, cache.BizConfigDefaultLocalExp).Err()
	if err != nil {
		return fmt.Errorf("[kuryr] failed to set biz config to redis: %w", err)
	}
	return nil
}

func NewRBizConfigCache(rc redis.Cmdable) *RBizConfigCache {
	return &RBizConfigCache{
		rc: rc,
	}
}

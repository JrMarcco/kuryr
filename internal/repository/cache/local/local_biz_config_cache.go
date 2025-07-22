package local

import (
	"context"
	"fmt"
	"strings"

	"github.com/JrMarcco/kuryr/internal/domain"
	"github.com/JrMarcco/kuryr/internal/repository/cache"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"

	gocache "github.com/patrickmn/go-cache"
)

var _ cache.BizConfigCache = (*LBizConfigCache)(nil)

type LBizConfigCache struct {
	rc     redis.Cmdable
	cc     *gocache.Cache
	logger *zap.Logger
}

func (l *LBizConfigCache) Set(_ context.Context, bizConfig domain.BizConfig) error {
	key := cache.BizConfigCacheKey(bizConfig.Id)
	l.cc.Set(key, bizConfig, cache.BizConfigDefaultLocalExp)
	return nil
}

// watchRedis 监听 redis 实时更新本地缓存
func (l *LBizConfigCache) watchRedis(ctx context.Context) {
	redisClient, ok := l.rc.(*redis.Client)
	if !ok {
		l.logger.Error("[biz config] failed to cast redis client to redis.Client")
		return
	}

	watchKey := fmt.Sprintf("%s*", cache.BizConfigCacheKeyPrefix)
	pubSub := redisClient.Subscribe(ctx, fmt.Sprintf("__keyspace@*__:%s", watchKey))
	defer func(pubSub *redis.PubSub) {
		err := pubSub.Close()
		if err != nil {
			l.logger.Error("[biz config] failed to close pub sub", zap.Error(err))
		}
	}(pubSub)

	l.logger.Info("[biz config] start watching redis keyspace event", zap.String("pattern", watchKey))

	ch := pubSub.Channel()
	for {
		select {
		case msg := <-ch:
			if msg == nil {
				return
			}
			l.handleKeyChange(ctx, msg)
		case <-ctx.Done():
			l.logger.Info("[biz config] stop watching redis keyspace event", zap.String("pattern", "biz_config*"))
			return
		}
	}
}

func (l *LBizConfigCache) handleKeyChange(ctx context.Context, msg *redis.Message) {
	// 解析消息
	// msg.Channel 格式: __keyspace@0__:biz_config_xxx
	// msg.Payload 是操作类型: set, del, expire 等
	parts := strings.Split(msg.Channel, ":")
	if len(parts) < 2 {
		return
	}

	key := parts[1]
	op := msg.Payload

	l.logger.Info("[biz config] redis keyspace event", zap.String("key", key), zap.String("operation", op))

	switch op {
	case "set":
		val, err := l.rc.Get(ctx, key).Result()
		if err != nil {
			l.logger.Error("[biz config] failed to get biz config from redis", zap.String("key", key), zap.Error(err))
			return
		}
		// 更新本地缓存
		l.cc.Set(key, val, cache.BizConfigDefaultLocalExp)
	case "del":
		l.cc.Delete(key)
	}
}

func NewLBizConfigCache(rc redis.Cmdable, cc *gocache.Cache, logger *zap.Logger) *LBizConfigCache {
	bizConfigCache := &LBizConfigCache{
		rc:     rc,
		cc:     cc,
		logger: logger,
	}

	go bizConfigCache.watchRedis(context.Background())

	return bizConfigCache
}

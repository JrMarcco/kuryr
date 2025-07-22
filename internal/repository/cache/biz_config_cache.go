package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/JrMarcco/kuryr/internal/domain"
)

const (
	BizConfigCacheKeyPrefix  = "biz_config"
	BizConfigDefaultLocalExp = 15 * time.Minute
)

type BizConfigCache interface {
	Set(ctx context.Context, bizConfig domain.BizConfig) error
}

func BizConfigCacheKey(configId uint64) string {
	return fmt.Sprintf("%s:%d", BizConfigCacheKeyPrefix, configId)
}

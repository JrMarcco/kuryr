package slidewindow

import (
	"context"
	_ "embed"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/JrMarcco/kuryr/internal/pkg/ratelimit"
	"github.com/redis/go-redis/v9"
)

//go:embed lua/slide_window.lua
var slideWindowLua string

var _ ratelimit.Limiter = (*Limiter)(nil)

type Limiter struct {
	rc        redis.Cmdable
	size      time.Duration
	threshold int64

	// 用于生成请求 id 的计数器
	reqCnt atomic.Uint64
}

func (l *Limiter) Allow(ctx context.Context, biz string) (bool, error) {
	requestId := fmt.Sprintf("%d-%d", time.Now().UnixMilli(), l.reqCnt.Add(1))

	res, err := l.rc.Eval(
		ctx,
		slideWindowLua,
		[]string{l.rateLimitKey(biz), l.cleanUpKey(biz)},
		l.size.Milliseconds(),
		l.threshold,
		time.Now().UnixMilli(),
		requestId,
	).Result()

	if err != nil {
		return false, err
	}
	return res == "ok", nil
}

func (l *Limiter) rateLimitKey(biz string) string {
	return fmt.Sprintf("kuryr:rate_limit:%s", biz)
}

func (l *Limiter) cleanUpKey(biz string) string {
	return fmt.Sprintf("kuryr:rate_limit:cleanup:%s", biz)
}

func NewLimiter(rc redis.Cmdable, size time.Duration, threshold int64) (*Limiter, error) {
	if size <= 0 {
		return nil, fmt.Errorf("window size must be positive")
	}
	if threshold <= 0 {
		return nil, fmt.Errorf("threshold must be positive")
	}

	return &Limiter{
		rc:        rc,
		size:      size,
		threshold: threshold,
	}, nil
}

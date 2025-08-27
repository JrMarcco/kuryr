package ratelimit

import (
	"context"
)

// Limiter 限流器。
type Limiter interface {
	// Allow 判断请求是否允许。
	// biz 为业务标识，用于区分不同的业务。
	// 如果请求允许放行则返回 true，否则（被限流）返回 false。
	Allow(ctx context.Context, biz string) (bool, error)
}

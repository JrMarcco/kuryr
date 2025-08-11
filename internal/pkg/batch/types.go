package batch

import (
	"context"
	"time"
)

//go:generate mockgen -source=./types.go -destination=./mock/batch.mock.go -package=batchmock -typed Adjuster

// Adjuster 批任务批大小调节器，根据任务响应时间动态调整批大小。
type Adjuster interface {
	Adjust(ctx context.Context, respTime time.Duration) (int, error)
}

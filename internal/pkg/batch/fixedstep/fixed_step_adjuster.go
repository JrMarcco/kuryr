package fixedstep

import (
	"context"
	"time"

	"github.com/JrMarcco/kuryr/internal/pkg/batch"
)

var _ batch.Adjuster = (*Adjuster)(nil)

// Adjuster 固定步长调节器。
type Adjuster struct {
	minSize    int // 最小批大小
	maxSize    int // 最大批大小
	currSize   int // 当前批大小
	adjustStep int // 调整步长

	lastAdjustTime    time.Time     // 上次调整时间
	minAdjustInterval time.Duration // 最小调整间隔

	fastThreshold time.Duration // 快速响应阈值
	slowThreshold time.Duration // 慢响应阈值
}

func (a *Adjuster) Adjust(_ context.Context, respTime time.Duration) (int, error) {
	if !a.lastAdjustTime.IsZero() && time.Since(a.lastAdjustTime) < a.minAdjustInterval {
		return a.currSize, nil
	}

	// 快响应，增长步长。
	if respTime < a.fastThreshold {
		if a.currSize < a.maxSize {
			a.currSize = min(a.currSize+a.adjustStep, a.maxSize)
			a.lastAdjustTime = time.Now()
		}
		return a.currSize, nil
	}

	// 慢响应，减少步长。
	if respTime > a.slowThreshold {
		if a.currSize > a.minSize {
			a.currSize = max(a.currSize-a.adjustStep, a.minSize)
			a.lastAdjustTime = time.Now()
		}
		return a.currSize, nil
	}

	// 正常响应，保持当前大小。
	return a.currSize, nil
}

func NewAdjuster(
	initSize, minSize, maxSize, adjustStep int,
	minAdjustInterval, fastThreshold, slowThreshold time.Duration,
) *Adjuster {
	if initSize < minSize {
		initSize = minSize
	}

	if initSize > maxSize {
		initSize = maxSize
	}

	return &Adjuster{
		minSize:           minSize,
		maxSize:           maxSize,
		currSize:          initSize,
		adjustStep:        adjustStep,
		minAdjustInterval: minAdjustInterval,
		fastThreshold:     fastThreshold,
		slowThreshold:     slowThreshold,
		lastAdjustTime:    time.Time{},
	}
}

package slidewindow

import (
	"context"
	"sync"
	"time"

	"github.com/JrMarcco/kuryr/internal/pkg/batch"
	"github.com/JrMarcco/kuryr/internal/pkg/ringbuffer"
)

var _ batch.Adjuster = (*Adjuster)(nil)

// Adjuster 滑动窗口调节器。
type Adjuster struct {
	mu sync.RWMutex

	minSize    int
	maxSize    int
	currSize   int
	adjustStep int

	lastAdjustTime    time.Time
	minAdjustInterval time.Duration

	buffer *ringbuffer.TimeDurationBuffer // 响应时间环形缓冲区，用来实现滑动窗口
}

// Adjust implements batch.Adjuster.
func (a *Adjuster) Adjust(_ context.Context, respTime time.Duration) (int, error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.buffer.Add(respTime)

	// 窗口没填满的情况下不进行调整。
	if a.buffer.Count() < a.buffer.Size() {
		return a.currSize, nil
	}

	if !a.lastAdjustTime.IsZero() && time.Since(a.lastAdjustTime) < a.minAdjustInterval {
		return a.currSize, nil
	}

	avg := a.buffer.Avg()

	// 快响应，增加步长
	if respTime < avg {
		if a.currSize < a.maxSize {
			a.currSize = min(a.currSize+a.adjustStep, a.maxSize)
			a.lastAdjustTime = time.Now()
		}
		return a.currSize, nil
	}

	// 慢响应，减少步长
	if respTime > avg {
		if a.currSize > a.minSize {
			a.currSize = max(a.currSize-a.adjustStep, a.minSize)
			a.lastAdjustTime = time.Now()
		}
		return a.currSize, nil
	}

	// 正常响应，保持当前大小
	return a.currSize, nil
}

func NewAdjuster(
	bufferSize int,
	initSize, minSize, maxSize, adjustStep int,
	minAdjustInterval time.Duration,
) (*Adjuster, error) {
	if initSize < minSize {
		initSize = minSize
	}

	if initSize > maxSize {
		initSize = maxSize
	}

	buffer, err := ringbuffer.NewTimeDurationBuffer(bufferSize)
	if err != nil {
		return nil, err
	}

	return &Adjuster{
		minSize:           minSize,
		maxSize:           maxSize,
		currSize:          initSize,
		adjustStep:        adjustStep,
		lastAdjustTime:    time.Time{},
		minAdjustInterval: minAdjustInterval,
		buffer:            buffer,
	}, nil
}

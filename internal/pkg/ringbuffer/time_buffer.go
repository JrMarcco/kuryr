package ringbuffer

import (
	"fmt"
	"sync"
	"time"
)

var (
	ErrInvalidBufferSize = fmt.Errorf("[kuryr] invalid buffer size [ %d ]", 0)
)

// TimeDurationBuffer 时间戳环形缓冲区。
// 线程安全且大小固定。
type TimeDurationBuffer struct {
	mu sync.RWMutex

	buffer []time.Duration

	size     int
	count    int
	writePos int

	sum time.Duration
}

func (b *TimeDurationBuffer) Add(d time.Duration) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.count == b.size {
		b.sum -= b.buffer[b.writePos]
	} else {
		b.count++
	}

	b.buffer[b.writePos] = d
	b.sum += d
	b.writePos = (b.writePos + 1) % b.size
}

func (b *TimeDurationBuffer) Avg() time.Duration {
	b.mu.RLock()
	defer b.mu.RUnlock()

	if b.count == 0 {
		return 0
	}

	return b.sum / time.Duration(b.count)
}

func (b *TimeDurationBuffer) Sum() time.Duration {
	b.mu.RLock()
	defer b.mu.RUnlock()

	return b.sum
}

func (b *TimeDurationBuffer) Count() int {
	b.mu.RLock()
	defer b.mu.RUnlock()

	return b.count
}

func (b *TimeDurationBuffer) Size() int {
	// size 在初始化之后保持不变，所以这里不需要加锁。
	return b.size
}

func (b *TimeDurationBuffer) Reset() {
	b.mu.Lock()
	defer b.mu.Unlock()

	clear(b.buffer)

	b.sum = 0
	b.count = 0
	b.writePos = 0
}

func NewTimeDurationBuffer(size int) (*TimeDurationBuffer, error) {
	if size <= 0 {
		return nil, ErrInvalidBufferSize
	}

	return &TimeDurationBuffer{
		buffer: make([]time.Duration, size),
		size:   size,
	}, nil
}

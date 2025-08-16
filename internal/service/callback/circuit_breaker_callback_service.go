package callback

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/JrMarcco/kuryr/internal/domain"
)

type state string

const (
	stateOpen     state = "open"
	stateClosed   state = "closed"
	stateHalfOpen state = "half-open"
)

var _ Service = (*CircuitBreakerService)(nil)

// CircuitBreakerService 带熔断机制的回调服务。
//
// 针对广播全表的回调请求，使用熔断机制来保护后端服务。
type CircuitBreakerService struct {
	DefaultService

	mu sync.RWMutex

	successCnt int
	failureCnt int

	failureThreshold int
	successThreshold int

	state state

	lastFailureTime time.Time

	coolDownPeriod time.Duration
}

func (s *CircuitBreakerService) Send(ctx context.Context, startTime int64, batchSize int) error {
	s.mu.RLock()
	state := s.state
	s.mu.RUnlock()

	if state == stateOpen {
		// 熔断器处于打开状态。
		s.mu.Lock()
		if time.Since(s.lastFailureTime) <= s.coolDownPeriod {
			// 还在冷却期，直接返回。
			defer s.mu.Unlock()
			return fmt.Errorf("[kuryr] circuit breaker is open, cooldown period not over")
		}

		// 过了冷却期，尝试半开。
		s.state = stateHalfOpen
		s.mu.Unlock()
	}

	err := s.DefaultService.Send(ctx, startTime, batchSize)

	s.mu.Lock()
	defer s.mu.Unlock()

	if err != nil {
		s.failureCnt++
		s.successCnt = 0
		s.lastFailureTime = time.Now()

		if s.failureCnt >= s.failureThreshold {
			s.state = stateOpen
		}
		return err
	}

	s.successCnt++
	if s.state == stateHalfOpen && s.successCnt >= s.successThreshold {
		s.state = stateClosed
		s.failureCnt = 0
	}

	return nil
}

func (s *CircuitBreakerService) SendByNotification(ctx context.Context, n domain.Notification) error {
	return s.DefaultService.SendByNotification(ctx, n)
}

func (s *CircuitBreakerService) SendByNotifications(ctx context.Context, ns []domain.Notification) error {
	return s.DefaultService.SendByNotifications(ctx, ns)
}

package callback

import (
	"context"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

// CallbackOptimizations 回调优化配置
type CallbackOptimizations struct {
	// 并发控制
	MaxConcurrency int
	
	// 限流配置
	RateLimiter *rate.Limiter
	
	// 熔断器配置
	CircuitBreaker *CircuitBreaker
	
	// 批量处理配置
	BatchTimeout time.Duration
	MaxBatchWait time.Duration
}

// CircuitBreaker 简单的熔断器实现
type CircuitBreaker struct {
	mu              sync.RWMutex
	failures        int
	successCount    int
	failureThreshold int
	successThreshold int
	state           string // "closed", "open", "half-open"
	lastFailureTime time.Time
	cooldownPeriod  time.Duration
}

func (cb *CircuitBreaker) Call(fn func() error) error {
	cb.mu.RLock()
	state := cb.state
	cb.mu.RUnlock()
	
	if state == "open" {
		cb.mu.RLock()
		if time.Since(cb.lastFailureTime) > cb.cooldownPeriod {
			cb.mu.RUnlock()
			cb.mu.Lock()
			cb.state = "half-open"
			cb.mu.Unlock()
		} else {
			cb.mu.RUnlock()
			return fmt.Errorf("circuit breaker is open")
		}
	}
	
	err := fn()
	
	cb.mu.Lock()
	defer cb.mu.Unlock()
	
	if err != nil {
		cb.failures++
		cb.successCount = 0
		cb.lastFailureTime = time.Now()
		
		if cb.failures >= cb.failureThreshold {
			cb.state = "open"
		}
		return err
	}
	
	cb.successCount++
	if cb.state == "half-open" && cb.successCount >= cb.successThreshold {
		cb.state = "closed"
		cb.failures = 0
	}
	
	return nil
}

// OptimizedSendCallback 优化后的发送回调方法
func (s *DefaultService) OptimizedSendCallback(ctx context.Context, startTime int64, batchSize int, opts CallbackOptimizations) error {
	dsts := s.shardingStrategy.Broadcast()
	
	// 使用 errgroup 实现并发处理
	g, ctx := errgroup.WithContext(ctx)
	
	// 限制并发数
	sem := make(chan struct{}, opts.MaxConcurrency)
	
	// 错误收集器
	var errorsMu sync.Mutex
	errors := make(map[string]error)
	
	for _, dst := range dsts {
		dst := dst // 捕获循环变量
		g.Go(func() error {
			// 获取信号量
			select {
			case sem <- struct{}{}:
				defer func() { <-sem }()
			case <-ctx.Done():
				return ctx.Err()
			}
			
			// 限流检查
			if opts.RateLimiter != nil {
				if err := opts.RateLimiter.Wait(ctx); err != nil {
					return err
				}
			}
			
			// 熔断器检查
			if opts.CircuitBreaker != nil {
				err := opts.CircuitBreaker.Call(func() error {
					return s.dealDstCallbackLogs(ctx, dst, startTime, batchSize)
				})
				
				if err != nil {
					errorsMu.Lock()
					errors[fmt.Sprintf("%s.%s", dst.DB, dst.Table)] = err
					errorsMu.Unlock()
					
					s.logger.Error("[kuryr] failed to deal with dst callback logs",
						zap.String("db", dst.DB),
						zap.String("table", dst.Table),
						zap.Error(err),
					)
				}
			} else {
				// 无熔断器时的正常处理
				if err := s.dealDstCallbackLogs(ctx, dst, startTime, batchSize); err != nil {
					errorsMu.Lock()
					errors[fmt.Sprintf("%s.%s", dst.DB, dst.Table)] = err
					errorsMu.Unlock()
					
					s.logger.Error("[kuryr] failed to deal with dst callback logs",
						zap.String("db", dst.DB),
						zap.String("table", dst.Table),
						zap.Error(err),
					)
				}
			}
			
			return nil // 不返回错误，继续处理其他分片
		})
	}
	
	// 等待所有goroutine完成
	if err := g.Wait(); err != nil {
		return err
	}
	
	// 如果有错误，返回聚合错误信息
	if len(errors) > 0 {
		s.logger.Warn("[kuryr] some shards failed during callback processing",
			zap.Int("failed_count", len(errors)),
			zap.Int("total_count", len(dsts)),
		)
		// 可以选择返回错误或仅记录
		// return fmt.Errorf("failed to process %d out of %d shards", len(errors), len(dsts))
	}
	
	return nil
}

// 批量聚合发送优化
func (s *DefaultService) batchSendWithAggregation(ctx context.Context, logs []domain.CallbackLog) error {
	// 按照回调URL分组
	grouped := make(map[string][]domain.CallbackLog)
	for _, log := range logs {
		cfg, err := s.getCallbackConfig(ctx, log.Notification.BizId)
		if err != nil {
			continue
		}
		grouped[cfg.CallbackUrl] = append(grouped[cfg.CallbackUrl], log)
	}
	
	// 并发发送各组
	var wg sync.WaitGroup
	for url, group := range grouped {
		wg.Add(1)
		go func(url string, logs []domain.CallbackLog) {
			defer wg.Done()
			// 批量发送逻辑
			s.batchSendToUrl(ctx, url, logs)
		}(url, group)
	}
	
	wg.Wait()
	return nil
}

//go:build e2e

package slidewindow

import (
	"context"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
)

func initRedis() redis.Cmdable {
	redisClient := redis.NewClient(&redis.Options{
		Addr:     "192.168.3.3:6379", // 根据实际测试环境调整
		Password: "<passwd>",
	})

	return redisClient
}

func TestLimiter(t *testing.T) {
	l, err := NewLimiter(initRedis(), time.Second, 100)
	assert.NoError(t, err)

	totalReq := 5000
	var (
		successCnt int
		limitedCnt int
	)

	start := time.Now()
	for range totalReq {
		limited, err := l.Allow(context.Background(), "test_biz")
		if err != nil {
			t.Fatalf("failed to limit: %v", err)
			return
		}

		if limited {
			limitedCnt++
		} else {
			successCnt++
		}
	}

	end := time.Now()
	t.Logf("test start at %v", start.Format(time.StampMilli))
	t.Logf("test end at %v", end.Format(time.StampMilli))

	t.Logf("total request count: %d, success count: %d, limited count: %d", totalReq, successCnt, limitedCnt)
}

func TestLimiter_Allow(t *testing.T) {
	rc := initRedis()

	tcs := []struct {
		name    string
		limiter *Limiter
		biz     string
		wantRes bool
		wantErr error
	}{
		{
			name: "allow",
			limiter: func(t *testing.T) *Limiter {
				limiter, err := NewLimiter(rc, 100*time.Millisecond, 3)
				assert.NoError(t, err)
				rc.Del(context.Background(), limiter.rateLimitKey("biz_1"))

				res, err := limiter.Allow(context.Background(), "biz_1")
				assert.NoError(t, err)
				assert.True(t, res)
				return limiter
			}(t),
			biz:     "biz_1",
			wantRes: true,
			wantErr: nil,
		}, {
			name: "limited",
			limiter: func(t *testing.T) *Limiter {
				limiter, err := NewLimiter(rc, 10*time.Second, 10)
				assert.NoError(t, err)
				rc.Del(context.Background(), limiter.rateLimitKey("biz_2"))

				for range 10 {
					res, err := limiter.Allow(context.Background(), "biz_2")
					assert.NoError(t, err)
					assert.True(t, res)
				}
				return limiter
			}(t),
			biz:     "biz_2",
			wantRes: false,
			wantErr: nil,
		}, {
			name: "window is free",
			limiter: func(t *testing.T) *Limiter {
				limiter, err := NewLimiter(rc, time.Second, 10)
				assert.NoError(t, err)
				rc.Del(context.Background(), limiter.rateLimitKey("biz_3"))
				return limiter
			}(t),
			biz:     "biz_3",
			wantRes: true,
			wantErr: nil,
		}, {
			name: "after time of window size",
			limiter: func(t *testing.T) *Limiter {
				limiter, err := NewLimiter(rc, time.Second, 10)
				assert.NoError(t, err)
				rc.Del(context.Background(), limiter.rateLimitKey("biz_4"))

				for range 10 {
					res, err := limiter.Allow(context.Background(), "biz_4")
					assert.NoError(t, err)
					assert.True(t, res)
				}
				time.Sleep(time.Second)
				return limiter
			}(t),
			biz:     "biz_4",
			wantRes: true,
			wantErr: nil,
		}, {
			name: "another window",
			limiter: func(t *testing.T) *Limiter {
				limiter, err := NewLimiter(rc, time.Second, 10)
				assert.NoError(t, err)
				rc.Del(context.Background(), limiter.rateLimitKey("another_biz_5"))

				for range 10 {
					res, err := limiter.Allow(context.Background(), "biz_5")
					assert.NoError(t, err)
					assert.True(t, res)
				}
				return limiter
			}(t),
			biz:     "another_biz_5",
			wantRes: true,
			wantErr: nil,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			limiter := tc.limiter
			allow, err := limiter.Allow(context.Background(), tc.biz)
			assert.Equal(t, tc.wantErr, err)

			if err != nil {
				return
			}

			assert.Equal(t, tc.wantRes, allow)
		})
	}
}

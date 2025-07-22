package retry

import (
	"fmt"
	"time"

	"github.com/JrMarcco/easy-kit/retry"
)

type Config struct {
	Type               string                    `json:"type"`
	FixedInterval      *FixedIntervalConfig      `json:"fixed_interval"`
	ExponentialBackoff *ExponentialBackoffConfig `json:"exponential_backoff"`
}

type FixedIntervalConfig struct {
	Interval      time.Duration `json:"interval"`
	MaxRetryTimes int32         `json:"max_retry_times"`
}

type ExponentialBackoffConfig struct {
	InitialInterval time.Duration `json:"initial_interval"`
	MaxInterval     time.Duration `json:"max_interval"`
	MaxRetryTimes   int32         `json:"max_retry_times"`
}

func NewRetryStrategy(cfg Config) (retry.Strategy, error) {
	switch cfg.Type {
	case "fixed_interval":
		return retry.NewFixedIntervalStrategy(cfg.FixedInterval.Interval, cfg.FixedInterval.MaxRetryTimes)
	case "exponential_backoff":
		return retry.NewExponentialBackoffStrategy(
			cfg.ExponentialBackoff.InitialInterval,
			cfg.ExponentialBackoff.MaxInterval,
			cfg.ExponentialBackoff.MaxRetryTimes,
		)
	default:
		return nil, fmt.Errorf("[kuryr] unsupported retry strategy: %s", cfg.Type)
	}
}

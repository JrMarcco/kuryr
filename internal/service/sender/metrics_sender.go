package sender

import (
	"context"
	"time"

	"github.com/JrMarcco/kuryr/internal/domain"
	"github.com/JrMarcco/kuryr/internal/service/ports"
	"github.com/prometheus/client_golang/prometheus"
)

// Prometheus 指标常量
const (
	metricsMaxAge = 5 * time.Minute

	metricsP50Percentile = 0.5
	metricsP50Error      = 0.05

	metricsP90Percentile = 0.9
	metricsP90Error      = 0.01

	metricsP95Percentile = 0.95
	metricsP95Error      = 0.005

	metricsP99Pencentile = 0.99
	metricsP99Error      = 0.001
)

var _ ports.NotificationSender = (*MetricsSender)(nil)

// MetricsSender 带指标的发送器。
// 使用装饰器模式为 ports.NotificationSender 添加 Prometheus 指标收集。
type MetricsSender struct {
	sender ports.NotificationSender

	durationSummary *prometheus.SummaryVec
	sendCounter     *prometheus.CounterVec // 发送次数
	statusCounter   *prometheus.CounterVec // 发送状态计数器
}

func (s *MetricsSender) Send(ctx context.Context, n domain.Notification) (domain.SendResp, error) {
	start := time.Now()

	// 增加计数
	s.sendCounter.WithLabelValues(string(n.Channel)).Inc()

	resp, err := s.sender.Send(ctx, n)

	// 记录耗时
	duration := time.Since(start).Seconds()
	s.durationSummary.WithLabelValues(
		string(n.Channel),
		string(n.SendStatus),
	).Observe(duration)

	// 记录状态
	s.statusCounter.WithLabelValues(
		string(n.Channel),
		string(resp.Result.SendStatus),
	).Inc()

	return resp, err
}

func (s *MetricsSender) BatchSend(ctx context.Context, ns []domain.Notification) (domain.BatchSendResp, error) {
	if len(ns) == 0 {
		return domain.BatchSendResp{}, nil
	}

	// TODO: implement me
	panic("implement me")
}

func NewMetricsSender(sender ports.NotificationSender) *MetricsSender {
	durationSummary := prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Name: "kuryr_notification_send_duration_seconds",
			Help: "The duration of sending a notification",
			Objectives: map[float64]float64{
				metricsP50Percentile: metricsP50Error,
				metricsP90Percentile: metricsP90Error,
				metricsP95Percentile: metricsP95Error,
				metricsP99Pencentile: metricsP99Error,
			},
			MaxAge: metricsMaxAge,
		},
		[]string{"channel", "status"},
	)

	sendCounter := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "kuryr_notification_send_count",
			Help: "The count of sending a notification",
		},
		[]string{"channel"},
	)

	statusCounter := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "kuryr_notification_send_status_count",
			Help: "The count of sending a notification by status",
		},
		[]string{"channel", "status"},
	)

	// 注册指标
	prometheus.MustRegister(durationSummary, sendCounter, statusCounter)

	return &MetricsSender{
		sender:          sender,
		durationSummary: durationSummary,
		sendCounter:     sendCounter,
		statusCounter:   statusCounter,
	}
}

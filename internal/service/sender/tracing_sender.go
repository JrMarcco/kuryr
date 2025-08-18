package sender

import (
	"context"
	"strconv"
	"strings"

	"github.com/JrMarcco/easy-kit/slice"
	"github.com/JrMarcco/kuryr/internal/domain"
	"github.com/JrMarcco/kuryr/internal/service/ports"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

var _ ports.NotificationSender = (*TracingSender)(nil)

// TracingSender 带链路追踪的发送器。
// 使用装饰器模式为 ports.NotificationSender 添加链路追踪。
type TracingSender struct {
	sender ports.NotificationSender
	tracer trace.Tracer
}

func (s *TracingSender) Send(ctx context.Context, n domain.Notification) (domain.SendResp, error) {
	ctx, span := s.tracer.Start(ctx, "NotificationSender.Send", trace.WithAttributes(
		attribute.String("notification.id", n.Id),
		attribute.String("notification.biz_id", strconv.FormatUint(n.BizId, 10)),
		attribute.String("notification.biz_key", n.BizKey),
		attribute.String("notification.channel", string(n.Channel)),
	))
	defer span.End()

	resp, err := s.sender.Send(ctx, n)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return resp, err
	}

	span.SetAttributes(
		attribute.String("notification.id", resp.Result.NotificationId),
		attribute.String("notification.send_status", string(resp.Result.SendStatus)),
	)
	return resp, err
}

func (s *TracingSender) BatchSend(ctx context.Context, ns []domain.Notification) (domain.BatchSendResp, error) {
	ctx, span := s.tracer.Start(ctx, "NotificationSender.BatchSend", trace.WithAttributes(
		attribute.Int("notification.count", len(ns)),
	))
	defer span.End()

	if len(ns) > 0 {
		span.SetAttributes(
			attribute.String("notification.biz_id", strconv.FormatUint(ns[0].BizId, 10)),
			attribute.String("notification.biz_key", strings.Join(slice.Map(ns, func(idx int, src domain.Notification) string {
				return src.BizKey
			}), ",")),
		)
	}

	resp, err := s.sender.BatchSend(ctx, ns)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return resp, err
	}

	var success, failure int
	for _, res := range resp.Results {
		if res.SendStatus == domain.SendStatusSuccess {
			success++
		} else {
			failure++
		}
	}

	span.SetAttributes(
		attribute.Int("notification.success", success),
		attribute.Int("notification.failure", failure),
	)
	return resp, err
}

func NewTracingSender(sender ports.NotificationSender) *TracingSender {
	return &TracingSender{
		sender: sender,
		tracer: otel.Tracer("kuryr.sender"),
	}
}

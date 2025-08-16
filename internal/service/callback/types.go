package callback

import (
	"context"

	"github.com/JrMarcco/kuryr/internal/domain"
)

type Service interface {
	Send(ctx context.Context, startTime int64, batchSize int) error
	SendByNotification(ctx context.Context, n domain.Notification) error
	SendByNotifications(ctx context.Context, ns []domain.Notification) error
}

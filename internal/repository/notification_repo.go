package repository

import (
	"context"

	"github.com/JrMarcco/kuryr/internal/domain"
)

type NotificationRepo interface {
	MarkSuccess(ctx context.Context, n domain.Notification) error
	MarkFailure(ctx context.Context, n domain.Notification) error
}

var _ NotificationRepo = (*DefaultNotificationRepo)(nil)

type DefaultNotificationRepo struct{}

func (r *DefaultNotificationRepo) MarkSuccess(ctx context.Context, n domain.Notification) error {
	// TODO: implement me
	panic("implement me")
}

func (r *DefaultNotificationRepo) MarkFailure(ctx context.Context, n domain.Notification) error {
	// TODO: implement me
	panic("implement me")
}

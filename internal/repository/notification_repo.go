package repository

import (
	"context"

	"github.com/JrMarcco/kuryr/internal/domain"
	"github.com/JrMarcco/kuryr/internal/repository/dao"
)

type NotificationRepo interface {
	Save(ctx context.Context, n domain.Notification) (domain.Notification, error)
	SaveWithCallback(ctx context.Context, n domain.Notification) (domain.Notification, error)

	BatchSave(ctx context.Context, ns []domain.Notification) ([]domain.Notification, error)
	BatchSaveWithCallback(ctx context.Context, ns []domain.Notification) ([]domain.Notification, error)

	MarkSuccess(ctx context.Context, n domain.Notification) error
	MarkFailure(ctx context.Context, n domain.Notification) error
}

var _ NotificationRepo = (*DefaultNotificationRepo)(nil)

type DefaultNotificationRepo struct {
	callbackLogDao  dao.CallbackLogDao
	notificationDao dao.NotificationDao
}

func (r *DefaultNotificationRepo) Save(ctx context.Context, n domain.Notification) (domain.Notification, error) {
	// TODO: implement me
	panic("implement me")
}

func (r *DefaultNotificationRepo) SaveWithCallback(ctx context.Context, n domain.Notification) (domain.Notification, error) {
	// TODO: implement me
	panic("implement me")
}

func (r *DefaultNotificationRepo) BatchSave(ctx context.Context, ns []domain.Notification) ([]domain.Notification, error) {
	// TODO: implement me
	panic("implement me")
}

func (r *DefaultNotificationRepo) BatchSaveWithCallback(ctx context.Context, ns []domain.Notification) ([]domain.Notification, error) {
	// TODO: implement me
	panic("implement me")
}

func (r *DefaultNotificationRepo) MarkSuccess(ctx context.Context, n domain.Notification) error {
	// TODO: implement me
	panic("implement me")
}

func (r *DefaultNotificationRepo) MarkFailure(ctx context.Context, n domain.Notification) error {
	// TODO: implement me
	panic("implement me")
}

func NewDefaultNotificationRepo(callbackLogDao dao.CallbackLogDao, notificationDao dao.NotificationDao) NotificationRepo {
	return &DefaultNotificationRepo{
		callbackLogDao:  callbackLogDao,
		notificationDao: notificationDao,
	}
}

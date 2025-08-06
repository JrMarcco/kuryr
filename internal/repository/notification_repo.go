package repository

type NotificationRepo interface{}

var _ NotificationRepo = (*DefaultNotificationRepo)(nil)

type DefaultNotificationRepo struct{}

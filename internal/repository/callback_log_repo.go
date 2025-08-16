package repository

import (
	"context"

	"github.com/JrMarcco/easy-kit/slice"
	"github.com/JrMarcco/kuryr/internal/domain"
	"github.com/JrMarcco/kuryr/internal/pkg/sharding"
	"github.com/JrMarcco/kuryr/internal/repository/dao"
)

type CallbackLogRepo interface {
	Save(ctx context.Context, log domain.CallbackLog) error

	BatchUpdate(ctx context.Context, dst sharding.Dst, logs []domain.CallbackLog) error

	FindByNotificationIds(ctx context.Context, notificationIds []uint64) ([]domain.CallbackLog, error)
	BatchFindByTime(ctx context.Context, dst sharding.Dst, startTime int64, startId uint64, batchSize int) ([]domain.CallbackLog, uint64, error)
}

var _ CallbackLogRepo = (*DefaultCallbackLogRepo)(nil)

type DefaultCallbackLogRepo struct {
	dao dao.CallbackLogDao
}

func (r *DefaultCallbackLogRepo) Save(ctx context.Context, log domain.CallbackLog) error {
	return r.dao.Save(ctx, r.toEntity(log))
}

func (r *DefaultCallbackLogRepo) BatchUpdate(ctx context.Context, dst sharding.Dst, logs []domain.CallbackLog) error {
	entities := slice.Map(logs, func(_ int, log domain.CallbackLog) dao.CallbackLog {
		return r.toEntity(log)
	})
	return r.dao.BatchUpdate(ctx, dst, entities)
}

func (r *DefaultCallbackLogRepo) FindByNotificationIds(ctx context.Context, notificationIds []uint64) ([]domain.CallbackLog, error) {
	// TODO: implement me
	panic("implement me")
}

func (r *DefaultCallbackLogRepo) BatchFindByTime(ctx context.Context, dst sharding.Dst, startTime int64, startId uint64, batchSize int) ([]domain.CallbackLog, uint64, error) {
	entities, nextStartId, err := r.dao.BatchFindByTime(ctx, dst, startTime, startId, batchSize)
	if err != nil {
		return nil, nextStartId, err
	}

	logs := slice.Map(entities, func(_ int, entity dao.CallbackLog) domain.CallbackLog {
		return r.toDomain(entity)
	})
	return logs, nextStartId, nil
}

func (r *DefaultCallbackLogRepo) toEntity(log domain.CallbackLog) dao.CallbackLog {
	return dao.CallbackLog{
		Id:                 log.Id,
		BizId:              log.BizId,
		BizKey:             log.BizKey,
		NotificationId:     log.Notification.Id,
		NotificationStatus: string(log.Notification.SendStatus),
		RetriedTimes:       log.RetriedTimes,
		NextRetryAt:        log.NextRetryAt,
		CallbackStatus:     string(log.Status),
	}
}

func (r *DefaultCallbackLogRepo) toDomain(entity dao.CallbackLog) domain.CallbackLog {
	return domain.CallbackLog{
		Id:     entity.Id,
		BizId:  entity.BizId,
		BizKey: entity.BizKey,
		Notification: domain.Notification{
			Id:         entity.NotificationId,
			SendStatus: domain.SendStatus(entity.NotificationStatus),
		},
		RetriedTimes: entity.RetriedTimes,
		NextRetryAt:  entity.NextRetryAt,
		Status:       domain.CallbackLogStatus(entity.CallbackStatus),
	}
}

func NewCallbackLogRepo(dao dao.CallbackLogDao) CallbackLogRepo {
	return &DefaultCallbackLogRepo{
		dao: dao,
	}
}

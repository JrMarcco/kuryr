package dao

import (
	"context"

	"github.com/JrMarcco/easy-kit/xsync"
	"github.com/JrMarcco/kuryr/internal/pkg/idgen"
	"github.com/JrMarcco/kuryr/internal/pkg/sharding"
	"gorm.io/gorm"
)

// CallbackLog 回调日志数据对象。
type CallbackLog struct {
	Id uint64 `gorm:"column:id"`

	NotificationId uint64 `gorm:"column:notification_id"`
	RetriedTimes   int32  `gorm:"column:retried_times"`
	NextRetryAt    int64  `gorm:"column:next_retry_at"`
	CallbackStatus string `gorm:"column:callback_status"`

	CreatedAt int64 `gorm:"column:created_at"`
	UpdatedAt int64 `gorm:"column:updated_at"`
}

func (CallbackLog) TableName() string {
	return "callback_log"
}

// CallbackLogDao 回调日志数据访问对象。
//
// 注意：
//
//	这里使用分库分表设计。
type CallbackLogDao interface {
	BatchUpdate(ctx context.Context, logs []CallbackLog) error

	FindByNotificationIds(ctx context.Context, notificationIds []uint64) ([]CallbackLog, error)
	BatchFindByTime(ctx context.Context, startTime int64, startId uint, batchSize int) ([]CallbackLog, uint64, error)
}

var _ CallbackLogDao = (*DefaultCallbackLogDao)(nil)

type DefaultCallbackLogDao struct {
	dbs *xsync.Map[string, *gorm.DB]

	idGenerator      idgen.Generator
	shardingStrategy sharding.Strategy
}

func (d *DefaultCallbackLogDao) BatchUpdate(ctx context.Context, logs []CallbackLog) error {
	//TODO: implement me
	panic("implement me")
}
func (d *DefaultCallbackLogDao) FindByNotificationIds(ctx context.Context, notificationIds []uint64) ([]CallbackLog, error) {
	//TODO: implement me
	panic("implement me")
}

func (d *DefaultCallbackLogDao) BatchFindByTime(ctx context.Context, startTime int64, startId uint, batchSize int) ([]CallbackLog, uint64, error) {
	//TODO: implement me
	panic("implement me")
}

func NewCallbackLogDao(dbs *xsync.Map[string, *gorm.DB], shardingStrategy sharding.Strategy, idGenerator idgen.Generator) CallbackLogDao {
	return &DefaultCallbackLogDao{
		dbs:              dbs,
		shardingStrategy: shardingStrategy,
		idGenerator:      idGenerator,
	}
}

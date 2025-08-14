package dao

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/JrMarcco/easy-kit/xsync"
	"github.com/JrMarcco/kuryr/internal/domain"
	"github.com/JrMarcco/kuryr/internal/pkg/idgen"
	"github.com/JrMarcco/kuryr/internal/pkg/sharding"
	"golang.org/x/sync/errgroup"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
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
	BatchUpdate(ctx context.Context, dst sharding.Dst, logs []CallbackLog) error

	FindByNotificationIds(ctx context.Context, notificationIds []uint64) ([]CallbackLog, error)
	BatchFindByTime(ctx context.Context, dst sharding.Dst, startTime int64, startId uint64, batchSize int) ([]CallbackLog, uint64, error)
}

var _ CallbackLogDao = (*DefaultCallbackLogDao)(nil)

type DefaultCallbackLogDao struct {
	dbs *xsync.Map[string, *gorm.DB]

	idGenerator      idgen.Generator
	shardingStrategy sharding.Strategy
}

func (d *DefaultCallbackLogDao) BatchUpdate(ctx context.Context, dst sharding.Dst, logs []CallbackLog) error {
	if len(logs) == 0 {
		return nil
	}

	now := time.Now().UnixMilli()

	db, ok := d.dbs.Load(dst.DB)
	if !ok {
		return fmt.Errorf("[kuryr] failed to load db [ %s ]", dst.DB)
	}

	return db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 分批处理，避免单次 SQL 过大或超过数据库参数限制。
		const batchSize = 1000

		for i := 0; i < len(logs); i += batchSize {
			end := min(i+batchSize, len(logs))
			batch := logs[i:end]

			// 准备批量更新的数据。
			updates := make([]CallbackLog, len(batch))
			for j, log := range batch {
				updates[j] = CallbackLog{
					Id:             log.Id,
					RetriedTimes:   log.RetriedTimes,
					NextRetryAt:    log.NextRetryAt,
					CallbackStatus: log.CallbackStatus,
					UpdatedAt:      now,
				}
			}

			// 需要指定更新的字段，避免误更新其他字段。
			err := tx.Table(dst.Table).
				Clauses(clause.OnConflict{
					// 只更新不插入。
					UpdateAll: false,
				}).
				Select("retried_times", "next_retry_at", "callback_status", "updated_at").
				Save(&updates).Error
			if err != nil {
				return err
			}
		}
		return nil
	})
}
func (d *DefaultCallbackLogDao) FindByNotificationIds(ctx context.Context, notificationIds []uint64) ([]CallbackLog, error) {
	tbs := make(map[string][]uint64)
	for _, id := range notificationIds {
		dst := d.shardingStrategy.DstFromId(id)
		tbs[dst.FullTable()] = append(tbs[dst.FullTable()], id)
	}

	var mu sync.Mutex
	res := make([]CallbackLog, 0, len(notificationIds))

	eg, ctx := errgroup.WithContext(ctx)
	sem := make(chan struct{}, 4)

	for fullTable, ids := range tbs {
		seg := strings.Split(fullTable, ".")

		eg.Go(func() error {
			select {
			case sem <- struct{}{}:
				defer func() { <-sem }()
			case <-ctx.Done():
				return ctx.Err()
			}

			if len(seg) != 2 || seg[1] == "" {
				return fmt.Errorf("[kuryr] invalid full table [ %s ]", fullTable)
			}

			db, ok := d.dbs.Load(seg[0])
			if !ok {
				return fmt.Errorf("[kuryr] failed to load db [ %s ]", seg[0])
			}

			var logs []CallbackLog
			err := db.WithContext(ctx).Table(seg[1]).
				Where("notification_id IN ?", ids).
				Find(&logs).Error
			if err != nil {
				return err
			}

			if len(logs) != 0 {
				mu.Lock()
				defer mu.Unlock()

				res = append(res, logs...)
			}
			return nil
		})

	}

	if err := eg.Wait(); err != nil {
		return nil, err
	}
	return res, nil
}

func (d *DefaultCallbackLogDao) BatchFindByTime(ctx context.Context, dst sharding.Dst, startTime int64, startId uint64, batchSize int) ([]CallbackLog, uint64, error) {
	nextStartId := uint64(0)

	db, ok := d.dbs.Load(dst.DB)
	if !ok {
		return nil, nextStartId, fmt.Errorf("[kuryr] failed to load db [ %s ]", dst.DB)
	}

	var logs []CallbackLog

	err := db.WithContext(ctx).Table(dst.Table).
		Where("next_retry_at <= ?", startTime).
		Where("callback_status IN ?", []string{string(domain.CallbackLogStatusPending), string(domain.CallbackLogStatusPrepare)}).
		Where("id > ?", startId).
		Order("id ASC").
		Limit(batchSize).
		Find(&logs).Error

	if err != nil {
		return logs, nextStartId, err
	}

	if len(logs) > 0 {
		nextStartId = logs[len(logs)-1].Id
	}

	return logs, nextStartId, nil
}

func NewCallbackLogDao(dbs *xsync.Map[string, *gorm.DB], shardingStrategy sharding.Strategy, idGenerator idgen.Generator) CallbackLogDao {
	return &DefaultCallbackLogDao{
		dbs:              dbs,
		shardingStrategy: shardingStrategy,
		idGenerator:      idGenerator,
	}
}

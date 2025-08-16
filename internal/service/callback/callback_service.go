package callback

import (
	"context"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/JrMarcco/easy-grpc/client"
	"github.com/JrMarcco/easy-kit/xsync"
	clientv1 "github.com/JrMarcco/kuryr-api/api/client/v1"
	commonv1 "github.com/JrMarcco/kuryr-api/api/common/v1"
	notificationv1 "github.com/JrMarcco/kuryr-api/api/notification/v1"
	"github.com/JrMarcco/kuryr/internal/domain"
	"github.com/JrMarcco/kuryr/internal/errs"
	"github.com/JrMarcco/kuryr/internal/pkg/retry"
	"github.com/JrMarcco/kuryr/internal/pkg/sharding"
	"github.com/JrMarcco/kuryr/internal/repository"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

var _ Service = (*DefaultService)(nil)

type DefaultService struct {
	callbackConfigMap xsync.Map[uint64, *domain.CallbackConfig] // biz id -> callback config

	grpcClinets *client.Manager[clientv1.CallbackServiceClient]

	shardingStrategy sharding.Strategy

	logRepo       repository.CallbackLogRepo
	bizConfigRepo repository.BizConfigRepo

	logger *zap.Logger
}

// SendCallback 以回调日志为依据发送回调请求。
func (s *DefaultService) Send(ctx context.Context, startTime int64, batchSize int) error {
	dsts := s.shardingStrategy.Broadcast()

	// 按数据库分组，避免连接集中在同一个库
	dbGroup := make(map[string][]sharding.Dst)
	for _, dst := range dsts {
		dbGroup[dst.DB] = append(dbGroup[dst.DB], dst)
	}

	var errMu sync.Mutex
	errMap := make(map[string]error)

	// 使用 errgroup 实现并发处理
	eg, ctx := errgroup.WithContext(ctx)

	for _, tbs := range dbGroup {
		dbTables := tbs

		eg.Go(func() error {
			// 限制并发数，避免数据库连接过多。
			// 注意：
			//  这里限制的是每个库的最大并发数。
			// TODO: 这里最大并发数可以改成配置。
			sem := make(chan struct{}, 4)

			// 每个库开一个 errgroup 并发查询。
			dbEg, dbCtx := errgroup.WithContext(ctx)

			for _, dst := range dbTables {
				d := dst

				dbEg.Go(func() error {
					select {
					case sem <- struct{}{}:
						defer func() { <-sem }()
					case <-dbCtx.Done():
						return dbCtx.Err()
					}

					if err := s.dealDstCallbackLogs(dbCtx, d, startTime, batchSize); err != nil {
						// 记录错误但不中断其他分片的处理
						s.logger.Error("[kuryr] failed to deal with dst callback logs",
							zap.String("table", d.FullTable()),
							zap.Error(err),
						)
						errMu.Lock()
						errMap[d.FullTable()] = err
						errMu.Unlock()
					}
					return nil
				})
			}
			return dbEg.Wait()
		})
	}

	_ = eg.Wait()

	if len(errMap) > 0 {
		s.logger.Warn("[kuryr] some shards failed during callback processing",
			zap.Int("total_count", len(dsts)),
			zap.Int("failed_count", len(errMap)),
		)

		return fmt.Errorf("[kuryr] some shards failed during callback processing")
	}

	return nil
}

func (s *DefaultService) SendByNotification(ctx context.Context, n domain.Notification) error {
	// TODO: implement me
	panic("implement me")
}

func (s *DefaultService) SendByNotifications(ctx context.Context, ns []domain.Notification) error {
	// TODO: implement me
	panic("implement me")
}

func (s *DefaultService) dealDstCallbackLogs(ctx context.Context, dst sharding.Dst, startTime int64, batchSize int) error {
	nextStartId := uint64(0)

	for {
		logs, nextStartId, err := s.logRepo.BatchFindByTime(ctx, dst, startTime, nextStartId, batchSize)
		if err != nil {
			s.logger.Error("[kuryr] failed to find callback logs",
				zap.Int64("start_time", startTime),
				zap.Uint64("start_id", nextStartId),
				zap.Int("batch_size", batchSize),
				zap.Error(err),
			)
			return err
		}

		if len(logs) == 0 {
			break
		}

		if err := s.batchSendAndUpdateStatus(ctx, dst, logs); err != nil {
			return err
		}
	}
	return nil
}

// batchSendAndUpdateStatus 批量发送回调请求并更新回调日志状态。
// 注意：
//
//	这里保证 logs 都是同一个分表里的记录。
func (s *DefaultService) batchSendAndUpdateStatus(ctx context.Context, dst sharding.Dst, logs []domain.CallbackLog) error {
	needUpdates := make([]domain.CallbackLog, 0, len(logs))

	for i := range logs {
		changed, err := s.sendAndSetChangedFields(ctx, &logs[i])
		if err != nil {
			s.logger.Warn("[kuryr] failed to send callback",
				zap.Uint64("notification_id", logs[i].Notification.Id),
				zap.Error(err),
			)
			continue
		}

		if changed {
			needUpdates = append(needUpdates, logs[i])
		}
	}

	return s.logRepo.BatchUpdate(ctx, dst, needUpdates)
}

// sendAndSetChangedFields 发送请求并设置需要更新的字段。
func (s *DefaultService) sendAndSetChangedFields(ctx context.Context, log *domain.CallbackLog) (bool, error) {
	resp, err := s.send(ctx, log.Notification)
	if err != nil {
		return false, err
	}

	if resp.Success {
		log.Status = domain.CallbackLogStatusSuccess
		return true, nil
	}

	// 发送失败意味着回调配置一定存在。
	cfg, _ := s.getCallbackConfig(ctx, log.Notification.BizId)
	retryStrategy, err := retry.NewRetryStrategy(*cfg.RetryPolicyConfig)
	if err != nil {
		// 获取重试策略异常只可能是配置错误，直接失败。
		log.Status = domain.CallbackLogStatusFailure
		return true, nil
	}

	// 成功获取重试策略，计算下一次请求时间
	interval, ok := retryStrategy.Next()
	if ok {
		log.NextRetryAt = time.Now().Add(interval).UnixMilli()
		log.RetriedTimes++
	} else {
		log.Status = domain.CallbackLogStatusFailure
	}
	return true, nil
}

// send 调用 grpc 接口发送回调通知。
func (s *DefaultService) send(ctx context.Context, notification domain.Notification) (*clientv1.SendResultNotifyResponse, error) {
	cfg, err := s.getCallbackConfig(ctx, notification.BizId)
	if err != nil {
		s.logger.Warn("[kuryr] failed to get callback config",
			zap.Uint64("biz_id", notification.BizId),
			zap.Error(err),
		)
		return nil, err
	}

	if cfg == nil {
		return nil, fmt.Errorf("%w: no callback config provided by the business side", errs.ErrRecordNotFound)
	}

	grpcClient, err := s.grpcClinets.Get(cfg.ServiceName)
	if err != nil {
		return nil, fmt.Errorf("[kuryr] failed to get grpc client: %w", err)
	}

	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	return grpcClient.SendResultNotify(ctx, s.buildGrpcRequest(notification))
}

// getCallbackConfig 获取回调配置。
// 包含 grpc 服务名，重试策略等。
func (s *DefaultService) getCallbackConfig(ctx context.Context, bizId uint64) (*domain.CallbackConfig, error) {
	if cfg, ok := s.callbackConfigMap.Load(bizId); ok {
		return cfg, nil
	}

	bizConfig, err := s.bizConfigRepo.FindById(ctx, bizId)
	if err != nil {
		return nil, err
	}

	if bizConfig.CallbackConfig != nil {
		s.callbackConfigMap.Store(bizId, bizConfig.CallbackConfig)
		return bizConfig.CallbackConfig, nil
	}

	return nil, nil
}

// buildGrpcRequest 构建 grpc 请求。
func (s *DefaultService) buildGrpcRequest(notification domain.Notification) *clientv1.SendResultNotifyRequest {
	tplPrams := make(map[string]string)

	if notification.Template.Params != nil {
		tplPrams = notification.Template.Params
	}

	return &clientv1.SendResultNotifyRequest{
		NotificationId: notification.Id,
		RawRequest: &notificationv1.SendRequest{
			Notification: &notificationv1.Notification{
				BizKey:    notification.BizKey,
				Receivers: notification.Receivers,
				Channel:   s.transferChannel(notification.Channel),
				TplId:     strconv.FormatUint(notification.Template.Id, 10),
				TplParams: tplPrams,
			},
		},
		Result: &notificationv1.SendResult{
			NotificationId: notification.Id,
			Status:         s.transferSendStatus(notification.SendStatus),
		},
	}
}

func (s *DefaultService) transferChannel(channel domain.Channel) commonv1.Channel {
	switch channel {
	case domain.ChannelSms:
		return commonv1.Channel_SMS
	case domain.ChannelEmail:
		return commonv1.Channel_EMAIL
	default:
		return commonv1.Channel_CHANNEL_UNSPECIFIED
	}
}

func (s *DefaultService) transferSendStatus(sendStatus domain.SendStatus) notificationv1.SendStatus {
	switch sendStatus {
	case domain.SendStatusPrepare:
		return notificationv1.SendStatus_PREPARE
	case domain.SendStatusPending:
		return notificationv1.SendStatus_PENDING
	case domain.SendStatusSuccess:
		return notificationv1.SendStatus_SUCCESS
	case domain.SendStatusFailure:
		return notificationv1.SendStatus_FAILURE
	case domain.SendStatusCancel:
		return notificationv1.SendStatus_CANCEL
	default:
		// 这里包含 domain.SendStatusSending
		return notificationv1.SendStatus_STATUS_UNSPECIFIED
	}
}

func NewDefaultService(
	grpcClinets *client.Manager[clientv1.CallbackServiceClient],
	shardingStrategy sharding.Strategy,
	logRepo repository.CallbackLogRepo,
	bizConfigRepo repository.BizConfigRepo,
	logger *zap.Logger,
) *DefaultService {
	return &DefaultService{
		callbackConfigMap: xsync.Map[uint64, *domain.CallbackConfig]{},
		grpcClinets:       grpcClinets,
		shardingStrategy:  shardingStrategy,
		logRepo:           logRepo,
		bizConfigRepo:     bizConfigRepo,
		logger:            logger,
	}
}

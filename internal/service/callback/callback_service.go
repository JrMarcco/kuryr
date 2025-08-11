package callback

import (
	"context"

	"github.com/JrMarcco/easy-kit/xsync"
	"github.com/JrMarcco/kuryr/internal/domain"
	"github.com/JrMarcco/kuryr/internal/repository"
	"go.uber.org/zap"
)

type Service interface {
	SendCallback(ctx context.Context, startTime int64, batchSize int) error
}

var _ Service = (*DefaultService)(nil)

type DefaultService struct {
	callbackConfigMap xsync.Map[uint64, *domain.CallbackConfig] // biz id -> callback config

	// TODO: 范型 client manager
	// grpcClinets *client.Manager[callbackv1.CallbackServiceClient]

	logRepo       repository.CallbackLogRepo
	bizConfigRepo repository.BizConfigRepo

	logger *zap.Logger
}

func (s *DefaultService) SendCallback(ctx context.Context, startTime int64, batchSize int) error {
	// TODO: implement me
	panic("implement me")
}

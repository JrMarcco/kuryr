package api

import (
	"context"
	"time"

	configv1 "github.com/JrMarcco/kuryr-api/api/config/v1"
	"github.com/JrMarcco/kuryr/internal/domain"
	"github.com/JrMarcco/kuryr/internal/pkg/retry"
	"github.com/JrMarcco/kuryr/internal/service/bizconf"
)

type BizConfigServer struct {
	configv1.UnimplementedBizConfigServiceServer
	svc bizconf.BizConfigService
}

func (s *BizConfigServer) Save(ctx context.Context, request *configv1.SaveRequest) (*configv1.SaveResponse, error) {
	if request == nil || request.Config == nil {
		return &configv1.SaveResponse{
			Success: false,
			ErrMsg:  "[kuryr] biz config request is invalid, config is nil",
		}, nil
	}

	bizConfig := s.protoToDomain(request.Config)

	err := s.svc.Save(ctx, bizConfig)
	if err != nil {
		return &configv1.SaveResponse{
			Success: false,
			ErrMsg:  err.Error(),
		}, err
	}

	return &configv1.SaveResponse{
		Success: true,
	}, nil
}

func (s *BizConfigServer) Delete(ctx context.Context, request *configv1.DeleteRequest) (*configv1.DeleteResponse, error) {
	panic("implement me")
}

func (s *BizConfigServer) GetById(context.Context, *configv1.GetByIdRequest) (*configv1.GetByIdResponse, error) {
	panic("implement me")
}

func (s *BizConfigServer) GetByIds(context.Context, *configv1.GetByIdsRequest) (*configv1.GetByIdsResponse, error) {
	panic("implement me")
}

// protoToDomain convert protobuf to domain
func (s *BizConfigServer) protoToDomain(pb *configv1.BizConfig) domain.BizConfig {
	bizConfig := domain.BizConfig{
		Id:        pb.BizId,
		RateLimit: int(pb.RateLimit),
	}

	if pb.ChannelConfig != nil {
		channelConfig := &domain.ChannelConfig{
			Channels: make([]domain.ChannelItem, len(pb.ChannelConfig.Items)),
		}

		for index, item := range pb.ChannelConfig.Items {
			channelConfig.Channels[index] = domain.ChannelItem{
				Channel:  item.Channel,
				Priority: int(item.Priority),
				Enabled:  item.Enabled,
			}
		}

		if pb.ChannelConfig.RetryPolicy != nil {
			retryPolicyConfig := s.convertRetry(pb.ChannelConfig.RetryPolicy)
			channelConfig.RetryPolicyConfig = retryPolicyConfig
		}
		bizConfig.ChannelConfig = channelConfig
	}

	if pb.QuotaConfig != nil {
		quotaConfig := &domain.QuotaConfig{}
		if pb.QuotaConfig.Daily != nil {
			dailyQuota := pb.QuotaConfig.Daily
			quotaConfig.Daily = &domain.Quota{
				SMS:   dailyQuota.Sms,
				Email: dailyQuota.Email,
			}
		}
		if pb.QuotaConfig.Monthly != nil {
			monthlyQuota := pb.QuotaConfig.Monthly
			quotaConfig.Monthly = &domain.Quota{
				SMS:   monthlyQuota.Sms,
				Email: monthlyQuota.Email,
			}
		}
		bizConfig.QuotaConfig = quotaConfig
	}

	if pb.CallbackConfig != nil {
		callbackConfig := &domain.CallbackConfig{
			ServiceName: pb.CallbackConfig.ServiceName,
		}

		if pb.CallbackConfig.RetryPolicy != nil {
			retryPolicyConfig := s.convertRetry(pb.CallbackConfig.RetryPolicy)
			callbackConfig.RetryPolicyConfig = retryPolicyConfig
		}
		bizConfig.CallbackConfig = callbackConfig
	}
	return bizConfig
}

func (s *BizConfigServer) convertRetry(pbRetry *configv1.RetryPolicyConfig) *retry.Config {
	return &retry.Config{
		Type: retry.StrategyTypeExponentialBackoff,
		ExponentialBackoff: &retry.ExponentialBackoffConfig{
			InitialInterval: time.Duration(pbRetry.InitIntervalMs) * time.Millisecond,
			MaxInterval:     time.Duration(pbRetry.MaxIntervalMs) * time.Millisecond,
			MaxRetryTimes:   pbRetry.MaxRetryTimes,
		},
	}
}

func NewBizConfigServer(svc bizconf.BizConfigService) *BizConfigServer {
	return &BizConfigServer{
		svc: svc,
	}
}

package api

import (
	"context"
	"time"

	configv1 "github.com/JrMarcco/kuryr-api/api/config/v1"
	"github.com/JrMarcco/kuryr/internal/api/interceptor/jwt"
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

	bizId, err := jwt.ContextBizId(ctx)
	if err != nil {
		return &configv1.SaveResponse{
			Success: false,
			ErrMsg:  err.Error(),
		}, err
	}

	bizConfig := s.protoToDomain(request.Config)
	bizConfig.OwnerId = bizId

	err = s.svc.Save(ctx, bizConfig)
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
		RateLimit: int(pb.RateLimit),
	}

	if pb.ChannelConfig != nil {
		channelConfig := &domain.ChannelConfig{
			Channels: make([]domain.ChannelItem, 0, len(pb.ChannelConfig.Items)),
		}

		for _, channelItem := range pb.ChannelConfig.Items {
			channelConfig.Channels = append(channelConfig.Channels, domain.ChannelItem{
				Channel:  channelItem.Channel,
				Priority: int(channelItem.Priority),
				Enabled:  channelItem.Enabled,
			})
		}

		if pb.ChannelConfig.RetryPolicy != nil {
			retryPolicyConfig := s.convertRetry(pb.ChannelConfig.RetryPolicy)
			channelConfig.RetryPolicyConfig = retryPolicyConfig
		}
		bizConfig.ChannelConfig = channelConfig
	}

	if pb.Quota != nil {
		quotaConfig := &domain.QuotaConfig{}
		if pb.Quota.Daily != nil {
			dailyQuota := pb.Quota.Daily
			quotaConfig.DailyQuota = &domain.QuotaDetail{
				SMS:   dailyQuota.Sms,
				Email: dailyQuota.Email,
			}
		}
		if pb.Quota.Monthly != nil {
			monthlyQuota := pb.Quota.Monthly
			quotaConfig.MonthlyQuota = &domain.QuotaDetail{
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
		Type: "exponential_backoff",
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

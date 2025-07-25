package api

import (
	"context"
	"fmt"
	"time"

	"github.com/JrMarcco/easy-kit/slice"
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

	bizConfig := s.pbToDomain(request.Config)

	err := s.svc.Save(ctx, bizConfig)
	if err != nil {
		return &configv1.SaveResponse{
			Success: false,
			ErrMsg:  err.Error(),
		}, err
	}

	return &configv1.SaveResponse{Success: true}, nil
}

// pbToDomain convert protobuf to domain
func (s *BizConfigServer) pbToDomain(pb *configv1.BizConfig) domain.BizConfig {
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

func (s *BizConfigServer) Delete(ctx context.Context, request *configv1.DeleteRequest) (*configv1.DeleteResponse, error) {
	if request.Id == 0 {
		return &configv1.DeleteResponse{
			Success: false,
			ErrMsg:  "[kuryr] biz id is invalid",
		}, nil
	}

	err := s.svc.Delete(ctx, request.Id)
	if err != nil {
		return &configv1.DeleteResponse{
			Success: false,
			ErrMsg:  err.Error(),
		}, err
	}
	return &configv1.DeleteResponse{Success: true}, nil
}

func (s *BizConfigServer) GetById(ctx context.Context, req *configv1.GetByIdRequest) (*configv1.GetByIdResponse, error) {
	if req.Id == 0 {
		return &configv1.GetByIdResponse{}, fmt.Errorf("[kuryr] biz id is invalid: %d", req.Id)
	}
	bizConfig, err := s.svc.GetById(ctx, req.Id)
	if err != nil {
		return &configv1.GetByIdResponse{}, fmt.Errorf("[kuryr] failed to get biz config by id: %w", err)
	}
	return &configv1.GetByIdResponse{Config: s.domainToPb(bizConfig)}, nil
}

func (s *BizConfigServer) GetByIds(ctx context.Context, req *configv1.GetByIdsRequest) (*configv1.GetByIdsResponse, error) {
	panic("implement me")
}

func (s *BizConfigServer) domainToPb(bizConfig domain.BizConfig) *configv1.BizConfig {
	pb := &configv1.BizConfig{
		BizId:     bizConfig.Id,
		RateLimit: int32(bizConfig.RateLimit),
	}

	if bizConfig.ChannelConfig != nil {
		items := make([]*configv1.ChannelItem, len(bizConfig.ChannelConfig.Channels))
		slice.Map(bizConfig.ChannelConfig.Channels, func(_ int, src domain.ChannelItem) configv1.ChannelItem {
			return configv1.ChannelItem{
				Channel:  src.Channel,
				Priority: int32(src.Priority),
				Enabled:  src.Enabled,
			}
		})

		retryPolicyConfig := bizConfig.ChannelConfig.RetryPolicyConfig.ExponentialBackoff
		pb.ChannelConfig = &configv1.ChannelConfig{
			Items: items,
			RetryPolicy: &configv1.RetryPolicyConfig{
				InitIntervalMs: int32(retryPolicyConfig.InitialInterval.Milliseconds()),
				MaxIntervalMs:  int32(retryPolicyConfig.MaxInterval.Milliseconds()),
				MaxRetryTimes:  retryPolicyConfig.MaxRetryTimes,
			},
		}
	}

	if bizConfig.QuotaConfig != nil {
		quotaConfig := &configv1.QuotaConfig{}
		if bizConfig.QuotaConfig.Daily != nil {
			dailyQuota := bizConfig.QuotaConfig.Daily
			quotaConfig.Daily = &configv1.Quota{
				Sms:   dailyQuota.SMS,
				Email: dailyQuota.Email,
			}
		}
		if bizConfig.QuotaConfig.Monthly != nil {
			monthlyQuota := bizConfig.QuotaConfig.Monthly
			quotaConfig.Monthly = &configv1.Quota{
				Sms:   monthlyQuota.SMS,
				Email: monthlyQuota.Email,
			}
		}
		pb.QuotaConfig = quotaConfig
	}

	if bizConfig.CallbackConfig != nil {
		retryPolicyConfig := bizConfig.CallbackConfig.RetryPolicyConfig.ExponentialBackoff
		pb.CallbackConfig = &configv1.CallbackConfig{
			ServiceName: bizConfig.CallbackConfig.ServiceName,
			RetryPolicy: &configv1.RetryPolicyConfig{
				InitIntervalMs: int32(retryPolicyConfig.InitialInterval.Milliseconds()),
				MaxIntervalMs:  int32(retryPolicyConfig.MaxInterval.Milliseconds()),
				MaxRetryTimes:  retryPolicyConfig.MaxRetryTimes,
			},
		}
	}

	return pb
}

func NewBizConfigServer(svc bizconf.BizConfigService) *BizConfigServer {
	return &BizConfigServer{
		svc: svc,
	}
}

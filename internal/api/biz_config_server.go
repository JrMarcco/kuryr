package api

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/JrMarcco/easy-kit/slice"
	commonv1 "github.com/JrMarcco/kuryr-api/api/common/v1"
	configv1 "github.com/JrMarcco/kuryr-api/api/config/v1"
	"github.com/JrMarcco/kuryr/internal/domain"
	"github.com/JrMarcco/kuryr/internal/errs"
	"github.com/JrMarcco/kuryr/internal/pkg/retry"
	"github.com/JrMarcco/kuryr/internal/service/bizconf"
)

var _ configv1.BizConfigServiceServer = (*BizConfigServer)(nil)

type BizConfigServer struct {
	svc bizconf.Service
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
		OwnerType: domain.OwnerType(pb.OwnerType),
		RateLimit: pb.RateLimit,
	}

	if pb.ChannelConfig != nil {
		channelConfig := &domain.ChannelConfig{
			Channels: make([]domain.ChannelItem, len(pb.ChannelConfig.Items)),
		}

		for index, item := range pb.ChannelConfig.Items {
			channelConfig.Channels[index] = domain.ChannelItem{
				Channel:  domain.Channel(item.Channel),
				Priority: item.Priority,
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
				Sms:   dailyQuota.Sms,
				Email: dailyQuota.Email,
			}
		}
		if pb.QuotaConfig.Monthly != nil {
			monthlyQuota := pb.QuotaConfig.Monthly
			quotaConfig.Monthly = &domain.Quota{
				Sms:   monthlyQuota.Sms,
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

func (s *BizConfigServer) FindById(ctx context.Context, req *configv1.FindByIdRequest) (*configv1.FindByIdResponse, error) {
	if req.Id == 0 {
		return &configv1.FindByIdResponse{}, fmt.Errorf("[kuryr] biz id is invalid: %d", req.Id)
	}

	bizConfig, err := s.svc.FindById(ctx, req.Id)
	if err != nil {
		if errors.Is(err, errs.ErrRecordNotFound) {
			return &configv1.FindByIdResponse{
				ErrCode: commonv1.ErrCode_BIZ_CONFIG_NOT_FOUND,
			}, nil
		}
		return &configv1.FindByIdResponse{}, fmt.Errorf("%w: failed to get biz config by id", err)
	}
	return &configv1.FindByIdResponse{Config: s.domainToPb(bizConfig)}, nil
}

func (s *BizConfigServer) domainToPb(bizConfig domain.BizConfig) *configv1.BizConfig {
	pb := &configv1.BizConfig{
		BizId:     bizConfig.Id,
		OwnerType: string(bizConfig.OwnerType),
		RateLimit: bizConfig.RateLimit,
	}

	if bizConfig.ChannelConfig != nil {
		items := slice.Map(bizConfig.ChannelConfig.Channels, func(_ int, src domain.ChannelItem) *configv1.ChannelItem {
			return &configv1.ChannelItem{
				Channel:  commonv1.Channel(src.Channel),
				Priority: src.Priority,
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
				Sms:   dailyQuota.Sms,
				Email: dailyQuota.Email,
			}
		}
		if bizConfig.QuotaConfig.Monthly != nil {
			monthlyQuota := bizConfig.QuotaConfig.Monthly
			quotaConfig.Monthly = &configv1.Quota{
				Sms:   monthlyQuota.Sms,
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

func NewBizConfigServer(svc bizconf.Service) *BizConfigServer {
	return &BizConfigServer{
		svc: svc,
	}
}

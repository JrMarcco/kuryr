package api

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/JrMarcco/easy-kit/slice"
	commonv1 "github.com/JrMarcco/kuryr-api/api/go/common/v1"
	configv1 "github.com/JrMarcco/kuryr-api/api/go/config/v1"
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
	if request == nil || request.BizId == 0 {
		return &configv1.SaveResponse{}, fmt.Errorf("%w: request is nil or biz id is invalid", errs.ErrInvalidParam)
	}

	bizConfig := s.saveReqToDomain(request)
	saved, err := s.svc.Save(ctx, bizConfig)
	if err != nil {
		return &configv1.SaveResponse{}, err
	}

	return &configv1.SaveResponse{
		BizConfig: s.domainToPb(saved),
	}, nil
}

// saveReqToDomain convert save request protobuf to domain
func (s *BizConfigServer) saveReqToDomain(req *configv1.SaveRequest) domain.BizConfig {
	bizConfig := domain.BizConfig{
		Id:        req.BizId,
		RateLimit: req.RateLimit,
	}

	if req.ChannelConfig != nil {
		channelConfig := &domain.ChannelConfig{
			Channels: make([]domain.ChannelItem, len(req.ChannelConfig.Items)),
		}

		for index, item := range req.ChannelConfig.Items {
			channelConfig.Channels[index] = domain.ChannelItem{
				Channel:  domain.Channel(item.Channel),
				Priority: item.Priority,
				Enabled:  item.Enabled,
			}
		}

		if req.ChannelConfig.RetryPolicy != nil {
			retryPolicyConfig := s.convertRetry(req.ChannelConfig.RetryPolicy)
			channelConfig.RetryPolicyConfig = retryPolicyConfig
		}
		bizConfig.ChannelConfig = channelConfig
	}

	if req.QuotaConfig != nil {
		quotaConfig := &domain.QuotaConfig{}
		if req.QuotaConfig.Daily != nil {
			dailyQuota := req.QuotaConfig.Daily
			quotaConfig.Daily = &domain.Quota{
				Sms:   dailyQuota.Sms,
				Email: dailyQuota.Email,
			}
		}
		if req.QuotaConfig.Monthly != nil {
			monthlyQuota := req.QuotaConfig.Monthly
			quotaConfig.Monthly = &domain.Quota{
				Sms:   monthlyQuota.Sms,
				Email: monthlyQuota.Email,
			}
		}
		bizConfig.QuotaConfig = quotaConfig
	}

	if req.CallbackConfig != nil {
		callbackConfig := &domain.CallbackConfig{
			ServiceName: req.CallbackConfig.ServiceName,
		}

		if req.CallbackConfig.RetryPolicy != nil {
			retryPolicyConfig := s.convertRetry(req.CallbackConfig.RetryPolicy)
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
		return &configv1.DeleteResponse{}, fmt.Errorf("%w: biz id is invalid", errs.ErrInvalidParam)
	}

	err := s.svc.Delete(ctx, request.Id)
	if err != nil {
		return &configv1.DeleteResponse{}, err
	}
	return &configv1.DeleteResponse{}, nil
}

func (s *BizConfigServer) FindById(ctx context.Context, req *configv1.FindByIdRequest) (*configv1.FindByIdResponse, error) {
	if req.Id == 0 {
		return &configv1.FindByIdResponse{}, fmt.Errorf("%w: biz id is invalid", errs.ErrInvalidParam)
	}

	bizConfig, err := s.svc.FindById(ctx, req.Id)
	if err != nil {
		if errors.Is(err, errs.ErrRecordNotFound) {
			// TODO: deal with record not found
			return &configv1.FindByIdResponse{}, nil
		}
		return &configv1.FindByIdResponse{}, err
	}
	return &configv1.FindByIdResponse{
		Config: s.domainToPb(bizConfig),
	}, nil
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

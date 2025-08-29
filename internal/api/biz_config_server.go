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
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
)

var _ configv1.BizConfigServiceServer = (*BizConfigServer)(nil)

type BizConfigServer struct {
	svc bizconf.Service
}

func (s *BizConfigServer) Save(ctx context.Context, request *configv1.SaveRequest) (*configv1.SaveResponse, error) {
	if request == nil || request.BizConfig == nil {
		return &configv1.SaveResponse{}, status.Errorf(codes.InvalidArgument, "request or biz config is nil")
	}

	bizConfig := s.pbToDomain(request.BizConfig)
	saved, err := s.svc.Save(ctx, bizConfig)
	if err != nil {
		return &configv1.SaveResponse{}, status.Errorf(codes.Internal, "failed to save biz config: %v", err)
	}

	return &configv1.SaveResponse{
		BizConfig: s.domainToPb(saved),
	}, nil
}

// pbToDomain convert biz config protobuf to domain
func (s *BizConfigServer) pbToDomain(pb *configv1.BizConfig) domain.BizConfig {
	bizConfig := domain.BizConfig{
		BizId:     pb.BizId,
		RateLimit: pb.RateLimit,
	}

	if pb.ChannelConfig != nil {
		bizConfig.ChannelConfig = s.convertChannelConfigPb(pb.ChannelConfig)
	}

	if pb.QuotaConfig != nil {
		bizConfig.QuotaConfig = s.convertQuotaConfigPb(pb.QuotaConfig)
	}

	if pb.CallbackConfig != nil {
		bizConfig.CallbackConfig = s.convertCallbackConfigPb(pb.CallbackConfig)
	}
	return bizConfig
}

func (s *BizConfigServer) Update(ctx context.Context, request *configv1.UpdateRequest) (*configv1.UpdateResponse, error) {
	if request == nil || request.BizConfig == nil {
		return &configv1.UpdateResponse{}, status.Errorf(codes.InvalidArgument, "request or biz config is nil")
	}

	bizConfig, err := s.applyMaskToDomain(request.BizConfig, request.FieldMask)
	if err != nil {
		return &configv1.UpdateResponse{}, status.Errorf(codes.InvalidArgument, "invalid biz config: %v", err)
	}

	updated, err := s.svc.Update(ctx, bizConfig)
	if err != nil {
		return &configv1.UpdateResponse{}, status.Errorf(codes.Internal, "failed to update biz config: %v", err)
	}

	return &configv1.UpdateResponse{
		BizConfig: s.domainToPb(updated),
	}, nil
}

func (s *BizConfigServer) applyMaskToDomain(pb *configv1.BizConfig, mask *fieldmaskpb.FieldMask) (domain.BizConfig, error) {
	if mask == nil || len(mask.Paths) == 0 {
		return domain.BizConfig{}, fmt.Errorf("%w: field mask is nil or paths is empty", errs.ErrInvalidParam)
	}

	bizConfig := domain.BizConfig{
		Id: pb.BizId,
	}

	for _, field := range mask.Paths {
		if _, ok := configv1.UpdatableFields[field]; !ok {
			return domain.BizConfig{}, fmt.Errorf("%w: field [ %s ] is not updatable", errs.ErrInvalidParam, field)
		}

		switch field {
		case configv1.FieldChannelConfig:
			bizConfig.ChannelConfig = s.convertChannelConfigPb(pb.ChannelConfig)
		case configv1.FieldQuotaConfig:
			bizConfig.QuotaConfig = s.convertQuotaConfigPb(pb.QuotaConfig)
		case configv1.FieldCallbackConfig:
			bizConfig.CallbackConfig = s.convertCallbackConfigPb(pb.CallbackConfig)
		case configv1.FieldRateLimit:
			bizConfig.RateLimit = pb.RateLimit
		}
	}

	return bizConfig, nil
}

func (s *BizConfigServer) convertChannelConfigPb(pb *configv1.ChannelConfig) *domain.ChannelConfig {
	channelConfig := &domain.ChannelConfig{
		Channels: make([]domain.ChannelItem, len(pb.Items)),
	}

	for index, item := range pb.Items {
		channelConfig.Channels[index] = domain.ChannelItem{
			Channel:  domain.Channel(item.Channel),
			Priority: item.Priority,
			Enabled:  item.Enabled,
		}
	}

	if pb.RetryPolicy != nil {
		retryPolicyConfig := s.convertRetryPb(pb.RetryPolicy)
		channelConfig.RetryPolicyConfig = retryPolicyConfig
	}

	return channelConfig
}

func (s *BizConfigServer) convertQuotaConfigPb(pb *configv1.QuotaConfig) *domain.QuotaConfig {
	quotaConfig := &domain.QuotaConfig{}
	if pb.Daily != nil {
		dailyQuota := pb.Daily
		quotaConfig.Daily = &domain.Quota{
			Sms:   dailyQuota.Sms,
			Email: dailyQuota.Email,
		}
	}
	if pb.Monthly != nil {
		monthlyQuota := pb.Monthly
		quotaConfig.Monthly = &domain.Quota{
			Sms:   monthlyQuota.Sms,
			Email: monthlyQuota.Email,
		}
	}
	return quotaConfig
}

func (s *BizConfigServer) convertCallbackConfigPb(pb *configv1.CallbackConfig) *domain.CallbackConfig {
	callbackConfig := &domain.CallbackConfig{
		ServiceName: pb.ServiceName,
	}

	if pb.RetryPolicy != nil {
		retryPolicyConfig := s.convertRetryPb(pb.RetryPolicy)
		callbackConfig.RetryPolicyConfig = retryPolicyConfig
	}

	return callbackConfig
}

func (s *BizConfigServer) convertRetryPb(pbRetry *configv1.RetryPolicyConfig) *retry.Config {
	return &retry.Config{
		Type: retry.StrategyTypeExponentialBackoff,
		ExponentialBackoff: &retry.ExponentialBackoffConfig{
			InitialInterval: time.Duration(pbRetry.InitIntervalMs) * time.Millisecond,
			MaxInterval:     time.Duration(pbRetry.MaxIntervalMs) * time.Millisecond,
			MaxRetryTimes:   pbRetry.MaxRetryTimes,
		},
	}
}

func (s *BizConfigServer) FindByBizId(ctx context.Context, req *configv1.FindByBizIdRequest) (*configv1.FindByBizIdResponse, error) {
	if req.BizId == 0 {
		return &configv1.FindByBizIdResponse{}, fmt.Errorf("%w: biz id is invalid", errs.ErrInvalidParam)
	}

	bizConfig, err := s.svc.FindByBizId(ctx, req.BizId)
	if err != nil {
		if errors.Is(err, errs.ErrRecordNotFound) {
			return &configv1.FindByBizIdResponse{}, status.Errorf(codes.NotFound, "biz config not found: %v", err)
		}
		return &configv1.FindByBizIdResponse{}, err
	}
	return &configv1.FindByBizIdResponse{
		BizConfig: s.applyMaskToPb(bizConfig, req.FieldMask),
	}, nil
}

func (s *BizConfigServer) applyMaskToPb(bizConfig domain.BizConfig, mask *fieldmaskpb.FieldMask) *configv1.BizConfig {
	pb := s.domainToPb(bizConfig)
	if mask == nil || len(mask.Paths) == 0 {
		return pb
	}

	res := &configv1.BizConfig{}
	for _, field := range mask.Paths {
		switch field {
		case configv1.FieldId:
			res.Id = pb.Id
		case configv1.FieldBizId:
			res.BizId = pb.BizId
		case configv1.FieldChannelConfig:
			res.ChannelConfig = pb.ChannelConfig
		case configv1.FieldQuotaConfig:
			res.QuotaConfig = pb.QuotaConfig
		case configv1.FieldCallbackConfig:
			res.CallbackConfig = pb.CallbackConfig
		case configv1.FieldRateLimit:
			res.RateLimit = pb.RateLimit
		case configv1.FieldCreatedAt:
			res.CreatedAt = pb.CreatedAt
		case configv1.FieldUpdatedAt:
			res.UpdatedAt = pb.UpdatedAt
		}
	}

	return res
}

func (s *BizConfigServer) domainToPb(bizConfig domain.BizConfig) *configv1.BizConfig {
	pb := &configv1.BizConfig{
		BizId:     bizConfig.BizId,
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

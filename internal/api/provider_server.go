package api

import (
	"context"
	"fmt"

	"github.com/JrMarcco/easy-kit/slice"
	commonv1 "github.com/JrMarcco/kuryr-api/api/go/common/v1"
	providerv1 "github.com/JrMarcco/kuryr-api/api/go/provider/v1"
	"github.com/JrMarcco/kuryr/internal/domain"
	"github.com/JrMarcco/kuryr/internal/errs"
	"github.com/JrMarcco/kuryr/internal/service/provider"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
)

var _ providerv1.ProviderServiceServer = (*ProviderServer)(nil)

type ProviderServer struct {
	svc provider.Service
}

func (s *ProviderServer) Save(ctx context.Context, request *providerv1.SaveRequest) (*providerv1.SaveResponse, error) {
	if request == nil || request.Provider == nil {
		return &providerv1.SaveResponse{}, status.Errorf(codes.InvalidArgument, "request or provider is nil")
	}

	saved, err := s.svc.Save(ctx, s.pbToDomain(request.Provider))
	if err != nil {
		return &providerv1.SaveResponse{}, status.Errorf(codes.Internal, "failed to save provider: %v", err)
	}

	return &providerv1.SaveResponse{
		Provider: s.domainToPb(saved),
	}, nil
}

func (s *ProviderServer) pbToDomain(pb *providerv1.Provider) domain.Provider {
	return domain.Provider{
		Id:               pb.Id,
		ProviderName:     pb.ProviderName,
		Channel:          domain.Channel(pb.Channel),
		Endpoint:         pb.Endpoint,
		RegionId:         pb.RegionId,
		AppId:            pb.AppId,
		ApiKey:           pb.ApiKey,
		ApiSecret:        pb.ApiSecret,
		Weight:           pb.Weight,
		QpsLimit:         pb.QpsLimit,
		DailyLimit:       pb.DailyLimit,
		AuditCallbackUrl: pb.AuditCallbackUrl,
		ActiveStatus:     domain.ActiveStatus(pb.ActiveStatus),
	}
}

func (s *ProviderServer) Delete(ctx context.Context, request *providerv1.DeleteRequest) (*providerv1.DeleteResponse, error) {
	// TODO: implement me
	panic("implement me")
}

func (s *ProviderServer) Update(ctx context.Context, request *providerv1.UpdateRequest) (*providerv1.UpdateResponse, error) {
	if request == nil || request.Provider == nil {
		return &providerv1.UpdateResponse{}, status.Errorf(codes.InvalidArgument, "request or provider is nil")
	}

	p, err := s.applyMaskToDomain(request.Provider, request.FieldMask)
	if err != nil {
		return &providerv1.UpdateResponse{}, status.Errorf(codes.InvalidArgument, "invalid provider: %v", err)
	}

	// updated, err := s.svc.Update(ctx, p)
	_, err = s.svc.Update(ctx, p)
	if err != nil {
		return &providerv1.UpdateResponse{}, status.Errorf(codes.Internal, "failed to update provider: %v", err)
	}
	return &providerv1.UpdateResponse{
		// TODO: 返回更新后的 provider
		// Provider: s.domainToPb(updated),
	}, nil
}

func (s *ProviderServer) applyMaskToDomain(pb *providerv1.Provider, mask *fieldmaskpb.FieldMask) (domain.Provider, error) {
	if mask == nil || len(mask.Paths) == 0 {
		return domain.Provider{}, fmt.Errorf("%w: field mask is nil or paths is empty", errs.ErrInvalidParam)
	}

	p := domain.Provider{
		Id: pb.Id,
	}

	for _, field := range mask.Paths {
		if _, ok := providerv1.UpdatableFields[field]; !ok {
			return domain.Provider{}, fmt.Errorf("%w: field [ %s ] is not updatable", errs.ErrInvalidParam, field)
		}

		switch field {
		case providerv1.FieldProviderName:
			p.ProviderName = pb.ProviderName
		case providerv1.FieldChannel:
			p.Channel = domain.Channel(pb.Channel)
		case providerv1.FieldEndpoint:
			p.Endpoint = pb.Endpoint
		case providerv1.FieldRegionId:
			p.RegionId = pb.RegionId
		case providerv1.FieldAppId:
			p.AppId = pb.AppId
		case providerv1.FieldApiKey:
			p.ApiKey = pb.ApiKey
		case providerv1.FieldApiSecret:
			p.ApiSecret = pb.ApiSecret
		case providerv1.FieldWeight:
			p.Weight = pb.Weight
		case providerv1.FieldQpsLimit:
			p.QpsLimit = pb.QpsLimit
		case providerv1.FieldDailyLimit:
			p.DailyLimit = pb.DailyLimit
		case providerv1.FieldAuditCallbackUrl:
			p.AuditCallbackUrl = pb.AuditCallbackUrl
		case providerv1.FieldActiveStatus:
			p.ActiveStatus = domain.ActiveStatus(pb.ActiveStatus)
		}
	}

	return p, nil
}

func (s *ProviderServer) List(ctx context.Context, request *providerv1.ListRequest) (*providerv1.ListResponse, error) {
	if request == nil {
		return &providerv1.ListResponse{}, fmt.Errorf("[kuryr] invalid request, provider is nil")
	}

	providers, err := s.svc.List(ctx)
	if err != nil {
		return &providerv1.ListResponse{}, fmt.Errorf("[kuryr] failed to list providers: %w", err)
	}

	pbs := slice.Map(providers, func(_ int, p domain.Provider) *providerv1.Provider {
		return s.applyMaskToPb(p, request.FieldMask)
	})
	return &providerv1.ListResponse{
		Providers: pbs,
	}, nil
}

func (s *ProviderServer) FindById(ctx context.Context, request *providerv1.FindByIdRequest) (*providerv1.FindByIdResponse, error) {
	if request.Id == 0 {
		return &providerv1.FindByIdResponse{}, fmt.Errorf("[kuryr] invalid request, provider id is invalid [ %d ]", request.Id)
	}

	p, err := s.svc.FindById(ctx, request.Id)
	if err != nil {
		return &providerv1.FindByIdResponse{}, fmt.Errorf("[kuryr] failed to find provider by id: %w", err)
	}

	return &providerv1.FindByIdResponse{
		Provider: s.applyMaskToPb(p, request.FieldMask),
	}, nil
}

func (s *ProviderServer) FindByChannel(ctx context.Context, request *providerv1.FindByChannelRequest) (*providerv1.FindByChannelResponse, error) {
	if request == nil {
		return &providerv1.FindByChannelResponse{}, fmt.Errorf("[kuryr] invalid request, provider is nil")
	}

	providers, err := s.svc.FindByChannel(ctx, domain.Channel(request.Channel))
	if err != nil {
		return &providerv1.FindByChannelResponse{}, fmt.Errorf("[kuryr] failed to find provider by channel: %w", err)
	}

	pbs := slice.Map(providers, func(_ int, p domain.Provider) *providerv1.Provider {
		return s.applyMaskToPb(p, request.FieldMask)
	})
	return &providerv1.FindByChannelResponse{
		Providers: pbs,
	}, nil
}

func (s *ProviderServer) applyMaskToPb(provider domain.Provider, mask *fieldmaskpb.FieldMask) *providerv1.Provider {
	if mask == nil || len(mask.Paths) == 0 {
		return s.domainToPb(provider)
	}

	pb := &providerv1.Provider{}

	for _, field := range mask.Paths {
		switch field {
		case providerv1.FieldId:
			pb.Id = provider.Id
		case providerv1.FieldProviderName:
			pb.ProviderName = provider.ProviderName
		case providerv1.FieldChannel:
			pb.Channel = commonv1.Channel_CHANNEL_UNSPECIFIED
			switch provider.Channel {
			case domain.ChannelSms:
				pb.Channel = commonv1.Channel_SMS
			case domain.ChannelEmail:
				pb.Channel = commonv1.Channel_EMAIL
			}
		case providerv1.FieldEndpoint:
			pb.Endpoint = provider.Endpoint
		case providerv1.FieldRegionId:
			pb.RegionId = provider.RegionId
		case providerv1.FieldAppId:
			pb.AppId = provider.AppId
		case providerv1.FieldApiKey:
			pb.ApiKey = provider.ApiKey
		case providerv1.FieldApiSecret:
			pb.ApiSecret = provider.ApiSecret
		case providerv1.FieldWeight:
			pb.Weight = provider.Weight
		case providerv1.FieldQpsLimit:
			pb.QpsLimit = provider.QpsLimit
		case providerv1.FieldDailyLimit:
			pb.DailyLimit = provider.DailyLimit
		case providerv1.FieldAuditCallbackUrl:
			pb.AuditCallbackUrl = provider.AuditCallbackUrl
		case providerv1.FieldActiveStatus:
			pb.ActiveStatus = string(provider.ActiveStatus)
		}
	}
	return pb
}

func (s *ProviderServer) domainToPb(p domain.Provider) *providerv1.Provider {
	channel := commonv1.Channel_CHANNEL_UNSPECIFIED
	switch p.Channel {
	case domain.ChannelSms:
		channel = commonv1.Channel_SMS
	case domain.ChannelEmail:
		channel = commonv1.Channel_EMAIL
	}

	return &providerv1.Provider{
		Id:               p.Id,
		ProviderName:     p.ProviderName,
		Channel:          channel,
		Endpoint:         p.Endpoint,
		RegionId:         p.RegionId,
		AppId:            p.AppId,
		ApiKey:           p.ApiKey,
		ApiSecret:        p.ApiSecret,
		Weight:           p.Weight,
		QpsLimit:         p.QpsLimit,
		DailyLimit:       p.DailyLimit,
		AuditCallbackUrl: p.AuditCallbackUrl,
		ActiveStatus:     string(p.ActiveStatus),
	}
}

func NewProviderServer(svc provider.Service) *ProviderServer {
	return &ProviderServer{
		svc: svc,
	}
}

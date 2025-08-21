package api

import (
	"context"
	"fmt"

	"github.com/JrMarcco/easy-kit/slice"
	commonv1 "github.com/JrMarcco/kuryr-api/api/go/common/v1"
	providerv1 "github.com/JrMarcco/kuryr-api/api/go/provider/v1"
	"github.com/JrMarcco/kuryr/internal/domain"
	"github.com/JrMarcco/kuryr/internal/service/provider"
)

var _ providerv1.ProviderServiceServer = (*ProviderServer)(nil)

type ProviderServer struct {
	svc provider.Service
}

func (s *ProviderServer) Save(ctx context.Context, request *providerv1.SaveRequest) (*providerv1.SaveResponse, error) {
	// if request == nil || request.Provider == nil {
	// 	return &providerv1.SaveResponse{
	// 		Success: false,
	// 		ErrMsg:  "[kuryr] invalid request, provider is nil",
	// 	}, nil
	// }

	// p, err := s.pbToDomain(request.Provider)
	// if err != nil {
	// 	return &providerv1.SaveResponse{}, err
	// }

	// if err = s.svc.Save(ctx, p); err != nil {
	// 	return &providerv1.SaveResponse{}, err
	// }
	// return &providerv1.SaveResponse{Success: true}, nil
	// TODO: implement me
	panic("implement me")
}

func (s *ProviderServer) Delete(ctx context.Context, request *providerv1.DeleteRequest) (*providerv1.DeleteResponse, error) {
	// TODO: implement me
	panic("implement me")
}

func (s *ProviderServer) Update(ctx context.Context, request *providerv1.UpdateRequest) (*providerv1.UpdateResponse, error) {
	// if request == nil || request.Provider == nil || request.Provider.Id == 0 {
	// 	return &providerv1.UpdateResponse{
	// 		Success: false,
	// 		ErrMsg:  "[kuryr] invalid request, provider is nil or provider id is invalid",
	// 	}, nil
	// }

	// p, err := s.pbToDomain(request.Provider)
	// if err != nil {
	// 	return &providerv1.UpdateResponse{}, err
	// }

	// if err = s.svc.Update(ctx, p); err != nil {
	// 	return &providerv1.UpdateResponse{
	// 		Success: false,
	// 		ErrMsg:  err.Error(),
	// 	}, err
	// }
	// return &providerv1.UpdateResponse{Success: true}, nil
	// TODO: implement me
	panic("implement me")
}

// func (s *ProviderServer) pbToDomain(pb *providerv1.Provider) (domain.Provider, error) {
// 	p := domain.Provider{
// 		Id:               pb.Id,
// 		ProviderName:     pb.ProviderName,
// 		Endpoint:         pb.Endpoint,
// 		RegionId:         pb.RegionId,
// 		AppId:            pb.AppId,
// 		ApiKey:           pb.ApiKey,
// 		ApiSecret:        pb.ApiSecret,
// 		Weight:           pb.Weight,
// 		QpsLimit:         pb.QpsLimit,
// 		DailyLimit:       pb.DailyLimit,
// 		AuditCallbackUrl: pb.AuditCallbackUrl,
// 		ActiveStatus:     domain.ActiveStatus(pb.ActiveStatus),
// 	}

// 	switch pb.Channel {
// 	case commonv1.Channel_SMS:
// 		p.Channel = domain.ChannelSms
// 	case commonv1.Channel_EMAIL:
// 		p.Channel = domain.ChannelEmail
// 	default:
// 		return domain.Provider{}, fmt.Errorf("[kuryr] invalid channel: %s", pb.Channel.String())
// 	}
// 	return p, nil
// }

func (s *ProviderServer) List(ctx context.Context, request *providerv1.ListRequest) (*providerv1.ListResponse, error) {
	if request == nil {
		return &providerv1.ListResponse{}, fmt.Errorf("[kuryr] invalid request, provider is nil")
	}

	providers, err := s.svc.List(ctx)
	if err != nil {
		return &providerv1.ListResponse{}, fmt.Errorf("[kuryr] failed to list providers: %w", err)
	}

	pbs := slice.Map(providers, func(_ int, p domain.Provider) *providerv1.Provider {
		return s.domainToPb(p)
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
		Provider: s.domainToPb(p),
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
		return s.domainToPb(p)
	})
	return &providerv1.FindByChannelResponse{
		Providers: pbs,
	}, nil
}

func (s *ProviderServer) domainToPb(p domain.Provider) *providerv1.Provider {
	pb := &providerv1.Provider{
		Id:               p.Id,
		ProviderName:     p.ProviderName,
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

	switch p.Channel {
	case domain.ChannelSms:
		pb.Channel = commonv1.Channel_SMS
	case domain.ChannelEmail:
		pb.Channel = commonv1.Channel_EMAIL
	default:
	}
	return pb
}

func NewProviderServer(svc provider.Service) *ProviderServer {
	return &ProviderServer{
		svc: svc,
	}
}

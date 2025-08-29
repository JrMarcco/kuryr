package api

import (
	"context"

	"github.com/JrMarcco/easy-kit/slice"
	commonv1 "github.com/JrMarcco/kuryr-api/api/go/common/v1"
	templatev1 "github.com/JrMarcco/kuryr-api/api/go/template/v1"
	"github.com/JrMarcco/kuryr/internal/domain"
	pkggorm "github.com/JrMarcco/kuryr/internal/pkg/gorm"
	"github.com/JrMarcco/kuryr/internal/service/template"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
)

var _ templatev1.TemplateServiceServer = (*TemplateServer)(nil)

type TemplateServer struct {
	svc template.Service
}

func (s *TemplateServer) SaveTemplate(ctx context.Context, request *templatev1.SaveTemplateRequest) (*templatev1.SaveTemplateResponse, error) {
	if request == nil || request.Template == nil {
		return &templatev1.SaveTemplateResponse{}, status.Errorf(codes.InvalidArgument, "request or template is nil")
	}

	saved, err := s.svc.SaveTemplate(ctx, s.pbToTemplateDomain(request.Template))
	if err != nil {
		return &templatev1.SaveTemplateResponse{}, status.Errorf(codes.Internal, "failed to save template: %v", err)
	}

	return &templatev1.SaveTemplateResponse{
		Template: s.templateDomainToPb(saved),
	}, nil
}

func (s *TemplateServer) pbToTemplateDomain(pb *templatev1.ChannelTemplate) domain.ChannelTemplate {
	return domain.ChannelTemplate{
		Id:                 pb.Id,
		BizId:              pb.BizId,
		BizType:            domain.BizType(pb.BizType),
		TplName:            pb.TplName,
		TplDesc:            pb.TplDesc,
		Channel:            domain.Channel(pb.Channel),
		NotificationType:   domain.NotificationType(pb.NotificationType),
		ActivatedVersionId: pb.ActivatedVersionId,
	}
}

func (s *TemplateServer) DeleteTemplate(ctx context.Context, request *templatev1.DeleteTemplateRequest) (*templatev1.DeleteTemplateResponse, error) {
	if request == nil || request.Id == 0 {
		return &templatev1.DeleteTemplateResponse{}, status.Errorf(codes.InvalidArgument, "request or id is nil")
	}

	err := s.svc.DeleteTemplate(ctx, request.Id)
	if err != nil {
		return &templatev1.DeleteTemplateResponse{}, status.Errorf(codes.Internal, "failed to delete template: %v", err)
	}

	return &templatev1.DeleteTemplateResponse{}, nil
}

func (s *TemplateServer) ListTemplateByBizId(ctx context.Context, request *templatev1.ListTemplateByBizIdRequest) (*templatev1.ListTemplateByBizIdResponse, error) {
	if request == nil || request.BizId == 0 {
		return &templatev1.ListTemplateByBizIdResponse{}, status.Errorf(codes.InvalidArgument, "request or biz_id is nil")
	}

	param := &pkggorm.PaginationParam{
		Offset: int(request.Offset),
		Limit:  int(request.Limit),
	}

	if param.Offset < 0 {
		param.Offset = 0
	}
	if param.Limit <= 0 {
		param.Limit = 10
	}

	res, err := s.svc.FindTemplateByBizId(ctx, request.BizId, param)
	if err != nil {
		return &templatev1.ListTemplateByBizIdResponse{}, status.Errorf(codes.Internal, "failed to list template by biz_id: %v", err)
	}

	pbs := slice.Map(res.Records, func(_ int, template domain.ChannelTemplate) *templatev1.ChannelTemplate {
		return s.applyMaskToTemplatePb(template, request.FieldMask)
	})

	return &templatev1.ListTemplateByBizIdResponse{
		Templates: pbs,
		Total:     res.Total,
	}, nil
}

func (s *TemplateServer) applyMaskToTemplatePb(template domain.ChannelTemplate, mask *fieldmaskpb.FieldMask) *templatev1.ChannelTemplate {
	if mask == nil || len(mask.Paths) == 0 {
		return s.templateDomainToPb(template)
	}

	res := &templatev1.ChannelTemplate{}
	for _, field := range mask.Paths {
		switch field {
		case templatev1.FieldId:
			res.Id = template.Id
		case templatev1.FieldBizId:
			res.BizId = template.BizId
		case templatev1.FieldBizType:
			res.BizType = string(template.BizType)
		case templatev1.FieldTplName:
			res.TplName = template.TplName
		case templatev1.FieldTplDesc:
			res.TplDesc = template.TplDesc
		case templatev1.FieldChannel:
			res.Channel = commonv1.Channel(template.Channel)
		case templatev1.FieldNotificationType:
			res.NotificationType = int32(template.NotificationType)
		case templatev1.FieldActivatedVersionId:
			res.ActivatedVersionId = template.ActivatedVersionId
		case templatev1.FieldCreatedAt:
			res.CreatedAt = template.CreatedAt
		case templatev1.FieldUpdatedAt:
			res.UpdatedAt = template.UpdatedAt
		}
	}
	return res
}

func (s *TemplateServer) templateDomainToPb(tpl domain.ChannelTemplate) *templatev1.ChannelTemplate {
	return &templatev1.ChannelTemplate{
		Id:                 tpl.Id,
		BizId:              tpl.BizId,
		BizType:            string(tpl.BizType),
		TplName:            tpl.TplName,
		TplDesc:            tpl.TplDesc,
		Channel:            commonv1.Channel(tpl.Channel),
		NotificationType:   int32(tpl.NotificationType),
		ActivatedVersionId: tpl.ActivatedVersionId,
		CreatedAt:          tpl.CreatedAt,
		UpdatedAt:          tpl.UpdatedAt,
	}
}

func (s *TemplateServer) SaveTemplateVersion(ctx context.Context, request *templatev1.SaveTemplateVersionRequest) (*templatev1.SaveTemplateVersionResponse, error) {
	if request == nil || request.Version == nil {
		return &templatev1.SaveTemplateVersionResponse{}, status.Errorf(codes.InvalidArgument, "request or version is nil")
	}

	saved, err := s.svc.SaveVersion(ctx, s.pbToVersionDomain(request.Version))
	if err != nil {
		return &templatev1.SaveTemplateVersionResponse{}, status.Errorf(codes.Internal, "failed to save template version: %v", err)
	}

	return &templatev1.SaveTemplateVersionResponse{
		Version: s.versionDomainToPb(saved),
	}, nil
}

func (s *TemplateServer) pbToVersionDomain(pb *templatev1.TemplateVersion) domain.ChannelTemplateVersion {
	return domain.ChannelTemplateVersion{
		Id:              pb.Id,
		TplId:           pb.TplId,
		VersionName:     pb.VersionName,
		Signature:       pb.Signature,
		Content:         pb.Content,
		ApplyRemark:     pb.ApplyRemark,
		AuditorId:       pb.AuditorId,
		AuditId:         pb.AuditId,
		AuditTime:       pb.AuditTime,
		AuditStatus:     domain.AuditStatus(pb.AuditStatus),
		RejectionReason: pb.RejectionReason,
		LastReviewAt:    pb.LastReviewAt,
		CreatedAt:       pb.CreatedAt,
		UpdatedAt:       pb.UpdatedAt,
	}
}

func (s *TemplateServer) DeleteTemplateVersion(ctx context.Context, request *templatev1.DeleteTemplateVersionRequest) (*templatev1.DeleteTemplateVersionResponse, error) {
	if request == nil || request.Id == 0 {
		return &templatev1.DeleteTemplateVersionResponse{}, status.Errorf(codes.InvalidArgument, "request or id is nil")
	}

	err := s.svc.DeleteVersion(ctx, request.Id)
	if err != nil {
		return &templatev1.DeleteTemplateVersionResponse{}, status.Errorf(codes.Internal, "failed to delete template version: %v", err)
	}

	return &templatev1.DeleteTemplateVersionResponse{}, nil
}

func (s *TemplateServer) ActivateTemplateVersion(ctx context.Context, request *templatev1.ActivateTemplateVersionRequest) (*templatev1.ActivateTemplateVersionResponse, error) {
	// TODO: implement me
	panic("implement me")
}

func (s *TemplateServer) ListTemplateVersion(ctx context.Context, request *templatev1.ListTemplateVersionRequest) (*templatev1.ListTemplateVersionResponse, error) {
	if request == nil || request.TplId == 0 {
		return &templatev1.ListTemplateVersionResponse{}, status.Errorf(codes.InvalidArgument, "request or tpl_id is nil")
	}

	versions, err := s.svc.FindVersionByTplId(ctx, request.TplId)
	if err != nil {
		return &templatev1.ListTemplateVersionResponse{}, status.Errorf(codes.Internal, "failed to list template version: %v", err)
	}

	pbs := slice.Map(versions, func(_ int, version domain.ChannelTemplateVersion) *templatev1.TemplateVersion {
		return s.applyMaskToVersionPb(version, request.FieldMask)
	})

	return &templatev1.ListTemplateVersionResponse{
		Versions: pbs,
	}, nil
}

func (s *TemplateServer) applyMaskToVersionPb(version domain.ChannelTemplateVersion, mask *fieldmaskpb.FieldMask) *templatev1.TemplateVersion {
	if mask == nil || len(mask.Paths) == 0 {
		return s.versionDomainToPb(version)
	}

	res := &templatev1.TemplateVersion{}
	for _, field := range mask.Paths {
		switch field {
		case templatev1.VersionFieldId:
			res.Id = version.Id
		case templatev1.VersionFieldTplId:
			res.TplId = version.TplId
		case templatev1.VersionFieldVersionName:
			res.VersionName = version.VersionName
		case templatev1.VersionFieldSignature:
			res.Signature = version.Signature
		case templatev1.VersionFieldContent:
			res.Content = version.Content
		case templatev1.VersionFieldApplyRemark:
			res.ApplyRemark = version.ApplyRemark
		case templatev1.VersionFieldAuditorId:
			res.AuditorId = version.AuditorId
		case templatev1.VersionFieldAuditId:
			res.AuditId = version.AuditId
		case templatev1.VersionFieldAuditTime:
			res.AuditTime = version.AuditTime
		case templatev1.VersionFieldAuditStatus:
			res.AuditStatus = string(version.AuditStatus)
		case templatev1.VersionFieldRejectionReason:
			res.RejectionReason = version.RejectionReason
		case templatev1.VersionFieldLastReviewAt:
			res.LastReviewAt = version.LastReviewAt
		case templatev1.VersionFieldCreatedAt:
			res.CreatedAt = version.CreatedAt
		case templatev1.VersionFieldUpdatedAt:
			res.UpdatedAt = version.UpdatedAt
		}
	}
	return res
}

func (s *TemplateServer) versionDomainToPb(version domain.ChannelTemplateVersion) *templatev1.TemplateVersion {
	return &templatev1.TemplateVersion{
		Id:              version.Id,
		TplId:           version.TplId,
		VersionName:     version.VersionName,
		Signature:       version.Signature,
		Content:         version.Content,
		ApplyRemark:     version.ApplyRemark,
		AuditId:         version.AuditId,
		AuditorId:       version.AuditorId,
		AuditTime:       version.AuditTime,
		AuditStatus:     string(version.AuditStatus),
		RejectionReason: version.RejectionReason,
		LastReviewAt:    version.LastReviewAt,
		CreatedAt:       version.CreatedAt,
		UpdatedAt:       version.UpdatedAt,
	}
}

func (s *TemplateServer) SaveTemplateProviders(ctx context.Context, request *templatev1.SaveTemplateProvidersRequest) (*templatev1.SaveTemplateProvidersResponse, error) {
	if request == nil || request.TplId == 0 || request.TplVersionId == 0 {
		return &templatev1.SaveTemplateProvidersResponse{}, status.Errorf(codes.InvalidArgument, "request or tpl_id or tpl_version_id is nil")
	}

	if len(request.RelatedProviders) == 0 {
		return &templatev1.SaveTemplateProvidersResponse{}, status.Errorf(codes.InvalidArgument, "providers is empty")
	}

	providers := slice.Map(request.RelatedProviders, func(_ int, provider *templatev1.RelatedProvider) domain.ChannelTemplateProvider {
		return domain.ChannelTemplateProvider{
			TplId:        request.TplId,
			TplVersionId: request.TplVersionId,
			ProviderId:   provider.ProviderId,
		}
	})

	err := s.svc.SaveProviders(ctx, providers)
	if err != nil {
		return &templatev1.SaveTemplateProvidersResponse{}, status.Errorf(codes.Internal, "failed to save template providers: %v", err)
	}

	return &templatev1.SaveTemplateProvidersResponse{}, nil
}

func (s *TemplateServer) DeleteTemplateProvider(ctx context.Context, request *templatev1.DeleteTemplateProviderRequest) (*templatev1.DeleteTemplateProviderResponse, error) {
	if request == nil || request.Id == 0 {
		return &templatev1.DeleteTemplateProviderResponse{}, status.Errorf(codes.InvalidArgument, "request or id is nil")
	}

	err := s.svc.DeleteProvider(ctx, request.Id)
	if err != nil {
		return &templatev1.DeleteTemplateProviderResponse{}, status.Errorf(codes.Internal, "failed to delete template provider: %v", err)
	}

	return &templatev1.DeleteTemplateProviderResponse{}, nil
}

func (s *TemplateServer) ListTemplateProvider(ctx context.Context, request *templatev1.ListTemplateProviderRequest) (*templatev1.ListTemplateProviderResponse, error) {
	if request == nil || request.VersionId == 0 {
		return &templatev1.ListTemplateProviderResponse{}, status.Errorf(codes.InvalidArgument, "request or version_id is nil")
	}

	providers, err := s.svc.FindProviderByVersionId(ctx, request.VersionId)
	if err != nil {
		return &templatev1.ListTemplateProviderResponse{}, status.Errorf(codes.Internal, "failed to list template provider: %v", err)
	}

	pbs := slice.Map(providers, func(_ int, provider domain.ChannelTemplateProvider) *templatev1.TemplateProvider {
		return s.applyMaskToProviderPb(provider, request.FieldMask)
	})

	return &templatev1.ListTemplateProviderResponse{
		Providers: pbs,
	}, nil
}

func (s *TemplateServer) applyMaskToProviderPb(provider domain.ChannelTemplateProvider, mask *fieldmaskpb.FieldMask) *templatev1.TemplateProvider {
	if mask == nil || len(mask.Paths) == 0 {
		return s.domainToProviderPb(provider)
	}

	res := &templatev1.TemplateProvider{}
	for _, field := range mask.Paths {
		switch field {
		case templatev1.ProviderFieldId:
			res.Id = provider.Id
		case templatev1.ProviderFieldTplId:
			res.TplId = provider.TplId
		case templatev1.ProviderFieldTplVersionId:
			res.TplVersionId = provider.TplVersionId
		case templatev1.ProviderFieldProviderId:
			res.ProviderId = provider.ProviderId
		case templatev1.ProviderFieldProviderName:
			res.ProviderName = provider.ProviderName
		case templatev1.ProviderFieldProviderTplId:
			res.ProviderTplId = provider.ProviderTplId
		case templatev1.ProviderFieldProviderChannel:
			res.ProviderChannel = commonv1.Channel(provider.ProviderChannel)
		case templatev1.ProviderFieldAuditRequestId:
			res.AuditRequestId = provider.AuditRequestId
		case templatev1.ProviderFieldAuditStatus:
			res.AuditStatus = string(provider.AuditStatus)
		case templatev1.ProviderFieldRejectionReason:
			res.RejectionReason = provider.RejectionReason
		case templatev1.ProviderFieldLastReviewAt:
			res.LastReviewAt = provider.LastReviewAt
		case templatev1.ProviderFieldCreatedAt:
			res.CreatedAt = provider.CreatedAt
		case templatev1.ProviderFieldUpdatedAt:
			res.UpdatedAt = provider.UpdatedAt
		}
	}
	return res
}

func (s *TemplateServer) domainToProviderPb(provider domain.ChannelTemplateProvider) *templatev1.TemplateProvider {
	return &templatev1.TemplateProvider{
		Id:              provider.Id,
		TplId:           provider.TplId,
		TplVersionId:    provider.TplVersionId,
		ProviderId:      provider.ProviderId,
		ProviderName:    provider.ProviderName,
		ProviderTplId:   provider.ProviderTplId,
		ProviderChannel: commonv1.Channel(provider.ProviderChannel),
		AuditRequestId:  provider.AuditRequestId,
		AuditStatus:     string(provider.AuditStatus),
		RejectionReason: provider.RejectionReason,
		LastReviewAt:    provider.LastReviewAt,
		CreatedAt:       provider.CreatedAt,
		UpdatedAt:       provider.UpdatedAt,
	}
}

func NewTemplateServer(svc template.Service) *TemplateServer {
	return &TemplateServer{
		svc: svc,
	}
}

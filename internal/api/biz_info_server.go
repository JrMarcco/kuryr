package api

import (
	"context"
	"fmt"

	"github.com/JrMarcco/easy-kit/slice"
	businessv1 "github.com/JrMarcco/kuryr-api/api/go/business/v1"
	"github.com/JrMarcco/kuryr/internal/domain"
	"github.com/JrMarcco/kuryr/internal/errs"
	pkggorm "github.com/JrMarcco/kuryr/internal/pkg/gorm"
	"github.com/JrMarcco/kuryr/internal/search"
	"github.com/JrMarcco/kuryr/internal/service/bizinfo"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
)

var _ businessv1.BusinessServiceServer = (*BizInfoServer)(nil)

type BizInfoServer struct {
	svc bizinfo.Service
}

func (s *BizInfoServer) Save(ctx context.Context, request *businessv1.SaveRequest) (*businessv1.SaveResponse, error) {
	if request == nil || request.BusinessInfo == nil {
		return &businessv1.SaveResponse{}, status.Errorf(codes.InvalidArgument, "request is nil or business info is nil")
	}

	bizInfo := domain.BizInfo{
		BizKey:       request.BusinessInfo.BizKey,
		BizName:      request.BusinessInfo.BizName,
		BizType:      domain.BizType(request.BusinessInfo.BizType),
		Contact:      request.BusinessInfo.Contact,
		ContactEmail: request.BusinessInfo.ContactEmail,
		CreatorId:    request.BusinessInfo.CreatorId,
	}

	if err := bizInfo.Validate(); err != nil {
		return &businessv1.SaveResponse{}, status.Errorf(codes.InvalidArgument, "invalid biz info: %v", err)
	}

	saved, err := s.svc.Save(ctx, bizInfo)
	if err != nil {
		return &businessv1.SaveResponse{}, status.Errorf(codes.Internal, "failed to save biz info: %v", err)
	}

	return &businessv1.SaveResponse{
		BusinessInfo: s.domainToPb(saved),
	}, nil
}

func (s *BizInfoServer) Delete(ctx context.Context, request *businessv1.DeleteRequest) (*businessv1.DeleteResponse, error) {
	if request == nil || request.BizId == 0 {
		return &businessv1.DeleteResponse{}, status.Errorf(codes.InvalidArgument, "request is nil or biz id is invalid")
	}

	err := s.svc.Delete(ctx, request.BizId)
	if err != nil {
		return &businessv1.DeleteResponse{}, status.Errorf(codes.Internal, "failed to delete biz info: %v", err)
	}

	return &businessv1.DeleteResponse{}, nil
}

func (s BizInfoServer) Update(ctx context.Context, request *businessv1.UpdateRequest) (*businessv1.UpdateResponse, error) {
	if request == nil || request.BusinessInfo == nil {
		return &businessv1.UpdateResponse{}, status.Errorf(codes.InvalidArgument, "request is nil or business info is nil")
	}

	bizInfo, err := s.applyMaskToDomain(request.BusinessInfo, request.FieldMask)
	if err != nil {
		return &businessv1.UpdateResponse{}, status.Errorf(codes.InvalidArgument, "invalid biz info: %v", err)
	}

	updated, err := s.svc.Update(ctx, bizInfo)
	if err != nil {
		return &businessv1.UpdateResponse{}, status.Errorf(codes.Internal, "failed to update biz info: %v", err)
	}

	return &businessv1.UpdateResponse{
		BusinessInfo: s.domainToPb(updated),
	}, nil
}

func (s *BizInfoServer) applyMaskToDomain(pb *businessv1.BusinessInfo, mask *fieldmaskpb.FieldMask) (domain.BizInfo, error) {
	if mask == nil || len(mask.Paths) == 0 {
		return domain.BizInfo{}, status.Errorf(codes.InvalidArgument, "field mask is nil or paths is empty")
	}

	bizInfo := domain.BizInfo{
		Id: pb.Id,
	}

	for _, field := range mask.Paths {
		if _, ok := businessv1.UpdatableFields[field]; !ok {
			return domain.BizInfo{}, fmt.Errorf("%w: field [ %s ] is not updatable", errs.ErrInvalidParam, field)
		}

		switch field {
		case businessv1.FieldBizName:
			bizInfo.BizName = pb.BizName
		case businessv1.FieldContact:
			bizInfo.Contact = pb.Contact
		case businessv1.FieldContactEmail:
			bizInfo.ContactEmail = pb.ContactEmail
		}
	}

	return bizInfo, nil
}

func (s *BizInfoServer) Search(ctx context.Context, request *businessv1.SearchRequest) (*businessv1.SearchResponse, error) {
	if request == nil {
		return &businessv1.SearchResponse{}, status.Errorf(codes.InvalidArgument, "request is nil")
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

	criteria := search.BizSearchCriteria{
		BizName: request.BizName,
	}

	res, err := s.svc.Search(ctx, criteria, param)
	if err != nil {
		return &businessv1.SearchResponse{}, status.Errorf(codes.Internal, "failed to search biz info: %v", err)
	}

	pbs := slice.Map(res.Records, func(idx int, src domain.BizInfo) *businessv1.BusinessInfo {
		return s.applyMaskToPb(src, request.FieldMask)
	})

	return &businessv1.SearchResponse{
		Records: pbs,
		Total:   res.Total,
	}, nil
}

func (s *BizInfoServer) applyMaskToPb(bizInfo domain.BizInfo, mask *fieldmaskpb.FieldMask) *businessv1.BusinessInfo {
	if mask == nil || len(mask.Paths) == 0 {
		return s.domainToPb(bizInfo)
	}

	pb := &businessv1.BusinessInfo{}

	for _, field := range mask.Paths {
		switch field {
		case businessv1.FieldId:
			pb.Id = bizInfo.Id
		case businessv1.FieldBizKey:
			pb.BizKey = bizInfo.BizKey
		case businessv1.FieldBizName:
			pb.BizName = bizInfo.BizName
		case businessv1.FieldBizType:
			pb.BizType = string(bizInfo.BizType)
		case businessv1.FieldBizSecret:
			pb.BizSecret = bizInfo.BizSecret
		case businessv1.FieldContact:
			pb.Contact = bizInfo.Contact
		case businessv1.FieldContactEmail:
			pb.ContactEmail = bizInfo.ContactEmail
		case businessv1.FieldCreatorId:
			pb.CreatorId = bizInfo.CreatorId
		case businessv1.FieldCreatedAt:
			pb.CreatedAt = bizInfo.CreatedAt
		case businessv1.FieldUpdatedAt:
			pb.UpdatedAt = bizInfo.UpdatedAt
		}
	}

	return pb
}

func (s *BizInfoServer) domainToPb(bizInfo domain.BizInfo) *businessv1.BusinessInfo {
	return &businessv1.BusinessInfo{
		Id:           bizInfo.Id,
		BizKey:       bizInfo.BizKey,
		BizName:      bizInfo.BizName,
		BizType:      string(bizInfo.BizType),
		BizSecret:    bizInfo.BizSecret,
		Contact:      bizInfo.Contact,
		ContactEmail: bizInfo.ContactEmail,
		CreatorId:    bizInfo.CreatorId,
		CreatedAt:    bizInfo.CreatedAt,
		UpdatedAt:    bizInfo.UpdatedAt,
	}
}

func NewBizInfoServer(svc bizinfo.Service) *BizInfoServer {
	return &BizInfoServer{
		svc: svc,
	}
}

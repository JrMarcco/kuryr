package repository

import (
	"context"
	"errors"

	"github.com/JrMarcco/kuryr/internal/domain"
	"github.com/JrMarcco/kuryr/internal/errs"
	"github.com/JrMarcco/kuryr/internal/repository/dao"
	"gorm.io/gorm"
)

type ChannelTplRepo interface {
	FindById(ctx context.Context, id uint64) (domain.ChannelTemplate, error)
}

var _ ChannelTplRepo = (*DefaultChannelTplRepo)(nil)

type DefaultChannelTplRepo struct {
	dao dao.ChannelTplDao
}

func (r *DefaultChannelTplRepo) FindById(ctx context.Context, id uint64) (domain.ChannelTemplate, error) {
	entity, err := r.dao.FindById(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return domain.ChannelTemplate{}, errs.ErrRecordNotFound
		}
		return domain.ChannelTemplate{}, err
	}

	templates, err := r.getTemplates(ctx, []dao.ChannelTemplate{entity})

	const first = 0
	return templates[first], nil
}

func (r *DefaultChannelTplRepo) getTemplates(ctx context.Context, entities []dao.ChannelTemplate) ([]domain.ChannelTemplate, error) {
	ids := make([]uint64, 0, len(entities))
	for i := range entities {
		ids[i] = entities[i].Id
	}

	// 获取关联版本
	versionEntities, err := r.dao.FindVersionsByIds(ctx, ids)
	if err != nil {
		return nil, err
	}

	versionIds := make([]uint64, 0, len(versionEntities))
	for i := range versionEntities {
		versionIds[i] = versionEntities[i].Id
	}

	// 获取关联供应商
	providerEntities, err := r.dao.FindProviderByVersionIds(ctx, versionIds)
	if err != nil {
		return nil, err
	}

	// 版本关联供应商
	versionToProvider := make(map[uint64][]domain.ChannelTemplateProvider)
	for _, providerEntity := range providerEntities {
		provider := r.toProviderDomain(providerEntity)
		versionToProvider[provider.TplVersionId] = append(versionToProvider[provider.TplVersionId], provider)
	}

	// 模板关联版本
	templateToVersion := make(map[uint64][]domain.ChannelTemplateVersion)
	for _, versionEntity := range versionEntities {
		version := r.toVersionDomain(versionEntity)
		version.Providers = versionToProvider[version.Id]
		templateToVersion[version.TplId] = append(templateToVersion[version.TplId], version)
	}

	res := make([]domain.ChannelTemplate, len(entities))
	for i, tplEntity := range entities {
		template := r.toTemplateDomain(tplEntity)
		template.Versions = templateToVersion[template.Id]
		res[i] = template
	}

	return res, nil
}

func (r *DefaultChannelTplRepo) toTemplateDomain(entity dao.ChannelTemplate) domain.ChannelTemplate {
	return domain.ChannelTemplate{
		Id:               entity.Id,
		OwnerId:          entity.OwnerId,
		OwnerType:        domain.OwnerType(entity.OwnerType),
		TplName:          entity.TplName,
		TplDesc:          entity.TplDesc,
		Channel:          domain.Channel(entity.Channel),
		NotificationType: domain.NotificationType(entity.NotificationType),
		ActivatedVersion: entity.ActivatedVersion,
		CreatedAt:        entity.CreatedAt,
		UpdatedAt:        entity.UpdatedAt,
	}
}

func (r *DefaultChannelTplRepo) toVersionDomain(entity dao.ChannelTemplateVersion) domain.ChannelTemplateVersion {
	return domain.ChannelTemplateVersion{
		Id:              entity.Id,
		TplId:           entity.TplId,
		VersionName:     entity.VersionName,
		Signature:       entity.Signature,
		Content:         entity.Content,
		ApplyRemark:     entity.ApplyRemark,
		AuditId:         entity.AuditId,
		AuditorId:       entity.AuditorId,
		AuditTime:       entity.AuditTime,
		AuditStatus:     domain.AuditStatus(entity.AuditStatus),
		RejectionReason: entity.RejectionReason,
		LastReviewAt:    entity.LastReviewAt,
		CreatedAt:       entity.CreatedAt,
		UpdatedAt:       entity.UpdatedAt,
	}
}

func (r *DefaultChannelTplRepo) toProviderDomain(entity dao.ChannelTemplateProvider) domain.ChannelTemplateProvider {
	return domain.ChannelTemplateProvider{
		Id:              entity.Id,
		TplId:           entity.TplId,
		TplVersionId:    entity.TplVersionId,
		ProviderId:      entity.ProviderId,
		ProviderName:    entity.ProviderName,
		ProviderChannel: domain.Channel(entity.ProviderChannel),
		ProviderTplId:   entity.ProviderTplId,
		AuditRequestId:  entity.AuditRequestId,
		AuditStatus:     domain.AuditStatus(entity.AuditStatus),
		RejectionReason: entity.RejectionReason,
		LastReviewAt:    entity.LastReviewAt,
		CreatedAt:       entity.CreatedAt,
		UpdatedAt:       entity.UpdatedAt,
	}
}

func NewDefaultChannelTplRepo(dao dao.ChannelTplDao) *DefaultChannelTplRepo {
	return &DefaultChannelTplRepo{
		dao: dao,
	}
}

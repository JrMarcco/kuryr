package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/JrMarcco/easy-kit/slice"
	"github.com/JrMarcco/kuryr/internal/domain"
	"github.com/JrMarcco/kuryr/internal/errs"
	"github.com/JrMarcco/kuryr/internal/repository/dao"
	"gorm.io/gorm"
)

type ChannelTplRepo interface {
	SaveTemplate(ctx context.Context, template domain.ChannelTemplate) error
	SaveVersion(ctx context.Context, version domain.ChannelTemplateVersion) error
	SaveProviders(ctx context.Context, providers []domain.ChannelTemplateProvider) error

	ActivateVersion(ctx context.Context, templateId uint64, versionId uint64) error

	// FindDetailById 查询详情，包含版本、供应商信息。
	FindById(ctx context.Context, id uint64) (domain.ChannelTemplate, error)
	FindDetailById(ctx context.Context, id uint64) (domain.ChannelTemplate, error)

	FindVersionById(ctx context.Context, id uint64) (domain.ChannelTemplateVersion, error)

	ListByOwner(ctx context.Context, ownerId uint64) ([]domain.ChannelTemplate, error)
}

var _ ChannelTplRepo = (*DefaultChannelTplRepo)(nil)

type DefaultChannelTplRepo struct {
	dao dao.ChannelTplDao
}

func (r *DefaultChannelTplRepo) SaveTemplate(ctx context.Context, template domain.ChannelTemplate) error {
	return r.dao.SaveTemplate(ctx, r.toTemplateEntity(template))
}

func (r *DefaultChannelTplRepo) SaveVersion(ctx context.Context, version domain.ChannelTemplateVersion) error {
	return r.dao.SaveVersion(ctx, r.toVersionEntity(version))
}

func (r *DefaultChannelTplRepo) SaveProviders(ctx context.Context, providers []domain.ChannelTemplateProvider) error {
	entities := slice.Map(providers, func(_ int, provider domain.ChannelTemplateProvider) dao.ChannelTemplateProvider {
		return r.toProviderEntity(provider)
	})
	return r.dao.SaveProviders(ctx, entities)
}

func (r *DefaultChannelTplRepo) ActivateVersion(ctx context.Context, templateId uint64, versionId uint64) error {
	return r.dao.ActivateVersion(ctx, templateId, versionId)
}

func (r *DefaultChannelTplRepo) FindById(ctx context.Context, id uint64) (domain.ChannelTemplate, error) {
	entity, err := r.dao.FindById(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return domain.ChannelTemplate{}, fmt.Errorf("%w: cannot find channel template, id = %d", errs.ErrRecordNotFound, id)
		}
		return domain.ChannelTemplate{}, err
	}
	return r.toTemplateDomain(entity), nil
}

func (r *DefaultChannelTplRepo) FindDetailById(ctx context.Context, id uint64) (domain.ChannelTemplate, error) {
	entity, err := r.dao.FindById(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return domain.ChannelTemplate{}, fmt.Errorf("%w: cannot find channel template, id = %d", errs.ErrRecordNotFound, id)
		}
		return domain.ChannelTemplate{}, err
	}

	templates, err := r.getTemplates(ctx, []dao.ChannelTemplate{entity})
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return domain.ChannelTemplate{}, fmt.Errorf("%w: cannot find template version or privider, template id = %d", errs.ErrRecordNotFound, id)
		}
		return domain.ChannelTemplate{}, err
	}

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

// FindVersionById implements ChannelTplRepo.
func (r *DefaultChannelTplRepo) FindVersionById(ctx context.Context, id uint64) (domain.ChannelTemplateVersion, error) {
	entity, err := r.dao.FindVersionById(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return domain.ChannelTemplateVersion{}, fmt.Errorf("%w: cannot find channel template version, id = %d", errs.ErrRecordNotFound, id)
		}
		return domain.ChannelTemplateVersion{}, err
	}
	return r.toVersionDomain(entity), nil
}

func (r *DefaultChannelTplRepo) ListByOwner(ctx context.Context, ownerId uint64) ([]domain.ChannelTemplate, error) {
	entities, err := r.dao.ListByOwner(ctx, ownerId)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("%w: cannot find channel template, ownerId = %d", errs.ErrRecordNotFound, ownerId)
		}
		return nil, err
	}

	templates := slice.Map(entities, func(_ int, entity dao.ChannelTemplate) domain.ChannelTemplate {
		return r.toTemplateDomain(entity)
	})
	return templates, nil
}

func (r *DefaultChannelTplRepo) toTemplateDomain(entity dao.ChannelTemplate) domain.ChannelTemplate {
	return domain.ChannelTemplate{
		Id:                 entity.Id,
		OwnerId:            entity.OwnerId,
		OwnerType:          domain.OwnerType(entity.OwnerType),
		TplName:            entity.TplName,
		TplDesc:            entity.TplDesc,
		Channel:            domain.Channel(entity.Channel),
		NotificationType:   domain.NotificationType(entity.NotificationType),
		ActivatedVersionId: entity.ActivatedVersionId,
		CreatedAt:          entity.CreatedAt,
		UpdatedAt:          entity.UpdatedAt,
	}
}

func (r *DefaultChannelTplRepo) toTemplateEntity(template domain.ChannelTemplate) dao.ChannelTemplate {
	return dao.ChannelTemplate{
		Id:                 template.Id,
		OwnerId:            template.OwnerId,
		OwnerType:          string(template.OwnerType),
		TplName:            template.TplName,
		TplDesc:            template.TplDesc,
		Channel:            int32(template.Channel),
		NotificationType:   int32(template.NotificationType),
		ActivatedVersionId: template.ActivatedVersionId,
		CreatedAt:          template.CreatedAt,
		UpdatedAt:          template.UpdatedAt,
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

func (r *DefaultChannelTplRepo) toVersionEntity(version domain.ChannelTemplateVersion) dao.ChannelTemplateVersion {
	return dao.ChannelTemplateVersion{
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

func (r *DefaultChannelTplRepo) toProviderEntity(provider domain.ChannelTemplateProvider) dao.ChannelTemplateProvider {
	return dao.ChannelTemplateProvider{
		Id:              provider.Id,
		TplId:           provider.TplId,
		TplVersionId:    provider.TplVersionId,
		ProviderId:      provider.ProviderId,
		ProviderName:    provider.ProviderName,
		ProviderChannel: int32(provider.ProviderChannel),
		ProviderTplId:   provider.ProviderTplId,
		AuditRequestId:  provider.AuditRequestId,
		AuditStatus:     string(provider.AuditStatus),
		RejectionReason: provider.RejectionReason,
		LastReviewAt:    provider.LastReviewAt,
		CreatedAt:       provider.CreatedAt,
		UpdatedAt:       provider.UpdatedAt,
	}
}

func NewDefaultChannelTplRepo(dao dao.ChannelTplDao) *DefaultChannelTplRepo {
	return &DefaultChannelTplRepo{
		dao: dao,
	}
}

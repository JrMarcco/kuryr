package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/JrMarcco/easy-kit/slice"
	"github.com/JrMarcco/kuryr/internal/domain"
	"github.com/JrMarcco/kuryr/internal/errs"
	pkggorm "github.com/JrMarcco/kuryr/internal/pkg/gorm"
	"github.com/JrMarcco/kuryr/internal/repository/dao"
	"gorm.io/gorm"
)

type ChannelTplRepo interface {
	SaveTemplate(ctx context.Context, template domain.ChannelTemplate) (domain.ChannelTemplate, error)
	DeleteTemplate(ctx context.Context, id uint64) error
	// GetDetailById 查询详情，包含版本、供应商信息。
	GetDetailById(ctx context.Context, id uint64) (domain.ChannelTemplate, error)
	FindTemplateById(ctx context.Context, id uint64) (domain.ChannelTemplate, error)
	FindTemplateByBizId(ctx context.Context, bizId uint64, param *pkggorm.PaginationParam) (*pkggorm.PaginationResult[domain.ChannelTemplate], error)

	SaveVersion(ctx context.Context, version domain.ChannelTemplateVersion) (domain.ChannelTemplateVersion, error)
	DeleteVersion(ctx context.Context, id uint64) error
	ActivateVersion(ctx context.Context, templateId uint64, versionId uint64) error
	FindVersionById(ctx context.Context, id uint64) (domain.ChannelTemplateVersion, error)
	FindVersionByTplId(ctx context.Context, tplId uint64) ([]domain.ChannelTemplateVersion, error)

	SaveProviders(ctx context.Context, providers []domain.ChannelTemplateProvider) error
	DeleteProvider(ctx context.Context, id uint64) error
	FindProviderByVersionId(ctx context.Context, versionId uint64) ([]domain.ChannelTemplateProvider, error)
}

var _ ChannelTplRepo = (*DefaultChannelTplRepo)(nil)

type DefaultChannelTplRepo struct {
	dao dao.ChannelTplDao
}

func (r *DefaultChannelTplRepo) SaveTemplate(ctx context.Context, template domain.ChannelTemplate) (domain.ChannelTemplate, error) {
	entity, err := r.dao.SaveTemplate(ctx, r.toTemplateEntity(template))
	if err != nil {
		return domain.ChannelTemplate{}, err
	}
	return r.toTemplateDomain(entity), nil
}

func (r *DefaultChannelTplRepo) DeleteTemplate(ctx context.Context, id uint64) error {
	return r.dao.DeleteTemplate(ctx, id)
}

func (r *DefaultChannelTplRepo) GetDetailById(ctx context.Context, id uint64) (domain.ChannelTemplate, error) {
	entity, err := r.dao.FindTemplateById(ctx, id)
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
	versionEntities, err := r.dao.FindVersionByIds(ctx, ids)
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

func (r *DefaultChannelTplRepo) FindTemplateById(ctx context.Context, id uint64) (domain.ChannelTemplate, error) {
	entity, err := r.dao.FindTemplateById(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return domain.ChannelTemplate{}, fmt.Errorf("%w: cannot find channel template, id = %d", errs.ErrRecordNotFound, id)
		}
		return domain.ChannelTemplate{}, err
	}
	return r.toTemplateDomain(entity), nil
}

func (r *DefaultChannelTplRepo) FindTemplateByBizId(ctx context.Context, bizId uint64, param *pkggorm.PaginationParam) (*pkggorm.PaginationResult[domain.ChannelTemplate], error) {
	res, err := r.dao.FindTemplateByBizId(ctx, bizId, param)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("%w: cannot find channel template, biz_id = %d", errs.ErrRecordNotFound, bizId)
		}
		return nil, err
	}

	if res.Total == 0 {
		return pkggorm.NewPaginationResult([]domain.ChannelTemplate{}, 0), nil
	}

	templates := slice.Map(res.Records, func(_ int, entity dao.ChannelTemplate) domain.ChannelTemplate {
		return r.toTemplateDomain(entity)
	})

	return pkggorm.NewPaginationResult(templates, res.Total), nil
}

func (r *DefaultChannelTplRepo) SaveVersion(ctx context.Context, version domain.ChannelTemplateVersion) (domain.ChannelTemplateVersion, error) {
	entity, err := r.dao.SaveVersion(ctx, r.toVersionEntity(version))
	if err != nil {
		return domain.ChannelTemplateVersion{}, err
	}
	return r.toVersionDomain(entity), nil
}

func (r *DefaultChannelTplRepo) DeleteVersion(ctx context.Context, id uint64) error {
	return r.dao.DeleteVersion(ctx, id)
}

func (r *DefaultChannelTplRepo) ActivateVersion(ctx context.Context, templateId uint64, versionId uint64) error {
	return r.dao.ActivateVersion(ctx, templateId, versionId)
}

func (r *DefaultChannelTplRepo) FindVersionById(ctx context.Context, id uint64) (domain.ChannelTemplateVersion, error) {

	entity, err := r.dao.FindVersionById(ctx, id)
	if err != nil {
		return domain.ChannelTemplateVersion{}, err
	}
	return r.toVersionDomain(entity), nil
}

func (r *DefaultChannelTplRepo) FindVersionByTplId(ctx context.Context, tplId uint64) ([]domain.ChannelTemplateVersion, error) {
	entities, err := r.dao.FindVersionByTplId(ctx, tplId)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("%w: cannot find channel template version, tpl_id = %d", errs.ErrRecordNotFound, tplId)
		}
		return nil, err
	}

	return slice.Map(entities, func(_ int, entity dao.ChannelTemplateVersion) domain.ChannelTemplateVersion {
		return r.toVersionDomain(entity)
	}), nil
}

func (r *DefaultChannelTplRepo) SaveProviders(ctx context.Context, providers []domain.ChannelTemplateProvider) error {
	entities := slice.Map(providers, func(_ int, provider domain.ChannelTemplateProvider) dao.ChannelTemplateProvider {
		return r.toProviderEntity(provider)
	})

	for i := range entities {
		// 审批状态设置为 “待审核”，
		// 等待调度任务提交到供应商审核。
		entities[i].AuditStatus = string(domain.AuditStatusPending)
		entities[i].LastReviewAt = 0
	}

	return r.dao.SaveProviders(ctx, entities)
}

func (r *DefaultChannelTplRepo) DeleteProvider(ctx context.Context, id uint64) error {
	return r.dao.DeleteProvider(ctx, id)
}

func (r *DefaultChannelTplRepo) FindProviderByVersionId(ctx context.Context, versionId uint64) ([]domain.ChannelTemplateProvider, error) {
	entities, err := r.dao.FindProviderByVersionId(ctx, versionId)
	if err != nil {
		return nil, err
	}
	return slice.Map(entities, func(_ int, entity dao.ChannelTemplateProvider) domain.ChannelTemplateProvider {
		return r.toProviderDomain(entity)
	}), nil
}

func (r *DefaultChannelTplRepo) toTemplateDomain(entity dao.ChannelTemplate) domain.ChannelTemplate {
	return domain.ChannelTemplate{
		Id:                 entity.Id,
		BizId:              entity.BizId,
		BizType:            domain.BizType(entity.BizType),
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
		BizId:              template.BizId,
		BizType:            string(template.BizType),
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

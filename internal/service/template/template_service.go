package template

import (
	"context"
	"fmt"

	"github.com/JrMarcco/kuryr/internal/domain"
	"github.com/JrMarcco/kuryr/internal/errs"
	pkggorm "github.com/JrMarcco/kuryr/internal/pkg/gorm"
	"github.com/JrMarcco/kuryr/internal/repository"
)

//go:generate mockgen -source=./template_service.go -destination=./mock/template_service.mock.go -package=templatemock -typed Service

type Service interface {
	SaveTemplate(ctx context.Context, template domain.ChannelTemplate) (domain.ChannelTemplate, error)
	DeleteTemplate(ctx context.Context, id uint64) error
	FindTemplateByBizId(ctx context.Context, bizId uint64, param *pkggorm.PaginationParam) (*pkggorm.PaginationResult[domain.ChannelTemplate], error)

	SaveVersion(ctx context.Context, version domain.ChannelTemplateVersion) (domain.ChannelTemplateVersion, error)
	DeleteVersion(ctx context.Context, id uint64) error
	ActivateVersion(ctx context.Context, templateId uint64, versionId uint64) error
	FindVersionByTplId(ctx context.Context, tplId uint64) ([]domain.ChannelTemplateVersion, error)

	SaveProviders(ctx context.Context, providers []domain.ChannelTemplateProvider) error
	DeleteProvider(ctx context.Context, id uint64) error
	FindProviderByVersionId(ctx context.Context, versionId uint64) ([]domain.ChannelTemplateProvider, error)
}

var _ Service = (*DefaultService)(nil)

type DefaultService struct {
	repo repository.ChannelTplRepo
}

// SaveTemplate 新增渠道模板。
// 注意：新创建的模板 activated_version_id 为 0，表示暂无可用版本，
// 需要用户后续创建并审核通过版本后才能激活使用。
func (s *DefaultService) SaveTemplate(ctx context.Context, template domain.ChannelTemplate) (domain.ChannelTemplate, error) {
	if err := template.Validate(); err != nil {
		return domain.ChannelTemplate{}, err
	}
	// 新模板的激活版本 id 应该为 0，表示暂无可用版本。
	template.ActivatedVersionId = 0

	return s.repo.SaveTemplate(ctx, template)
}

func (s *DefaultService) DeleteTemplate(ctx context.Context, id uint64) error {
	return s.repo.DeleteTemplate(ctx, id)
}

func (s *DefaultService) FindTemplateByBizId(ctx context.Context, bizId uint64, param *pkggorm.PaginationParam) (*pkggorm.PaginationResult[domain.ChannelTemplate], error) {
	return s.repo.FindTemplateByBizId(ctx, bizId, param)
}

func (s *DefaultService) SaveVersion(ctx context.Context, version domain.ChannelTemplateVersion) (domain.ChannelTemplateVersion, error) {
	if err := version.Validate(); err != nil {
		return domain.ChannelTemplateVersion{}, err
	}
	return s.repo.SaveVersion(ctx, version)
}

func (s *DefaultService) DeleteVersion(ctx context.Context, id uint64) error {
	return s.repo.DeleteVersion(ctx, id)
}

// TODO: 这里需要做一些业务上的判断。
// ActivateVersion 激活版本。
func (s *DefaultService) ActivateVersion(ctx context.Context, templateId uint64, versionId uint64) error {
	template, err := s.repo.FindTemplateById(ctx, templateId)
	if err != nil {
		return err
	}

	// 查询版本
	version, err := s.repo.FindVersionById(ctx, versionId)
	if err != nil {
		return err
	}

	if template.Id != version.TplId {
		return fmt.Errorf("%w: template id and version id mismatch", errs.ErrInvalidParam)
	}

	// 只允许激活审批通过的版本。
	if version.AuditStatus != domain.AuditStatusApproved {
		return fmt.Errorf("%w: version is not approved", errs.ErrInvalidStatus)
	}

	return s.repo.ActivateVersion(ctx, templateId, versionId)
}

func (s *DefaultService) FindVersionByTplId(ctx context.Context, tplId uint64) ([]domain.ChannelTemplateVersion, error) {
	return s.repo.FindVersionByTplId(ctx, tplId)
}

// SaveProviders 新增版本关联供应商。
// TODO: 需要增加调度任务自动提交到供应商侧审批。
func (s *DefaultService) SaveProviders(ctx context.Context, providers []domain.ChannelTemplateProvider) error {
	for i := range providers {
		if err := providers[i].Validate(); err != nil {
			return err
		}
	}

	err := s.repo.SaveProviders(ctx, providers)
	if err != nil {
		return err
	}

	// TODO: 提交消息到 kafka，消费者端来提交模板到供应商侧进行审核
	return nil
}

// DeleteProvider 删除版本关联供应商。
// TODO: 这里要考虑增加是否允许删除的逻辑，同时要考虑删除后是否需要通知供应商侧。
func (s *DefaultService) DeleteProvider(ctx context.Context, id uint64) error {
	return s.repo.DeleteProvider(ctx, id)
}

func (s *DefaultService) FindProviderByVersionId(ctx context.Context, versionId uint64) ([]domain.ChannelTemplateProvider, error) {
	return s.repo.FindProviderByVersionId(ctx, versionId)
}

func NewDefaultService(repo repository.ChannelTplRepo) *DefaultService {
	return &DefaultService{
		repo: repo,
	}
}

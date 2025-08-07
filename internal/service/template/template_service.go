package template

import (
	"context"
	"fmt"

	"github.com/JrMarcco/kuryr/internal/domain"
	"github.com/JrMarcco/kuryr/internal/errs"
	"github.com/JrMarcco/kuryr/internal/repository"
)

//go:generate mockgen -source=./template_service.go -destination=./mock/template_service.mock.go -package=templatemock -typed Service

type Service interface {
	SaveTemplate(ctx context.Context, template domain.ChannelTemplate) error
	SaveVersion(ctx context.Context, version domain.ChannelTemplateVersion) error
	SaveProviders(ctx context.Context, providers []domain.ChannelTemplateProvider) error

	ActivateVersion(ctx context.Context, templateId uint64, versionId uint64) error
}

var _ Service = (*DefaultService)(nil)

type DefaultService struct {
	repo repository.ChannelTplRepo
}

// SaveTemplate 新增渠道模板。
// 注意：新创建的模板 activated_version_id 为 0，表示暂无可用版本，
// 需要用户后续创建并审核通过版本后才能激活使用。
func (s *DefaultService) SaveTemplate(ctx context.Context, template domain.ChannelTemplate) error {
	if err := template.Validate(); err != nil {
		return err
	}
	// 新模板的激活版本 id 应该为 0，表示暂无可用版本。
	template.ActivatedVersionId = 0
	return s.repo.SaveTemplate(ctx, template)
}

func (s *DefaultService) SaveVersion(ctx context.Context, version domain.ChannelTemplateVersion) error {
	if err := version.Validate(); err != nil {
		return err
	}
	return s.repo.SaveVersion(ctx, version)
}

func (s *DefaultService) SaveProviders(ctx context.Context, providers []domain.ChannelTemplateProvider) error {
	//TODO: unimplemented
	panic("unimplemented")
}

func (s *DefaultService) ActivateVersion(ctx context.Context, templateId uint64, versionId uint64) error {
	template, err := s.repo.FindById(ctx, templateId)
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

func NewDefaultService(repo repository.ChannelTplRepo) *DefaultService {
	return &DefaultService{
		repo: repo,
	}
}

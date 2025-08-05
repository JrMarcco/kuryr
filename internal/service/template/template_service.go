package template

import (
	"context"

	"github.com/JrMarcco/kuryr/internal/domain"
	"github.com/JrMarcco/kuryr/internal/repository"
)

//go:generate mockgen -source=./template_service.go -destination=./mock/template_service.mock.go -package=templatemock -typed Service

type Service interface {
	SaveTemplate(ctx context.Context, template domain.ChannelTemplate) error
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
	// 新模板的激活版本 ID 应该为 0，表示暂无可用版本
	template.ActivatedVersionId = 0
	return s.repo.SaveTemplate(ctx, template)
}

func NewDefaultService(repo repository.ChannelTplRepo) *DefaultService {
	return &DefaultService{
		repo: repo,
	}
}

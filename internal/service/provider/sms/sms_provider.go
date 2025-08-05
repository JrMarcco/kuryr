package sms

import (
	"context"
	"fmt"
	"strings"

	"github.com/JrMarcco/kuryr/internal/domain"
	"github.com/JrMarcco/kuryr/internal/errs"
	"github.com/JrMarcco/kuryr/internal/repository"
	"github.com/JrMarcco/kuryr/internal/service/provider"
	"github.com/JrMarcco/kuryr/internal/service/provider/sms/client"
)

var _ provider.Provider = (*Provider)(nil)

type Provider struct {
	name   string
	client client.SmsClient

	providerRepo   repository.ProviderRepo
	channelTplRepo repository.ChannelTplRepo
}

func (p *Provider) Send(ctx context.Context, n domain.Notification) (domain.SendResp, error) {
	templateId := n.Template.Id

	template, err := p.channelTplRepo.FindById(ctx, templateId)
	if err != nil {
		return domain.SendResp{}, err
	}

	if template.Id == 0 {
		return domain.SendResp{}, fmt.Errorf("%w: cannot find channel template, id = %d", errs.ErrRecordNotFound, templateId)
	}

	version, err := template.GetActivatedVersion()
	if err != nil {
		return domain.SendResp{}, err
	}

	if len(version.Providers) == 0 {
		return domain.SendResp{}, fmt.Errorf("%w: cannot find provider for channel template version, version id = %d", errs.ErrRecordNotFound, version.Id)
	}

	const first = 0
	// 获取供应商侧的模板 id
	providerTplId := version.Providers[first].ProviderTplId
	resp, err := p.client.Send(client.SendReq{
		PhoneNumbers:   n.Receivers,
		SignName:       version.Signature,
		TemplateId:     providerTplId,
		TemplateParams: n.Template.Params,
	})
	if err != nil {
		return domain.SendResp{}, fmt.Errorf("%w: %w", errs.ErrFailedToSendNotification, err)
	}

	for _, status := range resp.PhoneNumbers {
		if !strings.EqualFold(status.Code, "OK") {
			return domain.SendResp{}, fmt.Errorf("%w: code = %s, message = %s", errs.ErrFailedToSendNotification, status.Code, status.Message)
		}
	}

	return domain.SendResp{
		Result: domain.SendResult{
			NotificationId: n.Id,
			SendStatus:     domain.SendStatusSuccess,
		},
	}, nil
}

func NewProvider(
	name string,
	client client.SmsClient,
	providerRepo repository.ProviderRepo,
	channelTplRepo repository.ChannelTplRepo,
) *Provider {
	return &Provider{
		name:           name,
		client:         client,
		providerRepo:   providerRepo,
		channelTplRepo: channelTplRepo,
	}
}

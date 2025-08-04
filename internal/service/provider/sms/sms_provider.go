package sms

import (
	"context"

	"github.com/JrMarcco/kuryr/internal/domain"
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
	//TODO implement me
	panic("implement me")
}

func NewProvider(name string, client client.SmsClient, providerRepo repository.ProviderRepo, channelTplRepo repository.ChannelTplRepo) *Provider {
	return &Provider{
		name:           name,
		client:         client,
		providerRepo:   providerRepo,
		channelTplRepo: channelTplRepo,
	}
}

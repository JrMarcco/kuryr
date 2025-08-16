package sms

import (
	"github.com/JrMarcco/kuryr/internal/service/channel"
	"github.com/JrMarcco/kuryr/internal/service/ports"
	"github.com/JrMarcco/kuryr/internal/service/provider"
)

var _ ports.ChannelSender = (*SmsSender)(nil)

type SmsSender struct {
	channel.DefaultChannelSender
}

func NewSmsSender(sb provider.SelectorBuilder) *SmsSender {
	return &SmsSender{
		DefaultChannelSender: channel.DefaultChannelSender{
			SelectorBuilder: sb,
		},
	}
}

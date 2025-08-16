package ports

import (
	"context"

	"github.com/JrMarcco/kuryr/internal/domain"
)

//go:generate mockgen -source=./channel_sender.go -destination=./mock/channel_sender.mock.go -package=portsmock -typed ChannelSender

// ChannelSender 渠道发送器接口，负责将消息发送给渠道, 稳定且最小化。
type ChannelSender interface {
	Send(ctx context.Context, n domain.Notification) (domain.SendResp, error)
}

package sendstrategy

import (
	"context"

	"github.com/JrMarcco/kuryr/internal/domain"
)

var _ SendStrategy = (*DefaultSendStrategy)(nil)

type DefaultSendStrategy struct{}

func (d *DefaultSendStrategy) Send(ctx context.Context, n domain.Notification) (domain.SendResp, error) {
	// TODO: implement me
	panic("implement me")
}

func (d *DefaultSendStrategy) BatchSend(ctx context.Context, ns []domain.Notification) (domain.BatchSendResp, error) {
	// TODO: implement me
	panic("implement me")
}

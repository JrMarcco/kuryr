package callback

import "context"

type Service interface {
	SendCallback(ctx context.Context, startTime int64, batchSize int) error
}

var _ Service = (*DefaultService)(nil)

type DefaultService struct{}

func (s *DefaultService) SendCallback(ctx context.Context, startTime int64, batchSize int) error {
	// TODO: implement me
	panic("implement me")
}

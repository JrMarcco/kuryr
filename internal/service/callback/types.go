package callback

import "context"

type Service interface {
	SendCallback(ctx context.Context, startTime int64, batchSize int) error
}

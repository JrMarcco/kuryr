package mq

import (
	"context"
)

//go:generate mockgen -source=./types.go -destination=./mock/mq.mock.go -package=mqmock -typed Producer, Consumer

type Producer[T any] interface {
	Produce(ctx context.Context, event T) error
	Close()
}

type Conmsumer interface{}

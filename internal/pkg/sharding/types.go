package sharding

import "context"

// Strategy 分库分表策略
type Strategy interface {
	Shard(bizId uint64, bizKey string) Dst
	ShardWithId(id uint64) Dst
	Broadcast() []Dst
}

// Dst 目标库 & 目标表信息
type Dst struct {
	DBSuffix    uint64
	TableSuffix uint64

	DB    string
	Table string
}

type contextKeyDst struct{}

func ContextWithDst(ctx context.Context, dst Dst) context.Context {
	return context.WithValue(ctx, contextKeyDst{}, dst)
}

func ContextDst(ctx context.Context) (Dst, bool) {
	val := ctx.Value(contextKeyDst{})
	dst, ok := val.(Dst)
	return dst, ok
}

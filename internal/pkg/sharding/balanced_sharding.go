package sharding

import "math/rand"

// BroadcastMode 广播模式
type BroadcastMode string

const (
	BroadcastModeDefault    BroadcastMode = "default"     // 默认模式：保持原有的广播顺序。
	BroadcastModeRoundRobin BroadcastMode = "round_robin" // 轮询模式：交替轮询不同数据库的表。
	BroadcastModeShuffle    BroadcastMode = "shuffle"     // 随机打乱模式：随机打乱所有分片的顺序。
)

var _ Strategy = (*BalancedSharding)(nil)

// BalancedSharding 负载均衡分库分表策略。
// 注意：
//
//  1. 创建时需要传入具体的分库分表策略实现。
//  2. 这是一个装饰器，提供对广播功能的负载均衡支持。
//  3. 本身不提供分库分表的实现。
type BalancedSharding struct {
	base Strategy
	mode BroadcastMode
}

func (s *BalancedSharding) Shard(bizId uint64, bizKey string) Dst {
	return s.base.Shard(bizId, bizKey)
}

func (s *BalancedSharding) DstFromId(id uint64) Dst {
	return s.base.DstFromId(id)
}

func (s *BalancedSharding) Broadcast() []Dst {
	dsts := s.base.Broadcast()

	switch s.mode {
	case BroadcastModeRoundRobin:
		return s.roundRobinBroadcast(dsts)
	case BroadcastModeShuffle:
		return s.shuffleBroadcast(dsts)
	default:
		return dsts
	}
}

func (s *BalancedSharding) roundRobinBroadcast(dsts []Dst) []Dst {
	if len(dsts) == 0 {
		return dsts
	}

	var dbs []string
	dbGroup := make(map[string][]Dst)

	for _, dst := range dsts {
		if _, ok := dbGroup[dst.DB]; !ok {
			dbs = append(dbs, dst.DB)
		}

		dbGroup[dst.DB] = append(dbGroup[dst.DB], dst)
	}

	res := make([]Dst, 0, len(dsts))

	maxTbCnt := 0
	for _, tbs := range dbGroup {
		if len(tbs) > maxTbCnt {
			maxTbCnt = len(tbs)
		}
	}

	for i := 0; i < maxTbCnt; i++ {
		for _, db := range dbs {
			if i < len(dbGroup[db]) {
				res = append(res, dbGroup[db][i])
			}
		}
	}

	return res
}

func (s *BalancedSharding) shuffleBroadcast(dsts []Dst) []Dst {
	res := make([]Dst, len(dsts))
	copy(res, dsts)

	// Fisher-Yates 洗牌算法
	for i := len(res) - 1; i > 0; i-- {
		j := rand.Intn(i + 1)
		res[i], res[j] = res[j], res[i]
	}

	return res
}

func NewBalancedSharding(base Strategy, mode BroadcastMode) *BalancedSharding {
	return &BalancedSharding{
		base: base,
		mode: mode,
	}
}

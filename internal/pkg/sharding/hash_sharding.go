package sharding

import (
	"fmt"

	"github.com/JrMarcco/kuryr/internal/pkg/sharding/hash"
	"github.com/JrMarcco/kuryr/internal/pkg/sharding/snowflake"
)

var _ Strategy = (*HashSharding)(nil)

type HashSharding struct {
	dbPrefix    string
	tablePrefix string

	dbShardCount    uint64
	tableShardCount uint64
}

// Shard 根据 biz_id 和 biz_key 分库分表。
func (s *HashSharding) Shard(bizId uint64, bizKey string) Dst {
	hashVal := hash.Hash(bizId, bizKey)
	dbSuffix := hashVal % s.dbShardCount
	tableSuffix := (hashVal / s.dbShardCount) % s.tableShardCount

	return Dst{
		DBSuffix:    dbSuffix,
		TableSuffix: tableSuffix,
		DB:          fmt.Sprintf("%s_%d", s.dbPrefix, dbSuffix),
		Table:       fmt.Sprintf("%s_%d", s.tablePrefix, tableSuffix),
	}
}

func (s *HashSharding) ShardWithId(id uint64) Dst {
	hashVal := snowflake.ExtractHash(id)
	dbSuffix := hashVal % s.dbShardCount
	tableSuffix := (hashVal / s.dbShardCount) % s.tableShardCount

	return Dst{
		DBSuffix:    dbSuffix,
		TableSuffix: tableSuffix,
		DB:          fmt.Sprintf("%s_%d", s.dbPrefix, dbSuffix),
		Table:       fmt.Sprintf("%s_%d", s.tablePrefix, tableSuffix),
	}
}

func (s *HashSharding) Broadcast() []Dst {
	res := make([]Dst, 0, s.dbShardCount*s.tableShardCount)
	for i := uint64(0); i < s.dbShardCount; i++ {
		for j := uint64(0); j < s.tableShardCount; j++ {
			res = append(res, Dst{
				DBSuffix:    i,
				TableSuffix: j,
				DB:          fmt.Sprintf("%s_%d", s.dbPrefix, i),
				Table:       fmt.Sprintf("%s_%d", s.tablePrefix, j),
			})
		}
	}
	return res
}

func NewHashSharding(dbPrefix string, tablePrefix string, dbShardCount uint64, tableShardCount uint64) *HashSharding {
	return &HashSharding{
		dbPrefix:        dbPrefix,
		tablePrefix:     tablePrefix,
		dbShardCount:    dbShardCount,
		tableShardCount: tableShardCount,
	}
}
